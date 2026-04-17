package proxy

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// resolveOpenClawBin: openclaw 바이너리 경로 탐색 (systemd 서비스 환경 대응)
func resolveOpenClawBin() string {
	// 1. 현재 PATH에서 탐색
	if p, err := exec.LookPath("openclaw"); err == nil {
		return p
	}
	// 2. 흔한 설치 위치 직접 확인
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".npm-global/bin/openclaw"),
		filepath.Join(home, ".local/bin/openclaw"),
		"/usr/local/bin/openclaw",
		"/usr/bin/openclaw",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// notifyOpenClaw: openclaw models set + sessions.patch으로 기본값 및 TUI 세션 모두 즉시 갱신.
func notifyOpenClaw(primaryModel string) {
	bin := resolveOpenClawBin()
	if bin == "" {
		log.Printf("[openclaw-sync] openclaw 바이너리를 찾을 수 없음 — TUI 미통보")
		return
	}
	home, _ := os.UserHomeDir()
	baseEnv := append(os.Environ(),
		"HOME="+home,
		"PATH="+filepath.Dir(bin)+":/usr/local/bin:/usr/bin:/bin",
	)

	// 1. 실행 중인 TUI tmux 패인에 /model 명령 주입 (즉시 반영 — openclaw 명령보다 먼저)
	injectModelToTUI(primaryModel)

	// 2. 기본 모델 갱신 (openclaw.json agents.defaults.model.primary)
	cmd := exec.Command(bin, "models", "set", primaryModel)
	cmd.Env = baseEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[openclaw-sync] models set 실패: %v — %s", err, string(out))
	} else {
		log.Printf("[openclaw-sync] 모델 적용됨: %s", primaryModel)
	}

	// 3. sessions.patch로 세션 스토어 갱신 (다음 TUI 연결 시 반영)
	type patchParams struct {
		Key   string `json:"key"`
		Model string `json:"model"`
	}
	paramsBytes, _ := json.Marshal(patchParams{Key: "main", Model: primaryModel})
	params := string(paramsBytes)
	cmd2 := exec.Command(bin, "gateway", "call", "sessions.patch", "--params", params)
	cmd2.Env = baseEnv
	if _, err := cmd2.CombinedOutput(); err != nil {
		log.Printf("[openclaw-sync] sessions.patch 실패: %v", err)
	}
}

// sanitizeModelForTmux strips control characters, newlines, and other dangerous
// bytes from a model name before passing it to tmux send-keys. This prevents
// tmux command injection via crafted model names.
func sanitizeModelForTmux(model string) string {
	var b strings.Builder
	b.Grow(len(model))
	for _, r := range model {
		// Allow only printable, non-control ASCII characters (space through tilde).
		// This blocks newlines, carriage returns, escape sequences, and any unicode
		// control characters that could be interpreted by tmux or the shell.
		if r >= 0x20 && r <= 0x7e {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// injectModelToTUI: 실행 중인 tmux 패인에서 openclaw TUI를 찾아 /model 명령 주입
func injectModelToTUI(primaryModel string) {
	// Sanitize model name to prevent tmux command injection
	primaryModel = sanitizeModelForTmux(primaryModel)
	if primaryModel == "" {
		return
	}

	tmuxBin, err := exec.LookPath("tmux")
	if err != nil {
		return // tmux 없음 — 조용히 건너뜀
	}

	// 모든 tmux 패인 목록 가져오기: pane_id + current_command
	out, err := exec.Command(tmuxBin, "list-panes", "-a", "-F", "#{pane_id}|#{pane_current_command}").Output()
	if err != nil {
		return
	}

	slash := "/model " + primaryModel
	injected := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		paneID, cmd := parts[0], parts[1]
		// openclaw 또는 node 프로세스가 TUI로 실행 중인 패인만 대상
		if !strings.Contains(cmd, "openclaw") && !strings.Contains(cmd, "node") {
			continue
		}
		// /model 명령 주입
		if err := exec.Command(tmuxBin, "send-keys", "-t", paneID, slash, "Enter").Run(); err == nil {
			injected++
			log.Printf("[openclaw-sync] tmux 패인 %s에 모델 주입됨", paneID)
		}
	}
	if injected == 0 {
		log.Printf("[openclaw-sync] TUI tmux 패인 없음 (정상)")
	}
}

// updateOpenClawJSON: 모델 변경 시 ~/.openclaw/openclaw.json 갱신 후 게이트웨이 통보.
//
// Supports openclaw.json v2026.3.12+ format:
//
//	models.providers.<name>.api = "openai-completions"
//	models.providers.<name>.models[].{id, name, reasoning, input, contextWindow, maxTokens}
func updateOpenClawJSON(service, model string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return // file doesn't exist — skip silently
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("[openclaw-sync] openclaw.json 파싱 실패: %v", err)
		return
	}

	// Build the provider/model string OpenClaw expects.
	// For Anthropic: use the "anthropic" provider so OpenClaw sends to /v1/messages
	// (which uses callAnthropicPassthrough — preserves tool calls).
	// For all others: use the "custom" provider (OpenAI-compat proxy at /v1).
	var primaryModel string
	if service == "anthropic" {
		primaryModel = "anthropic/" + model
	} else {
		primaryModel = "custom/" + model
	}

	changed := false

	// ── 1. agents.defaults.model.primary ──────────────────────────────────────
	agents, _ := cfg["agents"].(map[string]interface{})
	if agents == nil {
		agents = map[string]interface{}{}
		cfg["agents"] = agents
	}
	defaults, _ := agents["defaults"].(map[string]interface{})
	if defaults == nil {
		defaults = map[string]interface{}{}
		agents["defaults"] = defaults
	}
	mdl, _ := defaults["model"].(map[string]interface{})
	if mdl == nil {
		mdl = map[string]interface{}{}
		defaults["model"] = mdl
	}
	if mdl["primary"] != primaryModel {
		mdl["primary"] = primaryModel
		changed = true
	}

	// ── 2. models.providers — set up the appropriate provider ─────────────────
	modelsSection, _ := cfg["models"].(map[string]interface{})
	if modelsSection == nil {
		modelsSection = map[string]interface{}{}
		cfg["models"] = modelsSection
	}
	providers, _ := modelsSection["providers"].(map[string]interface{})
	if providers == nil {
		providers = map[string]interface{}{}
		modelsSection["providers"] = providers
	}

	if service == "anthropic" {
		// Point the "anthropic" provider at the local proxy so /v1/messages is handled
		// by callAnthropicPassthrough (preserves tools and tool_use responses).
		antProv, _ := providers["anthropic"].(map[string]interface{})
		if antProv == nil {
			antProv = map[string]interface{}{
				"baseUrl": "http://localhost:56244",
				"apiKey":  "proxy-managed",
			}
			providers["anthropic"] = antProv
			changed = true
		}
		// Redirect from real Anthropic API to proxy if needed.
		if bu, _ := antProv["baseUrl"].(string); bu != "http://localhost:56244" {
			antProv["baseUrl"] = "http://localhost:56244"
			changed = true
		}
		// Add model entry if missing.
		antModels, _ := antProv["models"].([]interface{})
		antFound := false
		for _, m := range antModels {
			if mm, ok := m.(map[string]interface{}); ok && mm["id"] == model {
				antFound = true
				break
			}
		}
		if !antFound {
			entry := map[string]interface{}{
				"id":            model,
				"name":          "proxy / " + model,
				"contextWindow": 1048576,
				"maxTokens":     65536,
			}
			antModels = append([]interface{}{entry}, antModels...)
			antProv["models"] = antModels
			changed = true
		}
	} else {
		// Non-Anthropic: use the "custom" provider (OpenAI-compat /v1 proxy).
		custom, _ := providers["custom"].(map[string]interface{})
		if custom == nil {
			custom = map[string]interface{}{
				"baseUrl":    "http://localhost:56244/v1",
				"apiKey":     "proxy-managed",
				"api":        "openai-completions",
				"authHeader": false,
			}
			providers["custom"] = custom
			changed = true
		}

		// Ensure provider-level api field is set (migration from per-model api)
		if custom["api"] == nil {
			custom["api"] = "openai-completions"
			changed = true
		}
		// Migrate old baseUrl from Gemini path to OpenAI path
		if bu, _ := custom["baseUrl"].(string); bu == "http://localhost:56244/google/v1beta" {
			custom["baseUrl"] = "http://localhost:56244/v1"
			changed = true
		}

		// Add model entry if missing
		models, _ := custom["models"].([]interface{})
		found := false
		for _, m := range models {
			if mm, ok := m.(map[string]interface{}); ok {
				if mm["id"] == model {
					found = true
					// Remove legacy per-model api field if present
					if _, hasAPI := mm["api"]; hasAPI {
						delete(mm, "api")
						changed = true
					}
					break
				}
			}
		}
		if !found {
			entry := map[string]interface{}{
				"id":            model,
				"name":          service + " / " + model,
				"reasoning":     false,
				"input":         []interface{}{"text"},
				"contextWindow": 1048576,
				"maxTokens":     65536,
			}
			models = append([]interface{}{entry}, models...)
			custom["models"] = models
			changed = true
		}
	}

	// ── 3. meta version bump ───────────────────────────────────────────────────
	meta, _ := cfg["meta"].(map[string]interface{})
	if meta == nil {
		meta = map[string]interface{}{}
		cfg["meta"] = meta
	}
	meta["lastTouchedVersion"] = "2026.3.12"
	meta["lastTouchedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	changed = true

	if !changed {
		return
	}

	if err := writeJSON(path, cfg); err != nil {
		log.Printf("[openclaw-sync] openclaw.json 쓰기 실패: %v", err)
		return
	}

	log.Printf("[openclaw-sync] openclaw.json 갱신: primary=%s", primaryModel)

	// 실행 중인 OpenClaw 게이트웨이에 즉시 적용 (TUI WebSocket 통보 포함)
	notifyOpenClaw(primaryModel)
}
