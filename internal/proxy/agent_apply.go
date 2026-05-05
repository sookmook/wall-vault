package proxy

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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
	case "openclaw":
		path, err = applyOpenClawConfig(body.BaseURL, body.Model, token)
	case "nanoclaw":
		path, err = applyNanoclawConfig(body.BaseURL, body.Model, token)
	case "econoworld":
		path, err = applyEconoWorldConfig(body.BaseURL, body.Model, token, body.WorkDir, s.cfg.Proxy.EconoWorldMaxTokens, s.cfg.Proxy.EconoWorldStream, s.cfg.Proxy.EconoWorldRequestTimeout)
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
//  2. WSL mount from $USERPROFILE — e.g. /mnt/c/Users/<windows-user>/.cline/data
//  3. Scan /mnt/c/Users/*/        — WSL fallback when $USERPROFILE is unset
func findClineDataDir() (string, error) {
	candidates := wslHomeCandidates(".cline", "data")
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
	return wslHomeCandidates(".claude", "settings.json")
}

// ── OpenClaw / NanoClaw ───────────────────────────────────────────────────────

// applyOpenClawConfig writes wall-vault proxy settings into ~/.openclaw/openclaw.json.
// Updates models.providers.custom (baseUrl + apiKey) and agents.defaults.model.primary.
// Returns the resolved openclaw.json path on success.
//
// Version awareness: detectOpenClawVersion() reads the local OpenClaw
// package.json so the writer can adapt when a future release forks the
// config schema. Every 2026.x release tested so far shares schemaTag()=v1
// and the same write path; the version is logged + written to
// meta.lastTouchedVersion for diagnostics.
func applyOpenClawConfig(baseURL, model, token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")
	version := detectOpenClawVersion()
	log.Printf("[agent-apply] openclaw version=%s path=%s", version.describe(), path)
	switch version.schemaTag() {
	case "v1":
		// fall through — every reachable version uses the v1 writer below.
	default:
		// Defensive: an unknown schemaTag should still attempt the v1 write
		// rather than refusing — the operator can always edit by hand if
		// the schema actually diverges in a way we haven't taught yet.
		log.Printf("[agent-apply] openclaw schema %q unrecognised, falling back to v1 writer",
			version.schemaTag())
	}

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
	// authHeader=true so OpenClaw forwards the token as
	// `Authorization: Bearer <token>`. Pre-v0.2.37 this was false (no header
	// was sent), which the new proxy gate rejects with 401.
	custom["authHeader"] = true

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
	if version.Raw != "" {
		meta["lastTouchedVersion"] = version.Raw
		meta["lastTouchedSchema"] = version.schemaTag()
	}

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

// ── NanoClaw ──────────────────────────────────────────────────────────────────

// applyNanoclawConfig writes wall-vault proxy settings into NanoClaw's
// ~/nanoclaw/.env (a dotenv file loaded by systemd's EnvironmentFile).
// Updates ANTHROPIC_BASE_URL and ONECLI_API_KEY in place; preserves any
// other lines (TELEGRAM_BOT_TOKEN etc.) untouched. Re-issues the file
// at 0600 since it holds bot tokens + the wall-vault credential.
func applyNanoclawConfig(baseURL, _ /*model*/, token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	path := filepath.Join(home, "nanoclaw", ".env")

	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(string(data), "\n")
		// Drop a single trailing empty line introduced by the final newline,
		// so we don't blow up the file with extra blanks on every rewrite.
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	setLine := func(in []string, key, value string) []string {
		prefix := key + "="
		for i, l := range in {
			if strings.HasPrefix(l, prefix) {
				in[i] = prefix + value
				return in
			}
		}
		return append(in, prefix+value)
	}
	lines = setLine(lines, "ANTHROPIC_BASE_URL", baseURL)
	lines = setLine(lines, "ONECLI_API_KEY", token)

	out := strings.Join(lines, "\n") + "\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(out), 0o600); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
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

// econoWorldDirCache remembers the workDir an /agent/apply call resolved to,
// so SSE-driven model refreshes and startup heals operate on exactly the
// directory the operator pinned via the dashboard's work_dir field. Without
// this cache, updateEconoWorldModel and healEconoWorldConfig fall back to
// resolveEconoWorldDir(""), which only inspects the hard-coded default
// candidate list — when the install lives outside that list, the cache is
// what keeps SSE refreshes pointed at the same file /agent/apply just wrote.
//
// The cache is process-local (not persisted) so a proxy restart re-learns
// the dir on the next /agent/apply, and the heal path's resolveEconoWorldDir("")
// fallback handles the gap before the first apply lands.
var econoWorldDirCache atomic.Pointer[string]

func setCachedEconoWorldDir(dir string) {
	if dir == "" {
		return
	}
	if v := os.Getenv("WV_ECONOWORLD_WORKDIR_CACHE_DISABLE"); v == "1" || strings.EqualFold(v, "true") {
		return
	}
	d := dir
	econoWorldDirCache.Store(&d)
}

func cachedEconoWorldDir() string {
	if v := os.Getenv("WV_ECONOWORLD_WORKDIR_CACHE_DISABLE"); v == "1" || strings.EqualFold(v, "true") {
		return ""
	}
	if p := econoWorldDirCache.Load(); p != nil {
		return *p
	}
	return ""
}

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
func applyEconoWorldConfig(baseURL, model, token, workDir string, maxTokens int, stream bool, requestTimeout int) (string, error) {
	dir := resolveEconoWorldDir(workDir)
	setCachedEconoWorldDir(dir)
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
	// Robustness fields below all use the "fill if absent" pattern so
	// operator-tuned values survive a re-bootstrap. maxTokens=0 means "skip
	// even on first write" so a deployment that genuinely wants the
	// EconoWorld default does not get an unwanted floor.
	if _, ok := compat["max_tokens"]; !ok {
		if maxTokens > 0 {
			compat["max_tokens"] = maxTokens
		} else {
			compat["max_tokens"] = 4096
		}
	}
	if _, ok := compat["stream"]; !ok {
		compat["stream"] = stream
	}
	if _, ok := compat["request_timeout_seconds"]; !ok && requestTimeout > 0 {
		compat["request_timeout_seconds"] = requestTimeout
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to create analyzer directory: %w", err)
	}
	if err := writeJSON(path, cfg); err != nil {
		return "", fmt.Errorf("failed to write ai_config.json: %w", err)
	}
	return path, nil
}

// updateEconoWorldModel is the SSE-driven counterpart to applyEconoWorldConfig:
// when a vault config_change event arrives for an econoworld client, we only
// need to refresh the `openai_compatible.model` field in the local
// ai_config.json. base_url / api_key / max_tokens stay as last bootstrapped.
// Hosts without EconoWorld installed silently skip (no ai_config.json to open).
func updateEconoWorldModel(model string) {
	if model == "" {
		return
	}
	dir := cachedEconoWorldDir()
	if dir == "" {
		dir = resolveEconoWorldDir("")
	}
	path := filepath.Join(dir, "analyzer", "ai_config.json")
	if err := updateEconoWorldModelAt(path, model); err != nil {
		log.Printf("[econoworld] updateModel failed: %v", err)
	}
}

// updateEconoWorldModelAt is the file-path-taking core of updateEconoWorldModel,
// split out so it can be unit-tested against a temp directory.
func updateEconoWorldModelAt(path, model string) error {
	if model == "" {
		return nil // empty model means "cleared by vault" — keep bootstrap intact
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // host doesn't have EconoWorld — silent skip
		}
		return err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	compat, _ := cfg["openai_compatible"].(map[string]interface{})
	if compat == nil {
		// ai_config.json exists but hasn't been bootstrapped with the
		// openai_compatible section yet — leave it to the next /agent/apply.
		return nil
	}
	compat["model"] = model
	return writeJSON(path, cfg)
}

// healEconoWorldConfig idempotently fixes a stale EconoWorld ai_config.json
// at startup. Three things can drift between bootstrap and runtime:
//
//  1. base_url points at https://localhost:56244/v1 but TLS trust on the
//     EconoWorld host is broken — the loopback plain-HTTP companion
//     (http://127.0.0.1:<plain_port>) sidesteps it. Heal only rewrites
//     same-host (localhost / 127.0.0.1) URLs; external base_urls are
//     never touched, so a remote EconoWorld pointed at our LAN address
//     keeps working.
//  2. stream / request_timeout_seconds were never written because the
//     bootstrap predates v0.2.57. Heal fills them in once.
//  3. max_tokens=4096 from the old hard-coded floor truncates Korean
//     analyses; if the configured floor is higher and the file value is
//     exactly 4096, bump it. Any other value (operator-tuned) is kept.
//
// Hosts without an EconoWorld install get a silent skip — the file does
// not exist, the function returns immediately. localBaseOrigin "" disables
// the URL rewrite leg; the field-fill leg still runs.
func healEconoWorldConfig(localBaseOrigin string, maxTokens int, stream bool, requestTimeout int) {
	dir := cachedEconoWorldDir()
	if dir == "" {
		dir = resolveEconoWorldDir("")
	}
	path := filepath.Join(dir, "analyzer", "ai_config.json")
	if err := healEconoWorldConfigAt(path, localBaseOrigin, maxTokens, stream, requestTimeout); err != nil {
		log.Printf("[econoworld-heal] %s: %v", path, err)
	}
}

// healEconoWorldConfigAt is the path-taking core of healEconoWorldConfig,
// split out so it can be unit-tested against a temp directory. Returns nil
// for missing files (silent skip) and for the no-change case so callers
// can log only on parse / write failures.
func healEconoWorldConfigAt(path, localBaseOrigin string, maxTokens int, stream bool, requestTimeout int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	compat, _ := cfg["openai_compatible"].(map[string]interface{})
	if compat == nil {
		return nil // not yet bootstrapped — /agent/apply will handle it
	}
	changed := false

	// Leg 1 — same-host URL rewrite. Preserve the path suffix (typically
	// /v1) so we don't drop client-required path components.
	if localBaseOrigin != "" {
		if cur, _ := compat["base_url"].(string); cur != "" && isLocalProxyBaseURL(cur) {
			suffix := pathSuffixAfterAuthority(cur)
			want := strings.TrimRight(localBaseOrigin, "/") + suffix
			if cur != want {
				compat["base_url"] = want
				changed = true
			}
		}
	}

	// Leg 2 — fill missing robustness fields. Operator-set values stay.
	if _, ok := compat["stream"]; !ok {
		compat["stream"] = stream
		changed = true
	}
	if _, ok := compat["request_timeout_seconds"]; !ok && requestTimeout > 0 {
		compat["request_timeout_seconds"] = requestTimeout
		changed = true
	}
	// Leg 3 — only the legacy 4096 default is replaced. Numbers reach us
	// as float64 after json.Unmarshal so cover both forms.
	if maxTokens > 0 {
		switch v := compat["max_tokens"].(type) {
		case float64:
			if int(v) == 4096 && maxTokens != 4096 {
				compat["max_tokens"] = maxTokens
				changed = true
			}
		case int:
			if v == 4096 && maxTokens != 4096 {
				compat["max_tokens"] = maxTokens
				changed = true
			}
		}
	}

	// Leg 4 — provider normalize. /agent/apply writes "openai_compatible"
	// unconditionally so wall-vault stays in the dispatch path, but a manual
	// edit (or a third-party tool that rewrites the file) can flip provider
	// back to "ollama", which routes every analyzer call straight to the
	// upstream backend with whatever model the file lists. Flipping it back
	// to "openai_compatible" restores wall-vault as the single dispatch
	// point — model and key choices then come from the dashboard / vault.
	// Only acts when the openai_compatible section is already present (i.e.
	// the file has been bootstrapped at least once); otherwise heal returns
	// up at the compat == nil check before reaching this leg.
	if prov, _ := cfg["provider"].(string); prov == "ollama" {
		cfg["provider"] = "openai_compatible"
		changed = true
	}

	if !changed {
		return nil
	}
	if err := writeJSON(path, cfg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	log.Printf("[econoworld-heal] normalized %s", path)
	return nil
}

// pathSuffixAfterAuthority returns the URL portion starting at the first
// "/" after "scheme://host[:port]". Empty string when no path component
// is present. Used by healEconoWorldConfig to keep the original /v1 (or
// any other suffix) when rewriting just the origin.
func pathSuffixAfterAuthority(u string) string {
	for _, scheme := range []string{"http://", "https://"} {
		if strings.HasPrefix(u, scheme) {
			rest := u[len(scheme):]
			if i := strings.Index(rest, "/"); i >= 0 {
				return rest[i:]
			}
			return ""
		}
	}
	return ""
}

// runStartupEconoWorldHeal mirrors runStartupOpenClawHeal: fires once
// shortly after boot in a goroutine so a slow disk doesn't delay startup.
// Disabled (no goroutine) when WV_ECONOWORLD_HEAL_DISABLE is truthy so
// operators can fully opt out without recompiling.
func runStartupEconoWorldHeal(localBaseOrigin string, maxTokens int, stream bool, requestTimeout int) {
	if v := os.Getenv("WV_ECONOWORLD_HEAL_DISABLE"); v == "1" || strings.EqualFold(v, "true") {
		return
	}
	go func() {
		time.Sleep(900 * time.Millisecond)
		healEconoWorldConfig(localBaseOrigin, maxTokens, stream, requestTimeout)
	}()
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

// wslHomeCandidates returns the list of plausible per-user roots for an
// agent that may live in either the WSL native home or under a Windows user
// profile reachable via /mnt/c/Users/<name>. The relative path components
// are appended to each root. Search order:
//
//  1. $HOME (Linux / macOS / WSL native)
//  2. WSL mount of $USERPROFILE (when running inside WSL with the env set)
//  3. Each /mnt/c/Users/<entry>/ dir, skipping Public and hidden entries
//
// When findClineDataDir / findClaudeSettingsPaths used to enumerate this
// list inline, the order and filtering rules drifted between callers;
// keeping it in one place removes the drift and makes adding a fourth root
// (e.g. /mnt/d/Users) a single edit.
func wslHomeCandidates(rel ...string) []string {
	var out []string
	if home, err := os.UserHomeDir(); err == nil {
		out = append(out, filepath.Join(append([]string{home}, rel...)...))
	}
	if up := os.Getenv("USERPROFILE"); up != "" {
		if wsl := windowsPathToWSL(up); wsl != "" {
			out = append(out, filepath.Join(append([]string{wsl}, rel...)...))
		}
	}
	if entries, err := os.ReadDir("/mnt/c/Users"); err == nil {
		for _, e := range entries {
			if !e.IsDir() || e.Name() == "Public" || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			out = append(out, filepath.Join(append([]string{"/mnt/c/Users", e.Name()}, rel...)...))
		}
	}
	return out
}

// windowsPathToWSL converts a Windows-style path (C:\Users\foo) to a WSL mount path (/mnt/c/Users/foo).
func windowsPathToWSL(winPath string) string {
	p := strings.ReplaceAll(winPath, "\\", "/")
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		return "/mnt/" + drive + p[2:]
	}
	return ""
}

// writeJSON writes v as indented JSON to path using an atomic temp+rename
// pattern. If the process is killed between the temp write and rename, the
// original file survives intact — this prevents the 0-byte clobber that
// occurs when os.WriteFile's internal O_TRUNC runs but the data write
// doesn't complete (observed during pkill -x wall-vault deploys).
//
// Also re-checks the parent directory exists: findClineDataDir / other
// discovery helpers return a path that was valid at lookup time, but the
// user may have uninstalled or moved the agent between discovery and the
// write (TOCTOU). Failing early with a clear "parent missing" message
// beats a cryptic os.Rename error in the logs.
func writeJSON(path string, v interface{}) error {
	parent := filepath.Dir(path)
	if fi, err := os.Stat(parent); err != nil {
		return fmt.Errorf("agent config parent %s missing: %w", parent, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("agent config parent %s is not a directory", parent)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		// Leave the .tmp behind for inspection; caller will surface the error.
		return fmt.Errorf("rename %s → %s: %w", tmp, path, err)
	}
	return nil
}
