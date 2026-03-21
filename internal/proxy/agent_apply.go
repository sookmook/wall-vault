package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// handleAgentApply: POST /agent/apply
// Dispatches to the correct config writer based on agentType.
// The handler writes local config files for the target agent so the user
// never has to copy-paste credentials manually.
//
// Authorization: Bearer <agent-token>
// Body: {"agentType":"cline|claude-code|openclaw|nanoclaw","baseUrl":"...","model":"..."}
func (s *Server) handleAgentApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		jsonError(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	var body struct {
		AgentType string `json:"agentType"`
		BaseURL   string `json:"baseUrl"`
		Model     string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Fill missing model from vault token config
	if body.Model == "" {
		if entry := s.lookupTokenConfig(token); entry != nil {
			body.Model = entry.model
		}
	}
	if body.BaseURL == "" {
		jsonError(w, "baseUrl is required", http.StatusBadRequest)
		return
	}

	var (
		path string
		err  error
	)
	switch body.AgentType {
	case "cline":
		path, err = applyClineConfig(body.BaseURL, body.Model, token)
	case "claude-code":
		path, err = applyClaudeCodeConfig(body.BaseURL, token)
	case "openclaw", "nanoclaw":
		path, err = applyOpenClawConfig(body.BaseURL, body.Model, token)
	default:
		jsonError(w, "unsupported agentType: "+body.AgentType, http.StatusBadRequest)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"ok":      true,
		"path":    path,
		"baseUrl": body.BaseURL,
		"model":   body.Model,
	})
}

// ── Cline ─────────────────────────────────────────────────────────────────────

// applyClineConfig writes provider/URL/model into Cline's globalState.json
// and the agent token into secrets.json.
// Returns the resolved Cline data directory path on success.
func applyClineConfig(baseURL, model, token string) (string, error) {
	dataDir, err := findClineDataDir()
	if err != nil {
		return "", err
	}
	if err := updateClineGlobalState(dataDir, baseURL, model); err != nil {
		return "", fmt.Errorf("failed to write globalState.json: %w", err)
	}
	if err := updateClineSecrets(dataDir, token); err != nil {
		return "", fmt.Errorf("failed to write secrets.json: %w", err)
	}
	return dataDir, nil
}

// findClineDataDir returns the path to the Cline data directory.
// Search order:
//  1. $HOME/.cline/data           — native Linux / macOS
//  2. WSL mount from $USERPROFILE — e.g. /mnt/c/Users/sookmook/.cline/data
//  3. Scan /mnt/c/Users/*/        — WSL fallback when $USERPROFILE is unset
func findClineDataDir() (string, error) {
	var candidates []string

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".cline", "data"))
	}
	if up := os.Getenv("USERPROFILE"); up != "" {
		if wsl := windowsPathToWSL(up); wsl != "" {
			candidates = append(candidates, filepath.Join(wsl, ".cline", "data"))
		}
	}
	if entries, err := os.ReadDir("/mnt/c/Users"); err == nil {
		for _, e := range entries {
			if !e.IsDir() || e.Name() == "Public" || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			candidates = append(candidates, filepath.Join("/mnt/c/Users", e.Name(), ".cline", "data"))
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	tried := strings.Join(candidates, "\n  ")
	if tried == "" {
		tried = "(none found)"
	}
	return "", fmt.Errorf("~/.cline/data/ not found. Is Cline installed?\nSearched:\n  %s", tried)
}

// updateClineGlobalState merges wall-vault proxy settings into Cline's globalState.json.
func updateClineGlobalState(dataDir, baseURL, model string) error {
	path := filepath.Join(dataDir, "globalState.json")

	state := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &state)
	}

	state["actModeApiProvider"] = "openai"
	state["planModeApiProvider"] = "openai"
	state["openAiBaseUrl"] = baseURL
	state["actModeApiModelId"] = model
	state["planModeApiModelId"] = model
	state["openAiModelId"] = model

	return writeJSON(path, state)
}

// updateClineSecrets merges the wall-vault agent token into Cline's secrets.json.
func updateClineSecrets(dataDir, apiKey string) error {
	path := filepath.Join(dataDir, "secrets.json")

	secrets := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &secrets)
	}

	secrets["openAiApiKey"] = apiKey

	return writeJSON(path, secrets)
}

// ── Claude Code ───────────────────────────────────────────────────────────────

// applyClaudeCodeConfig writes wall-vault proxy settings into ~/.claude/settings.json.
// Returns the resolved settings.json path on success.
func applyClaudeCodeConfig(baseURL, token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	path := filepath.Join(home, ".claude", "settings.json")

	settings := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &settings)
	}

	settings["apiProvider"] = "openai"
	settings["baseUrl"] = baseURL
	settings["apiKey"] = token

	if err := writeJSON(path, settings); err != nil {
		return "", fmt.Errorf("failed to write settings.json: %w", err)
	}
	return path, nil
}

// ── OpenClaw / NanoClaw ───────────────────────────────────────────────────────

// applyOpenClawConfig writes wall-vault proxy settings into ~/.openclaw/openclaw.json.
// Updates models.providers.custom (baseUrl + apiKey) and agents.defaults.model.primary.
// Returns the resolved openclaw.json path on success.
func applyOpenClawConfig(baseURL, model, token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")

	cfg := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// models.providers.custom
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
		custom = map[string]interface{}{}
		providers["custom"] = custom
	}
	custom["baseUrl"] = baseURL
	custom["apiKey"] = token
	custom["api"] = "openai-completions"
	custom["authHeader"] = false

	// Ensure model entry exists
	if model != "" {
		models, _ := custom["models"].([]interface{})
		found := false
		for _, m := range models {
			if mm, ok := m.(map[string]interface{}); ok && mm["id"] == model {
				found = true
				break
			}
		}
		if !found {
			entry := map[string]interface{}{
				"id":            model,
				"name":          model,
				"reasoning":     false,
				"input":         []interface{}{"text"},
				"contextWindow": 1048576,
				"maxTokens":     65536,
			}
			custom["models"] = append([]interface{}{entry}, models...)
		}
	}

	// agents.defaults.model.primary
	if model != "" {
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
		mdl["primary"] = "custom/" + model
	}

	// meta
	meta, _ := cfg["meta"].(map[string]interface{})
	if meta == nil {
		meta = map[string]interface{}{}
		cfg["meta"] = meta
	}
	meta["lastTouchedAt"] = time.Now().UTC().Format(time.RFC3339)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to create .openclaw directory: %w", err)
	}
	if err := writeJSON(path, cfg); err != nil {
		return "", fmt.Errorf("failed to write openclaw.json: %w", err)
	}
	return path, nil
}

// ── Shared helpers ─────────────────────────────────────────────────────────────

// windowsPathToWSL converts a Windows-style path (C:\Users\foo) to a WSL mount path (/mnt/c/Users/foo).
func windowsPathToWSL(winPath string) string {
	p := strings.ReplaceAll(winPath, "\\", "/")
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		return "/mnt/" + drive + p[2:]
	}
	return ""
}

// writeJSON writes v as indented JSON to path, creating the file if needed.
func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
