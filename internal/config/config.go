// Package config: load and save wall-vault configuration
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ─── Top-Level Config ────────────────────────────────────────────────────────

type Config struct {
	Mode    string       `yaml:"mode"`    // standalone | distributed
	Lang    string       `yaml:"lang"`    // ko | en | ja | ...
	Theme   string       `yaml:"theme"`   // dark | light | cherry | ocean | ...
	Proxy   ProxyConfig  `yaml:"proxy"`
	Vault   VaultConfig  `yaml:"vault"`
	Doctor  DoctorConfig `yaml:"doctor"`
	Hooks   HooksConfig  `yaml:"hooks"`

	// runtime-only — excluded from YAML serialization (populated by LoadPlugins)
	Plugins []ServicePlugin `yaml:"-"`
}

// ─── Proxy Config ─────────────────────────────────────────────────────────────

type ProxyConfig struct {
	Port         int           `yaml:"port"`          // default 56244
	Host         string        `yaml:"host"`          // default 0.0.0.0
	ClientID     string        `yaml:"client_id"`     // stable identifier for this proxy in vault (free-form slug)
	VaultURL     string        `yaml:"vault_url"`     // distributed mode
	VaultToken   string        `yaml:"vault_token"`
	ToolFilter   string        `yaml:"tool_filter"`   // strip_all | whitelist | passthrough
	AllowedTools []string      `yaml:"allowed_tools"` // for whitelist mode
	Services     []string      `yaml:"services"`      // active service list
	Timeout      time.Duration `yaml:"timeout"`       // API timeout
	Avatar       string        `yaml:"avatar"`        // relative path under ~/.openclaw/ (e.g. workspace/avatars/<client-id>.png)
	// ClaudeCodeClientID overrides automatic claude-code client selection. When
	// empty, the proxy auto-picks the claude-code client whose Host matches
	// os.Hostname(). When set, this value wins — lets operators pin the mapping
	// on hosts where hostname detection is unreliable (WSL, renamed boxes).
	ClaudeCodeClientID string    `yaml:"claude_code_client_id"`
	TLS                TLSConfig `yaml:"tls"`
	// PlainPort runs an additional plain-HTTP listener bound to loopback
	// only (127.0.0.1) when TLS is enabled, so same-host clients that
	// cannot be coerced into trusting the proxy's self-signed CA can still
	// reach it. Triggered by (operator host, earlier): OpenClaw's macOS daemon
	// rewrites NODE_EXTRA_CA_CERTS to /etc/ssl/cert.pem at spawn,
	// dropping any operator-provided CA hint, so every embedded fetch to
	// the TLS listener fails handshake. The plain listener bypasses the
	// trust problem entirely while keeping the TLS listener for LAN
	// callers (other fleet machines using ca.crt). 0 disables. Default
	// 56245. Ignored when TLS.Enabled is false (the main listener is
	// already plain HTTP in that case).
	PlainPort int `yaml:"plain_port"`
	// OllamaKeepAlive controls how long Ollama keeps the model loaded after a
	// response. "30m" means thirty minutes idle before unload; "-1" never
	// unloads; "0" unloads immediately. Default Ollama behaviour is 5 minutes,
	// which causes 80-100s cold reloads on the 27B fleet model when calls are
	// sparse. Empty string leaves Ollama on its own default.
	OllamaKeepAlive string `yaml:"ollama_keep_alive"`
	// OllamaNumCtx pins the Ollama context window. Default Ollama (2048) is
	// too small for long Korean conversations; 8192 is a reasonable starting
	// point. Zero = leave Ollama default in place.
	OllamaNumCtx int `yaml:"ollama_num_ctx"`
	// EconoWorldMaxTokens is the default max_tokens written into a freshly
	// bootstrapped EconoWorld ai_config.json. The previous hard-coded 4096
	// truncated long Korean analyses mid-sentence. Existing values in the
	// file are preserved — this only fills in when the field is missing.
	EconoWorldMaxTokens int `yaml:"econoworld_max_tokens"`
	// EconoWorldStream toggles `stream:true` in EconoWorld's ai_config.json.
	// True is safer when EconoWorld's openai_compatible client supports
	// streaming — partial output keeps appearing instead of waiting for the
	// whole response, hiding ollama prompt-eval latency. Existing values
	// are preserved.
	EconoWorldStream bool `yaml:"econoworld_stream"`
	// EconoWorldRequestTimeout (seconds) is the default request_timeout_seconds
	// written into EconoWorld's ai_config.json. EconoWorld may ignore the
	// field; presence is harmless. Zero = field omitted.
	EconoWorldRequestTimeout int `yaml:"econoworld_request_timeout"`

	// TokenSentinelFallback, when true, lets a loopback caller present a
	// well-known sentinel ("proxy-managed", "dummy", "") in its Bearer
	// header and have the proxy substitute its own VaultToken. This rescues
	// agents (notably OpenClaw on a host whose heal pass has not finalised
	// the token rewrite yet) from a hard 401 dead-end without giving up
	// vault-side authn for non-loopback callers. Default off — operators
	// opt in per host with WV_TOKEN_SENTINEL_FALLBACK=1.
	TokenSentinelFallback bool `yaml:"token_sentinel_fallback"`
	// OAIStreamForward toggles real backend stream passthrough for
	// oaiCompatServices clients. When off (default) the OpenAI-compatible
	// handler keeps its v0.2.61 fake-chunk replay behaviour. When on, a
	// caller's stream:true triggers a streamLocalService call that pipes
	// backend SSE chunks through with no aggregate buffering. Stream-mode
	// callers do NOT consult the fallback chain — see
	// docs/superpowers/specs/2026-05-04-oai-stream-passthrough-design.md §3.3.
	OAIStreamForward bool `yaml:"oai_stream_forward"`

	// AnthropicFallbackModel: when an anthropic dispatch arrives with a
	// non-Claude model id, the proxy uses this id instead. Empty (default)
	// makes dispatch return an error so the misrouting surfaces immediately.
	// Pre-v0.2.62 wall-vault silently rewrote any non-Claude id to
	// claude-haiku-4-5-20251001, which hid bugs (a fleet host that lost
	// vault keys would invisibly burn Anthropic credits on a model the
	// operator never asked for). Operators who relied on the historic
	// behaviour can opt back in by setting this to "claude-haiku-4-5-20251001"
	// or whichever Claude id their account is provisioned for.
	AnthropicFallbackModel string `yaml:"anthropic_fallback_model"`
}

// ─── Key Vault Config ─────────────────────────────────────────────────────────

type VaultConfig struct {
	Port             int       `yaml:"port"`                         // default 56243
	Host             string    `yaml:"host"`                         // default 0.0.0.0
	AdminToken       string    `yaml:"admin_token"`
	AdminIPWhitelist []string  `yaml:"admin_ip_whitelist,omitempty"` // IPs/CIDRs allowed to use admin token; empty = unrestricted
	MasterPass       string    `yaml:"master_password"`
	DataDir          string    `yaml:"data_dir"`     // default ~/.wall-vault/data
	ServicesDir      string    `yaml:"services_dir"` // YAML service plugin folder
	TLS              TLSConfig `yaml:"tls"`
	// BootstrapPort runs a tiny plain-HTTP listener that serves only ca.crt
	// + an install help page. This breaks the catch-22 where new clients
	// need the CA to speak HTTPS to vault but the CA itself is only on
	// HTTPS. CA is public-info; the listener exposes nothing else. Set to
	// 0 to disable. Default 56247.
	BootstrapPort int `yaml:"bootstrap_port"`
}

// TLSConfig holds the per-listener TLS settings. When Enabled is false the
// listener falls back to plain HTTP (current default). Both CertFile and
// KeyFile must point at PEM files when Enabled is true; the loader does not
// auto-generate certs — use `wall-vault cert init` + `wall-vault cert issue
// <hostname>` to provision them under ~/.wall-vault.
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"` // path to PEM-encoded cert
	KeyFile  string `yaml:"key_file"`  // path to PEM-encoded private key
}

// ─── Doctor Config ────────────────────────────────────────────────────────────

type DoctorConfig struct {
	Interval  time.Duration `yaml:"interval"`   // default 5 minutes
	AutoFix   bool          `yaml:"auto_fix"`   // default true
	LogFile   string        `yaml:"log_file"`   // default /tmp/wall-vault-doctor.log
}

// ─── Hooks Config ─────────────────────────────────────────────────────────────

type HooksConfig struct {
	OnModelChange    string `yaml:"on_model_change"`    // shell command
	OnKeyExhausted   string `yaml:"on_key_exhausted"`
	OnServiceDown    string `yaml:"on_service_down"`
	OnDoctorFix      string `yaml:"on_doctor_fix"`
	OpenClawSocket   string `yaml:"openclaw_socket"`    // OpenClaw TUI socket path
}

// ─── Defaults ────────────────────────────────────────────────────────────────

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Mode: "standalone",
		Lang: "ko",
		// Default to "light" so fresh installs pick the most accessible theme.
		// Existing installs keep whatever they had — vault.json's Settings.Theme
		// overrides this, and handleAdminTheme persists user choice.
		Theme: "light",
		Proxy: ProxyConfig{
			Port: 56244,
			// Host intentionally left empty — see applyHostDefaults: standalone
			// binds to 127.0.0.1, distributed to 0.0.0.0.
			Host:       "",
			ClientID:   "local",
			ToolFilter: "strip_all",
			Services:   []string{"google", "openrouter", "ollama"},
			// 10-minute upstream timeout — the local-call budget. 27B
			// reasoning-class models on Apple Silicon can take 80-113s for
			// cold-load alone; the reasoning pass on a long-form turn can
			// push total wall-clock past 4 minutes. Anything shorter made
			// the bot's main lane fire context-cancel mid-stream, drop the
			// outbound chat reply, and re-arm the health-monitor restart
			// loop. 600s is the bot-side deadline operators should align
			// with.
			Timeout: 600 * time.Second,
			// Keep the 27B fleet model warm for 30 minutes after each call so
			// sparse traffic doesn't pay the 80-113s cold-reload tax on every
			// request. Operators with tight RAM can override via WV_OLLAMA_KEEP_ALIVE.
			OllamaKeepAlive: "30m",
			// 8K context covers long Korean conversations + tool-call payloads
			// without spilling into the 27B model's slow path. Override via
			// WV_OLLAMA_NUM_CTX or per-host config.
			OllamaNumCtx: 8192,
			// EconoWorld defaults — see ProxyConfig docs above. The previous
			// hard-coded 4096 max_tokens cut Korean analyses mid-output; 8192
			// is the new floor. Streaming is on by default so cold-load latency
			// surfaces as gradually arriving tokens rather than a long silence
			// followed by a wall of text. 600s timeout matches the local-call
			// budget shared with Proxy.Timeout.
			EconoWorldMaxTokens:      8192,
			EconoWorldStream:         true,
			EconoWorldRequestTimeout: 600,
			// Loopback-only plain HTTP companion. Disabled (0) when TLS is
			// off; activated when TLS is on so same-host clients that
			// cannot honour our CA still have a path in.
			PlainPort: 56245,
		},
		Vault: VaultConfig{
			Port:          56243,
			Host:          "",
			DataDir:       filepath.Join(home, ".wall-vault", "data"),
			ServicesDir:   filepath.Join(home, ".wall-vault", "services"),
			BootstrapPort: 56247,
		},
		Doctor: DoctorConfig{
			Interval: 5 * time.Minute,
			AutoFix:  true,
			LogFile:  "/tmp/wall-vault-doctor.log",
		},
	}
}

// ─── Load ─────────────────────────────────────────────────────────────────────

// Load: config file search order
//  1. path passed as argument
//  2. ./wall-vault.yaml
//  3. ~/.wall-vault/config.yaml
func Load(path string) (*Config, error) {
	cfg := Default()

	candidates := []string{}
	if path != "" {
		candidates = append(candidates, path)
	}
	candidates = append(candidates, "wall-vault.yaml")
	home, _ := os.UserHomeDir()
	candidates = append(candidates, filepath.Join(home, ".wall-vault", "config.yaml"))

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		checkConfigPermission(p)
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("설정 파싱 오류 (%s): %w", p, err)
		}
		// override with env vars
		applyEnv(cfg)
		applyHostDefaults(cfg)
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("설정 검증 실패 (%s): %w", p, err)
		}
		// load service plugins
		cfg.Plugins, _ = LoadPlugins(cfg.Vault.ServicesDir)
		return cfg, nil
	}

	// no config file found — use defaults
	applyEnv(cfg)
	applyHostDefaults(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("설정 검증 실패 (defaults): %w", err)
	}
	cfg.Plugins, _ = LoadPlugins(cfg.Vault.ServicesDir)
	return cfg, nil
}

// Validate checks required fields / value ranges after Load() has merged YAML +
// env + defaults. Returns on the first violation so the caller sees a clear
// reason why the binary refused to start.
func (c *Config) Validate() error {
	if c.Mode != "" && c.Mode != "standalone" && c.Mode != "distributed" {
		return fmt.Errorf("mode must be 'standalone' or 'distributed', got %q", c.Mode)
	}
	if c.Proxy.Port <= 0 || c.Proxy.Port > 65535 {
		return fmt.Errorf("proxy.port out of range: %d", c.Proxy.Port)
	}
	if c.Vault.Port <= 0 || c.Vault.Port > 65535 {
		return fmt.Errorf("vault.port out of range: %d", c.Vault.Port)
	}
	if c.Proxy.Timeout <= 0 {
		return fmt.Errorf("proxy.timeout must be positive, got %v", c.Proxy.Timeout)
	}
	switch c.Proxy.ToolFilter {
	case "", "strip_all", "whitelist", "passthrough":
	default:
		return fmt.Errorf("proxy.tool_filter must be one of strip_all|whitelist|passthrough, got %q", c.Proxy.ToolFilter)
	}
	if len(c.Proxy.Services) == 0 {
		return fmt.Errorf("proxy.services is empty — at least one service must be listed")
	}
	return nil
}

// applyHostDefaults fills empty Host fields with a mode-appropriate default.
// Standalone binds to loopback only; distributed must accept remote proxies.
// If the user explicitly sets Host in YAML or via WV_*_HOST, that value wins.
func applyHostDefaults(cfg *Config) {
	defaultHost := "127.0.0.1"
	if cfg.Mode == "distributed" {
		defaultHost = "0.0.0.0"
	}
	if cfg.Proxy.Host == "" {
		cfg.Proxy.Host = defaultHost
	}
	if cfg.Vault.Host == "" {
		cfg.Vault.Host = defaultHost
	}
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("WV_LANG"); v != "" {
		cfg.Lang = v
	}
	if v := os.Getenv("WV_THEME"); v != "" {
		cfg.Theme = v
	}
	if v := os.Getenv("WV_VAULT_URL"); v != "" {
		cfg.Proxy.VaultURL = v
	}
	if v := os.Getenv("WV_VAULT_TOKEN"); v != "" {
		cfg.Proxy.VaultToken = v
	}
	if v := os.Getenv("WV_ADMIN_TOKEN"); v != "" {
		cfg.Vault.AdminToken = v
	}
	if v := os.Getenv("WV_MASTER_PASS"); v != "" {
		cfg.Vault.MasterPass = v
	}
	if v := os.Getenv("WV_PROXY_PORT"); v != "" {
		var p int
		fmt.Sscanf(v, "%d", &p)
		if p > 0 {
			cfg.Proxy.Port = p
		}
	}
	if v := os.Getenv("WV_VAULT_PORT"); v != "" {
		var p int
		fmt.Sscanf(v, "%d", &p)
		if p > 0 {
			cfg.Vault.Port = p
		}
	}
	// legacy compatibility (OpenClaw env vars)
	if v := os.Getenv("VAULT_URL"); v != "" && cfg.Proxy.VaultURL == "" {
		cfg.Proxy.VaultURL = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" && cfg.Proxy.VaultToken == "" {
		cfg.Proxy.VaultToken = v
	}
	if v := os.Getenv("VAULT_CLIENT_ID"); v != "" {
		cfg.Proxy.ClientID = v
	}
	if v := os.Getenv("WV_AVATAR"); v != "" {
		cfg.Proxy.Avatar = v
	}
	if v := os.Getenv("WV_TOOL_FILTER"); v != "" {
		cfg.Proxy.ToolFilter = v
	}
	if v := os.Getenv("WV_CC_CLIENT_ID"); v != "" {
		cfg.Proxy.ClaudeCodeClientID = v
	}
	if v := os.Getenv("WV_PROXY_HOST"); v != "" {
		cfg.Proxy.Host = v
	}
	if v := os.Getenv("WV_VAULT_HOST"); v != "" {
		cfg.Vault.Host = v
	}
	// TLS (proxy)
	if v := os.Getenv("WV_PROXY_TLS_ENABLED"); v == "1" || v == "true" {
		cfg.Proxy.TLS.Enabled = true
	}
	if v := os.Getenv("WV_PROXY_TLS_CERT"); v != "" {
		cfg.Proxy.TLS.CertFile = v
	}
	if v := os.Getenv("WV_PROXY_TLS_KEY"); v != "" {
		cfg.Proxy.TLS.KeyFile = v
	}
	// TLS (vault)
	if v := os.Getenv("WV_VAULT_TLS_ENABLED"); v == "1" || v == "true" {
		cfg.Vault.TLS.Enabled = true
	}
	if v := os.Getenv("WV_VAULT_TLS_CERT"); v != "" {
		cfg.Vault.TLS.CertFile = v
	}
	if v := os.Getenv("WV_VAULT_TLS_KEY"); v != "" {
		cfg.Vault.TLS.KeyFile = v
	}
	// Bootstrap (plain-HTTP CA distribution) port. Set to 0 to disable.
	if v := os.Getenv("WV_VAULT_BOOTSTRAP_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 65535 {
			cfg.Vault.BootstrapPort = n
		}
	}
	// Loopback-only plain HTTP companion port for the proxy (see
	// ProxyConfig.PlainPort docs). Set to 0 to disable.
	if v := os.Getenv("WV_PROXY_PLAIN_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 65535 {
			cfg.Proxy.PlainPort = n
		}
	}
	// Ollama tuning — env vars win so operators can hot-tune without rewriting
	// YAML. Empty/zero values fall back to whatever the YAML or Default()
	// already set.
	if v := os.Getenv("WV_OLLAMA_KEEP_ALIVE"); v != "" {
		cfg.Proxy.OllamaKeepAlive = v
	}
	if v := os.Getenv("WV_OLLAMA_NUM_CTX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Proxy.OllamaNumCtx = n
		}
	}
	// EconoWorld ai_config.json defaults — env wins over YAML so an operator
	// can hot-tune without redeploy. See ProxyConfig docs for semantics.
	if v := os.Getenv("WV_ECONOWORLD_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Proxy.EconoWorldMaxTokens = n
		}
	}
	if v := os.Getenv("WV_ECONOWORLD_STREAM"); v != "" {
		// Accept the usual truthy / falsy spellings. Anything unrecognised
		// leaves the existing value alone — silently ignoring typos beats
		// flipping the flag the operator did not intend.
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			cfg.Proxy.EconoWorldStream = true
		case "0", "false", "no", "off":
			cfg.Proxy.EconoWorldStream = false
		}
	}
	if v := os.Getenv("WV_ECONOWORLD_REQUEST_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.Proxy.EconoWorldRequestTimeout = n
		}
	}
	if v := os.Getenv("WV_TOKEN_SENTINEL_FALLBACK"); v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			cfg.Proxy.TokenSentinelFallback = true
		case "0", "false", "no", "off":
			cfg.Proxy.TokenSentinelFallback = false
		}
	}
	if v := os.Getenv("WV_ANTHROPIC_FALLBACK_MODEL"); v != "" {
		cfg.Proxy.AnthropicFallbackModel = strings.TrimSpace(v)
	}
	if v := os.Getenv("WV_OAI_STREAM_FORWARD"); v != "" {
		// Accept the usual truthy / falsy spellings. Anything unrecognised
		// leaves the existing value alone — silently ignoring typos beats
		// flipping the flag the operator did not intend.
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			cfg.Proxy.OAIStreamForward = true
		case "0", "false", "no", "off":
			cfg.Proxy.OAIStreamForward = false
		}
	}
	// Windows: auto-set data path based on APPDATA
	if cfg.Vault.DataDir == "" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			cfg.Vault.DataDir = filepath.Join(appdata, "wall-vault", "data")
		}
	}
}

// checkConfigPermission logs a warning when the config file is world/group
// readable. Config files contain secrets (admin_token, master_password), so a
// 0644 file on a multi-user machine leaks them. We don't hard-fail — some
// users intentionally run on a single-user box and don't want a noisy refusal
// to boot — but the warning makes the risk visible in `journalctl`.
func checkConfigPermission(path string) {
	st, err := os.Stat(path)
	if err != nil {
		return
	}
	mode := st.Mode().Perm()
	if mode&0077 != 0 {
		fmt.Fprintf(os.Stderr, "[config] warning: %s has permissions %#o — secrets readable by group/other. Run: chmod 600 %s\n", path, mode, path)
	}
}

// Save: save config file
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	// Enforce 0600 even if the file already existed with looser permissions.
	// os.WriteFile preserves existing mode for existing files; Chmod is the
	// explicit belt-and-suspenders that guarantees secrets aren't world-
	// readable after a save.
	if err := os.Chmod(path, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "[config] warning: chmod 0600 failed on %s: %v\n", path, err)
	}
	return nil
}
