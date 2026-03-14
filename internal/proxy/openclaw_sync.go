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
	params := `{"key":"main","model":"` + primaryModel + `"}`
	cmd2 := exec.Command(bin, "gateway", "call", "sessions.patch", "--params", params)
	cmd2.Env = baseEnv
	if _, err := cmd2.CombinedOutput(); err != nil {
		log.Printf("[openclaw-sync] sessions.patch 실패: %v", err)
	}
}

// injectModelToTUI: 실행 중인 tmux 패인에서 openclaw TUI를 찾아 /model 명령 주입
func injectModelToTUI(primaryModel string) {
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

	// Build the provider/model string OpenClaw expects (e.g. "custom/gemini-2.5-flash")
	primaryModel := "custom/" + model

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

	// ── 2. models.providers.custom — v2026.3.12 format ────────────────────────
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

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[openclaw-sync] marshal 실패: %v", err)
		return
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		log.Printf("[openclaw-sync] openclaw.json 쓰기 실패: %v", err)
		return
	}

	log.Printf("[openclaw-sync] openclaw.json 갱신: primary=%s", primaryModel)

	// 실행 중인 OpenClaw 게이트웨이에 즉시 적용 (TUI WebSocket 통보 포함)
	notifyOpenClaw(primaryModel)
}
