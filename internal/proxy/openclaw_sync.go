package proxy

import (
	"encoding/json"
	"fmt"
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
// tmux command injection via crafted model names. Returns the cleaned value and
// a boolean indicating whether any runes were dropped — callers use that to
// avoid injecting a silently-mangled model name (e.g. a Korean alias reduced
// to the empty string, which would become a bare `/model` command and confuse
// the OpenClaw TUI).
func sanitizeModelForTmux(model string) (string, bool) {
	var b strings.Builder
	b.Grow(len(model))
	dropped := false
	for _, r := range model {
		// Allow only printable, non-control ASCII characters (space through tilde).
		// This blocks newlines, carriage returns, escape sequences, and any unicode
		// control characters that could be interpreted by tmux or the shell.
		if r >= 0x20 && r <= 0x7e {
			b.WriteRune(r)
		} else {
			dropped = true
		}
	}
	return b.String(), dropped
}

// injectModelToTUI: 실행 중인 tmux 패인에서 openclaw TUI를 찾아 /model 명령 주입
func injectModelToTUI(primaryModel string) {
	// Sanitize model name to prevent tmux command injection
	original := primaryModel
	sanitized, dropped := sanitizeModelForTmux(primaryModel)
	if dropped {
		log.Printf("[openclaw-sync] tmux: dropped non-ASCII characters from model %q → %q", original, sanitized)
	}
	if sanitized == "" {
		log.Printf("[openclaw-sync] tmux: sanitized model is empty (original %q), skipping injection", original)
		return
	}
	primaryModel = sanitized

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

// DefaultProxyOrigin returns the canonical origin (`scheme://localhost:<port>`)
// other components should write into config files when no plain-HTTP companion
// is configured. Honours the operator's proxy.port + proxy.tls.enabled rather
// than the previous hardcoded `https://localhost:56244`, which silently broke
// any install that moved the proxy to a non-default port.
func DefaultProxyOrigin(port int, tlsEnabled bool) string {
	if port <= 0 {
		port = 56244
	}
	scheme := "http"
	if tlsEnabled {
		scheme = "https"
	}
	return fmt.Sprintf("%s://localhost:%d", scheme, port)
}

// updateOpenClawJSON: 모델 변경 시 ~/.openclaw/openclaw.json 갱신 후 게이트웨이 통보.
//
// Supports openclaw.json v2026.3.12+ format:
//
//	models.providers.<name>.api = "openai-completions"
//	models.providers.<name>.models[].{id, name, reasoning, input, contextWindow, maxTokens}
//
// defaultOrigin is the origin (e.g. `https://localhost:56244` or
// `http://localhost:7777`) heal writes into provider baseUrl fields when no
// plain-HTTP companion is in play. Empty string falls back to the legacy
// `https://localhost:56244` for backwards compatibility with callers that
// still don't pass it through.
func updateOpenClawJSON(service, model, defaultOrigin string) {
	if defaultOrigin == "" {
		defaultOrigin = "https://localhost:56244"
	}
	// Early-return on an empty model: the SSE config_change path can fire
	// before the vault has resolved a model for this client (or after a
	// soft-clear), and writing primary="<service>/" produces an unresolvable
	// reference that OpenClaw then rejects with `Invalid model reference`
	// ((operator host, earlier): every restart wrote primary="custom/" and broke
	// the gateway's model selection until a real model arrived later).
	if model == "" {
		return
	}
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
				"baseUrl": defaultOrigin,
				"apiKey":  "proxy-managed",
			}
			providers["anthropic"] = antProv
			changed = true
		}
		// Force baseUrl back to the local proxy. A previous wall-vault
		// release (or a hand edit) could have written an upstream host
		// directly here — observed on (operator host, earlier) with
		// http://<internal-host>:11434, which bypasses the proxy entirely
		// and breaks routing/auth for any provider but ollama.
		if bu, _ := antProv["baseUrl"].(string); !isLocalProxyBaseURL(bu) {
			antProv["baseUrl"] = defaultOrigin
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
				"baseUrl":    defaultOrigin + "/v1",
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
		// Force baseUrl back to the local proxy. Same reason as the
		// anthropic branch above — stale config could be pointing at an
		// upstream host ((operator host, earlier): http://<internal-host>:11434/v1
		// directly to ollama, bypassing the proxy).
		if bu, _ := custom["baseUrl"].(string); bu != "" && !isLocalProxyBaseURL(bu) {
			custom["baseUrl"] = defaultOrigin + "/v1"
			changed = true
		}
		// Migrate legacy Gemini path to the OpenAI-compat path. The hardcoded
		// 56244 string is intentional here — it matches the historic value
		// every prior release wrote, so older configs migrate forward
		// regardless of the operator's current port choice.
		if bu, _ := custom["baseUrl"].(string); bu == "https://localhost:56244/google/v1beta" {
			custom["baseUrl"] = defaultOrigin + "/v1"
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

	// Prune stale models[] entries (duplicate ids, half-written names) from
	// every provider, not just the one we just touched. A historic host-A
	// config carried 11 entries with identical id but different name, which
	// effectively broke OpenClaw model selection.
	if pruneStaleModelsAcross(providers) {
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

	if err := writeJSON(path, cfg); err != nil {
		log.Printf("[openclaw-sync] openclaw.json 쓰기 실패: %v", err)
		return
	}

	log.Printf("[openclaw-sync] openclaw.json 갱신: primary=%s", primaryModel)

	// 실행 중인 OpenClaw 게이트웨이에 즉시 적용 (TUI WebSocket 통보 포함)
	notifyOpenClaw(primaryModel)
}

// isLocalProxyBaseURL reports whether u points at this host (localhost or
// 127.0.0.1) on any port. Anything else means the OpenClaw config has an
// upstream URL written into it directly, which bypasses the proxy.
func isLocalProxyBaseURL(u string) bool {
	if u == "" {
		return false
	}
	return strings.HasPrefix(u, "http://localhost:") ||
		strings.HasPrefix(u, "https://localhost:") ||
		strings.HasPrefix(u, "http://127.0.0.1:") ||
		strings.HasPrefix(u, "https://127.0.0.1:")
}

// pruneStaleModelsAcross filters the models[] of every provider in-place,
// dropping entries that are obviously broken: empty id, name ending in a
// dangling "/" (e.g. "openrouter / "), or duplicate id within the same
// provider (keep first occurrence). Returns true when anything was removed.
func pruneStaleModelsAcross(providers map[string]interface{}) bool {
	changed := false
	for _, pv := range providers {
		p, ok := pv.(map[string]interface{})
		if !ok {
			continue
		}
		raw, ok := p["models"].([]interface{})
		if !ok {
			continue
		}
		seen := map[string]bool{}
		filtered := make([]interface{}, 0, len(raw))
		for _, m := range raw {
			mm, ok := m.(map[string]interface{})
			if !ok {
				filtered = append(filtered, m)
				continue
			}
			id, _ := mm["id"].(string)
			if id == "" {
				continue
			}
			nm, _ := mm["name"].(string)
			if strings.HasSuffix(strings.TrimSpace(nm), "/") {
				continue
			}
			if seen[id] {
				continue
			}
			seen[id] = true
			filtered = append(filtered, m)
		}
		if len(filtered) != len(raw) {
			p["models"] = filtered
			changed = true
		}
	}
	return changed
}

// healOpenClawConfig forcibly normalizes ~/.openclaw/openclaw.json on every
// proxy boot. Earlier wall-vault releases (and hand-edited configs we
// observed on (operator host, earlier) / (operator host, earlier)) could leave behind:
//   - models.providers.{custom,anthropic,google}.baseUrl pointing at a
//     non-localhost host (e.g. http://<internal-host>:11434), bypassing the
//     proxy entirely or, worse, sending google calls into ollama
//   - models[] entries whose name ends in "<provider> / " (empty alias),
//     which OpenClaw resolves to the literal "custom/" and then falls back
//     to a non-existent default
//   - duplicate models[] entries with identical id but differing name
//   - apiKey="dummy" / authHeader=false from pre-v0.2.37 installs that
//     never had a token written, so every OpenClaw call to the proxy
//     401s with `token not registered with vault`
//
// The vault token is the only thing wall-vault knows about its own
// identity that is also what OpenClaw needs to authenticate, so the heal
// pass writes that into every provider's apiKey + sets authHeader=true.
//
// Steady-state SSE config_change (updateOpenClawJSON) keeps things in sync
// once a model change actually happens; this boot-time pass guarantees a
// stale config is corrected even when no model change ever fires.
func healOpenClawConfig(vaultToken, caBundlePath, localBaseOrigin, defaultOrigin string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return // most fleet hosts don't run OpenClaw
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	if !normalizeOpenClawProviders(cfg, vaultToken, caBundlePath, localBaseOrigin, defaultOrigin) {
		return
	}
	if err := writeJSON(path, cfg); err != nil {
		log.Printf("[openclaw-heal] write failed: %v", err)
		return
	}
	log.Printf("[openclaw-heal] normalized providers in %s", path)
}

// normalizeOpenClawProviders mutates cfg in-place to enforce the local-proxy
// invariant: provider baseUrls must point at localhost, models[] must be
// free of stale duplicates / broken-name entries, agents default model
// must not be a dangling "<provider>/" reference, and provider apiKey /
// authHeader must reflect the proxy's actual vault token (when one is
// known). Returns true when any change was applied.
func normalizeOpenClawProviders(cfg map[string]interface{}, vaultToken, caBundlePath, localBaseOrigin, defaultOrigin string) bool {
	modelsSection, ok := cfg["models"].(map[string]interface{})
	if !ok {
		return false
	}
	providers, ok := modelsSection["providers"].(map[string]interface{})
	if !ok {
		return false
	}
	customURL, anthropicURL, googleURL := providerHealURLs(localBaseOrigin, defaultOrigin)
	changed := false
	// When the plain-HTTP companion is active (localBaseOrigin set) we use
	// forceExactBaseURL so existing https://localhost: targets get rewritten
	// to the http://127.0.0.1: companion — same host, but a different
	// scheme/port that the same-host client can actually reach without
	// trusting the self-signed CA. Without the companion we keep the old
	// "leave any localhost URL alone" semantics.
	rewriteFn := forceLocalhostBaseURL
	if localBaseOrigin != "" {
		rewriteFn = forceExactBaseURL
	}
	if rewriteFn(providers, "custom", customURL) {
		changed = true
	}
	if rewriteFn(providers, "anthropic", anthropicURL) {
		changed = true
	}
	// google provider has been observed in the wild ((operator host, earlier))
	// with a stale baseUrl pointing at an upstream ollama
	// (http://<internal-host>:11434/v1) — i.e. an ollama URL accidentally
	// written into the google provider slot. Same heal rule as the
	// other providers: must be localhost; OpenClaw appends its own
	// /google/v1beta/... path so a bare proxy origin is the right shape.
	if rewriteFn(providers, "google", googleURL) {
		changed = true
	}
	if pruneStaleModelsAcross(providers) {
		changed = true
	}
	if repairDanglingPrimaryModel(cfg) {
		changed = true
	}
	if vaultToken != "" {
		if normalizeProviderAuth(providers, vaultToken) {
			changed = true
		}
	}
	if caBundlePath != "" {
		if normalizeProviderTLSCA(providers, caBundlePath) {
			changed = true
		}
	}
	if alignActiveMemoryModelToAgentDefault(cfg) {
		changed = true
	}
	if relaxChannelStaleThreshold(cfg) {
		changed = true
	}
	return changed
}

// relaxChannelStaleThreshold raises the gateway's channel-stale window
// when the operator left it unset (or set it lower than 60 minutes).
// Observed: a gateway was caught in a 300s SIGTERM-restart loop:
// OpenClaw's health monitor flagged the telegram socket as stale every
// 5 minutes when idle, and the gateway terminated for restart while a
// pending TUI/Telegram turn was mid-flight. Moving the threshold to
// 60 minutes keeps OpenClaw's failure-detection logic in place for
// genuinely-dead sockets while letting a quiet bot just sit there.
// Idempotent — only writes when the existing value is missing or
// shorter than the floor.
func relaxChannelStaleThreshold(cfg map[string]interface{}) bool {
	const floorMinutes = 60
	gw, _ := cfg["gateway"].(map[string]interface{})
	if gw == nil {
		gw = map[string]interface{}{}
	}
	current := -1
	switch v := gw["channelStaleEventThresholdMinutes"].(type) {
	case float64:
		current = int(v)
	case int:
		current = v
	}
	if current >= floorMinutes {
		return false
	}
	gw["channelStaleEventThresholdMinutes"] = floorMinutes
	cfg["gateway"] = gw
	return true
}

// alignActiveMemoryModelToAgentDefault rewrites
// plugins.entries.active-memory.config.model to match
// agents.defaults.model.primary when the two diverge.
// Triggered when active-memory shipped with a hard-coded default of
// `custom/gemini-2.5-flash-lite`, but the host has no google credentials in
// vault (services: [anthropic, ollama, lmstudio]), so every plugin tick
// failed to dispatch. The plugin's failure is then misread by the
// gateway's health-monitor as an unhealthy state, which SIGTERMs the
// gateway every 300s — and that restart loop kills any in-flight
// telegram/TUI turn. Pinning the plugin to the agent's primary model
// keeps it on a path that's actually known-good for this host (whatever
// the operator chose for their main agent), and the heal is idempotent
// so manual operator overrides survive: change `agents.defaults.model.primary`
// and the next boot resyncs.
//
// No-op when either side is empty / missing — we won't materialize a
// model field for a plugin that wasn't already configured.
func alignActiveMemoryModelToAgentDefault(cfg map[string]interface{}) bool {
	agents, ok := cfg["agents"].(map[string]interface{})
	if !ok {
		return false
	}
	defaults, ok := agents["defaults"].(map[string]interface{})
	if !ok {
		return false
	}
	mdl, ok := defaults["model"].(map[string]interface{})
	if !ok {
		return false
	}
	primary, _ := mdl["primary"].(string)
	if primary == "" {
		return false
	}
	plugins, ok := cfg["plugins"].(map[string]interface{})
	if !ok {
		return false
	}
	entries, ok := plugins["entries"].(map[string]interface{})
	if !ok {
		return false
	}
	am, ok := entries["active-memory"].(map[string]interface{})
	if !ok {
		return false
	}
	pluginCfg, ok := am["config"].(map[string]interface{})
	if !ok {
		return false
	}
	cur, _ := pluginCfg["model"].(string)
	if cur == primary {
		return false
	}
	pluginCfg["model"] = primary
	return true
}

// normalizeProviderTLSCA writes models.providers.<id>.request.tls.ca for
// every wall-vault-fronted provider so OpenClaw's embedded fetch trusts
// the proxy's self-signed cert chain. Needed because the OpenClaw daemon
// wrapper rewrites NODE_EXTRA_CA_CERTS to the system CA bundle on spawn
// (<internal incident>), so any plist-level CA hint we set is
// silently dropped before the embedded agent runs. The provider-level
// TLS CA hint, by contrast, is read directly from openclaw.json and
// flows through to undici's connect options. Idempotent — only writes
// when the existing value differs.
func normalizeProviderTLSCA(providers map[string]interface{}, caBundlePath string) bool {
	wallVaultFronted := []string{"custom", "anthropic", "google"}
	changed := false
	for _, name := range wallVaultFronted {
		p, ok := providers[name].(map[string]interface{})
		if !ok {
			continue
		}
		req, _ := p["request"].(map[string]interface{})
		if req == nil {
			req = map[string]interface{}{}
		}
		tls, _ := req["tls"].(map[string]interface{})
		if tls == nil {
			tls = map[string]interface{}{}
		}
		if cur, _ := tls["ca"].(string); cur == caBundlePath {
			continue
		}
		tls["ca"] = caBundlePath
		req["tls"] = tls
		p["request"] = req
		changed = true
	}
	return changed
}

// normalizeProviderAuth ensures every wall-vault-fronted provider carries
// the proxy's actual vault token + authHeader=true. Pre-v0.2.37 installs
// shipped with apiKey="dummy" / "proxy-managed" / "" / authHeader=false
// (or absent), and OpenClaw faithfully sent those literals as the bearer
// token, which the post-v0.2.39 token-auth gate rejects with
// `token not registered with vault`. Heal rewrites the auth fields to the
// known-good values; OpenClaw hot-reloads (or restarts) and proceeds.
//
// Only runs over providers known to be wall-vault-fronted (custom,
// anthropic, google). Third-party providers added to OpenClaw by the
// operator are left untouched on purpose — we don't know their auth
// scheme.
func normalizeProviderAuth(providers map[string]interface{}, vaultToken string) bool {
	wallVaultFronted := []string{"custom", "anthropic", "google"}
	changed := false
	for _, name := range wallVaultFronted {
		p, ok := providers[name].(map[string]interface{})
		if !ok {
			continue
		}
		if cur, _ := p["apiKey"].(string); cur != vaultToken {
			p["apiKey"] = vaultToken
			changed = true
		}
		if cur, _ := p["authHeader"].(bool); !cur {
			p["authHeader"] = true
			changed = true
		}
	}
	return changed
}

// repairDanglingPrimaryModel rewrites agents.defaults.model.primary when it
// holds a dangling "<provider>/" reference (the empty-id failure mode that
// OpenClaw rejects with `Invalid model reference: custom/`). The first
// fallback entry, if any, takes the primary slot; if none exists, the
// primary key is removed entirely so OpenClaw falls back to its own default
// resolution. Returns true when a change was applied.
func repairDanglingPrimaryModel(cfg map[string]interface{}) bool {
	agents, ok := cfg["agents"].(map[string]interface{})
	if !ok {
		return false
	}
	defaults, ok := agents["defaults"].(map[string]interface{})
	if !ok {
		return false
	}
	mdl, ok := defaults["model"].(map[string]interface{})
	if !ok {
		return false
	}
	primary, _ := mdl["primary"].(string)
	if !isDanglingModelRef(primary) {
		return false
	}
	if fbs, ok := mdl["fallbacks"].([]interface{}); ok {
		for _, f := range fbs {
			if s, ok := f.(string); ok && !isDanglingModelRef(s) && s != "" {
				mdl["primary"] = s
				return true
			}
		}
	}
	delete(mdl, "primary")
	return true
}

// isDanglingModelRef reports whether ref is a "<provider>/" string with an
// empty id half (e.g. "custom/", "anthropic/"). Empty input also counts.
func isDanglingModelRef(ref string) bool {
	if ref == "" {
		return true
	}
	idx := strings.IndexByte(ref, '/')
	if idx < 0 {
		return false // bare provider with no id half — leave alone
	}
	return strings.TrimSpace(ref[idx+1:]) == ""
}

// forceLocalhostBaseURL replaces providers[name].baseUrl when it points
// anywhere other than localhost. Returns true on change.
func forceLocalhostBaseURL(providers map[string]interface{}, name, want string) bool {
	p, ok := providers[name].(map[string]interface{})
	if !ok {
		return false
	}
	cur, _ := p["baseUrl"].(string)
	if cur == "" || isLocalProxyBaseURL(cur) {
		return false
	}
	p["baseUrl"] = want
	return true
}

// forceExactBaseURL is the strict variant of forceLocalhostBaseURL: it
// rewrites baseUrl whenever the current value differs from `want` at all,
// including local-proxy URLs that disagree on scheme or port (e.g.
// http://localhost:56244/v1 → https://localhost:56244/v1 once TLS is on).
// Used by the per-agent cache heal where a stale models.json may carry
// the right host but the wrong scheme. Empty / absent baseUrl is left
// alone so we don't materialize a config slot the operator never set.
func forceExactBaseURL(providers map[string]interface{}, name, want string) bool {
	p, ok := providers[name].(map[string]interface{})
	if !ok {
		return false
	}
	cur, _ := p["baseUrl"].(string)
	if cur == "" || cur == want {
		return false
	}
	p["baseUrl"] = want
	return true
}

// healAgentSpecificModels normalizes the per-agent provider cache at
// ~/.openclaw/agents/<id>/agent/models.json. OpenClaw maintains this file
// alongside the main openclaw.json — it carries the same providers section
// but at the top level (no models.providers wrapper) and is consulted
// during embedded-agent dispatch. On hosts upgraded from pre-v0.2.37 the
// cache typically lags the main config, so even after healOpenClawConfig
// rewrites baseUrl=https://... and apiKey=<vault token> on openclaw.json,
// the cache still holds baseUrl=http://localhost:56244/v1 + apiKey="dummy"
// + authHeader=false (observed (operator host, earlier)). Every embedded dispatch
// then either fails the TLS handshake (HTTP→HTTPS) or trips the post-
// v0.2.39 token-auth gate before reaching an LLM, and OpenClaw silently
// rotates to a fallback model that is not present locally. Heal rewrites
// the same fields the main pass touches, idempotently.
func healAgentSpecificModels(vaultToken, caBundlePath, localBaseOrigin, defaultOrigin string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	agentsDir := filepath.Join(home, ".openclaw", "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(agentsDir, e.Name(), "agent", "models.json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg map[string]interface{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		providers, ok := cfg["providers"].(map[string]interface{})
		if !ok {
			continue
		}
		customURL, anthropicURL, googleURL := providerHealURLs(localBaseOrigin, defaultOrigin)
		changed := false
		if forceExactBaseURL(providers, "custom", customURL) {
			changed = true
		}
		if forceExactBaseURL(providers, "anthropic", anthropicURL) {
			changed = true
		}
		if forceExactBaseURL(providers, "google", googleURL) {
			changed = true
		}
		if vaultToken != "" {
			if normalizeProviderAuth(providers, vaultToken) {
				changed = true
			}
		}
		if caBundlePath != "" {
			if normalizeProviderTLSCA(providers, caBundlePath) {
				changed = true
			}
		}
		if !changed {
			continue
		}
		if err := writeJSON(path, cfg); err != nil {
			log.Printf("[openclaw-heal] write %s failed: %v", path, err)
			continue
		}
		log.Printf("[openclaw-heal] normalized providers in %s", path)
	}
}

// runStartupOpenClawHeal fires healOpenClawConfig once at proxy boot in a
// goroutine so a slow disk doesn't delay startup. Mirrors runStartupSanitize.
// vaultToken is the proxy's own vault token, used by the heal pass to
// rewrite stale provider apiKey fields; pass "" to skip the auth heal.
// caBundlePath, when non-empty, is written into every wall-vault-fronted
// provider's request.tls.ca so OpenClaw trusts the proxy's self-signed
// chain even when the gateway daemon rewrites NODE_EXTRA_CA_CERTS away
// from our bundle.
// localBaseOrigin, when non-empty, points at the loopback-only plain-HTTP
// companion (e.g. http://127.0.0.1:56245). Heal rewrites every wall-vault-
// fronted provider's baseUrl to this origin so same-host clients dodge
// the self-signed-cert trust problem entirely. Empty preserves the
// existing https://localhost:56244 targets.
func runStartupOpenClawHeal(vaultToken, caBundlePath, localBaseOrigin, defaultOrigin string) {
	go func() {
		time.Sleep(750 * time.Millisecond)
		healOpenClawConfig(vaultToken, caBundlePath, localBaseOrigin, defaultOrigin)
		healAgentSpecificModels(vaultToken, caBundlePath, localBaseOrigin, defaultOrigin)
	}()
}

// providerHealURLs returns the per-provider baseUrl heal targets,
// switching to the plain-HTTP companion origin when localBaseOrigin is
// set. defaultOrigin is the operator-configured proxy origin (scheme +
// host + port) used when no companion is in play; an empty defaultOrigin
// preserves the legacy v0.2.43 fallback (`https://localhost:56244`) so
// callers that haven't been wired through to the new signature still
// produce a sensible URL.
func providerHealURLs(localBaseOrigin, defaultOrigin string) (custom, anthropic, google string) {
	origin := defaultOrigin
	if origin == "" {
		origin = "https://localhost:56244"
	}
	if localBaseOrigin != "" {
		origin = localBaseOrigin
	}
	return origin + "/v1", origin, origin
}
