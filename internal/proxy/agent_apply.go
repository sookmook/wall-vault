package proxy

import (
	"crypto/subtle"
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

	// Validate token: must be either the proxy's own vault token or a known client token.
	// Without this check, any non-empty token could write local agent config files.
	if !s.isValidAgentToken(token) {
		jsonError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxConfigBodySize)
	var body struct {
		AgentType string `json:"agentType"`
		BaseURL   string `json:"baseUrl"`
		Model     string `json:"model"`
		WorkDir   string `json:"workDir"`
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
		path, err = applyClaudeCodeConfig(body.BaseURL, body.Model, token)
	case "openclaw", "nanoclaw":
		path, err = applyOpenClawConfig(body.BaseURL, body.Model, token)
	case "econoworld":
		path, err = applyEconoWorldConfig(body.BaseURL, body.Model, token, body.WorkDir)
	default:
		jsonError(w, "unsupported agentType: "+body.AgentType, http.StatusBadRequest)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Record applied client so heartbeat keeps reporting it to vault (signal light)
	if entry := s.lookupTokenConfig(token); entry != nil && entry.clientID != "" {
		s.clientActMu.Lock()
		s.clientActs[entry.clientID] = &clientAct{
			service:  entry.service,
			model:    entry.model,
			lastSeen: time.Now(),
			applied:  true,
		}
		s.clientActMu.Unlock()
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
	state["actModeOpenAiModelId"] = model
	state["planModeOpenAiModelId"] = model
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
func applyClaudeCodeConfig(baseURL, model, token string) (string, error) {
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
	if model != "" {
		settings["model"] = model
	}

	if err := writeJSON(path, settings); err != nil {
		return "", fmt.Errorf("failed to write settings.json: %w", err)
	}
	return path, nil
}

// updateClaudeCodeModel: update Claude Code's model in settings.json when vault config changes.
// Checks both WSL (~/.claude/) and Windows (/mnt/c/Users/.../.claude/) locations.
// Only writes Claude-compatible model names; non-Claude models (e.g. google/gemini-*)
// are handled by the proxy's token-based routing and must NOT be written to settings.json
// because Claude Code validates model names against its internal registry.
// Best-effort: errors are silently ignored.
func updateClaudeCodeModel(model string) {
	if model == "" {
		return
	}
	// Only write Claude-compatible model names to settings.json.
	// Claude Code rejects non-Claude models (shows "There's an issue with the selected model").
	// The proxy routes non-Claude models via token-based lookup in handleOpenAI — no settings
	// change needed.
	if !isClaudeModel(model) {
		return
	}
	for _, path := range findClaudeSettingsPaths() {
		settings := map[string]interface{}{}
		if data, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(data, &settings)
		} else {
			continue // file doesn't exist at this location
		}
		settings["model"] = model
		_ = writeJSON(path, settings)
	}
}

// isClaudeModel returns true if the model name is a Claude-compatible identifier
// that Claude Code's internal model registry accepts.
func isClaudeModel(model string) bool {
	// Strip provider prefix if present (e.g. "anthropic/claude-opus-4-6")
	if i := strings.LastIndex(model, "/"); i >= 0 {
		model = model[i+1:]
	}
	// Full model IDs: claude-opus-4-6, claude-sonnet-4-6, claude-haiku-4-5-20251001, etc.
	if strings.HasPrefix(model, "claude-") {
		return true
	}
	// Short aliases used by Claude Code: opus, sonnet, haiku
	switch model {
	case "opus", "sonnet", "haiku":
		return true
	}
	return false
}

// findClaudeSettingsPaths returns all candidate paths for Claude Code's settings.json.
func findClaudeSettingsPaths() []string {
	var paths []string
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".claude", "settings.json"))
	}
	if up := os.Getenv("USERPROFILE"); up != "" {
		if wsl := windowsPathToWSL(up); wsl != "" {
			paths = append(paths, filepath.Join(wsl, ".claude", "settings.json"))
		}
	}
	if entries, err := os.ReadDir("/mnt/c/Users"); err == nil {
		for _, e := range entries {
			if !e.IsDir() || e.Name() == "Public" || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			paths = append(paths, filepath.Join("/mnt/c/Users", e.Name(), ".claude", "settings.json"))
		}
	}
	return paths
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

	// Sanitize any previously-written entries with an empty id. A pre-guard
	// version of this function, or an external writer, could have left a
	// `{"id": ""}` entry in the list — OpenClaw's config validator rejects
	// the whole file in that state and crash-loops. Filter every call.
	if models, ok := custom["models"].([]interface{}); ok {
		filtered := make([]interface{}, 0, len(models))
		for _, m := range models {
			mm, ok := m.(map[string]interface{})
			if !ok {
				continue
			}
			if id, _ := mm["id"].(string); id != "" {
				filtered = append(filtered, m)
			}
		}
		custom["models"] = filtered
	}

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

// updateClineModel: update only model fields in Cline's globalState.json when vault config changes.
// Preserves the existing openAiBaseUrl — does NOT overwrite it.
// Best-effort: errors are silently ignored (Cline may not be installed on this machine).
func updateClineModel(model string) {
	if model == "" {
		return
	}
	dataDir, err := findClineDataDir()
	if err != nil {
		return // Cline not found on this machine — skip silently
	}
	path := filepath.Join(dataDir, "globalState.json")
	state := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &state)
	}
	state["actModeApiModelId"] = model
	state["planModeApiModelId"] = model
	state["actModeOpenAiModelId"] = model
	state["planModeOpenAiModelId"] = model
	state["openAiModelId"] = model
	_ = writeJSON(path, state)
}

// isValidAgentToken checks whether a Bearer token is authorized to call /agent/apply.
// Accepted tokens: the proxy's own vault token, or any token that the vault recognizes
// as a registered client (verified via lookupTokenConfig which calls /api/token/config).
func (s *Server) isValidAgentToken(token string) bool {
	// Accept the proxy's own vault token
	if s.cfg.Proxy.VaultToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.Proxy.VaultToken)) == 1 {
		return true
	}
	// Accept any token registered as a vault client
	if entry := s.lookupTokenConfig(token); entry != nil {
		return true
	}
	return false
}

// ── EconoWorld ────────────────────────────────────────────────────────────────

// applyEconoWorldConfig writes wall-vault proxy settings into EconoWorld's
// analyzer/ai_config.json. The existing multi-provider structure is preserved;
// we flip `provider` to "openai_compatible" and populate that section with the
// wall-vault base URL, the caller's token, and the requested model.
//
// workDir accepts a comma-separated list of candidate project roots and picks
// the first one whose analyzer/ directory already exists on this host. Paths
// may be either POSIX (/mnt/e/...) or Windows style (E:\Work\...); Windows
// drive paths are converted to their WSL mount equivalents. When none of the
// candidates exist the first listed candidate (or the hard-coded default) is
// used so the resulting error clearly points at the missing directory.
func applyEconoWorldConfig(baseURL, model, token, workDir string) (string, error) {
	dir := resolveEconoWorldDir(workDir)
	path := filepath.Join(dir, "analyzer", "ai_config.json")

	cfg := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	cfg["provider"] = "openai_compatible"
	compat, _ := cfg["openai_compatible"].(map[string]interface{})
	if compat == nil {
		compat = map[string]interface{}{}
		cfg["openai_compatible"] = compat
	}
	compat["base_url"] = baseURL
	compat["api_key"] = token
	if model != "" {
		compat["model"] = model
	}
	if _, ok := compat["max_tokens"]; !ok {
		compat["max_tokens"] = 4096
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to create analyzer directory: %w", err)
	}
	if err := writeJSON(path, cfg); err != nil {
		return "", fmt.Errorf("failed to write ai_config.json: %w", err)
	}
	return path, nil
}

// resolveEconoWorldDir returns the first candidate POSIX directory whose
// analyzer/ subdirectory exists on the host. workDir may be a comma-separated
// list; each item is trimmed and Windows drive paths are converted to WSL
// mount paths. If none of the candidates have an analyzer/ directory, the
// first candidate (or the hard-coded default) is returned so downstream file
// operations produce an obviously targeted error message.
func resolveEconoWorldDir(workDir string) string {
	candidates := splitEconoWorldDirs(workDir)
	if len(candidates) == 0 {
		return "/mnt/e/Work/Dev/EconoWorld"
	}
	for _, c := range candidates {
		if info, err := os.Stat(filepath.Join(c, "analyzer")); err == nil && info.IsDir() {
			return c
		}
	}
	return candidates[0]
}

// splitEconoWorldDirs parses a comma-delimited workDir string into a slice of
// normalised candidate directories (Windows drive paths rewritten to
// /mnt/<drive>/...). Empty segments are dropped.
func splitEconoWorldDirs(workDir string) []string {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return nil
	}
	parts := strings.Split(workDir, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len(p) >= 2 && p[1] == ':' {
			if wsl := windowsPathToWSL(p); wsl != "" {
				p = wsl
			}
		}
		out = append(out, p)
	}
	return out
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
