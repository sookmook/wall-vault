package proxy

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

// notifyOpenClaw: openclaw models set으로 실행 중인 게이트웨이에 모델 변경 즉시 적용.
// TUI가 WebSocket을 통해 즉시 반응한다.
func notifyOpenClaw(primaryModel string) {
	bin := resolveOpenClawBin()
	if bin == "" {
		log.Printf("[openclaw-sync] openclaw 바이너리를 찾을 수 없음 — TUI 미통보")
		return
	}
	home, _ := os.UserHomeDir()
	cmd := exec.Command(bin, "models", "set", primaryModel)
	// systemd 서비스는 최소 환경을 가지므로 HOME과 PATH를 명시적으로 전달
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"PATH="+filepath.Dir(bin)+":/usr/local/bin:/usr/bin:/bin",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[openclaw-sync] models set 실패: %v — %s", err, string(out))
	} else {
		log.Printf("[openclaw-sync] 모델 적용됨: %s", primaryModel)
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
