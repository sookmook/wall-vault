// Package config: load and save wall-vault configuration
package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	ClaudeCodeClientID string `yaml:"claude_code_client_id"`
}

// ─── Key Vault Config ─────────────────────────────────────────────────────────

type VaultConfig struct {
	Port             int      `yaml:"port"`                         // default 56243
	Host             string   `yaml:"host"`                         // default 0.0.0.0
	AdminToken       string   `yaml:"admin_token"`
	AdminIPWhitelist []string `yaml:"admin_ip_whitelist,omitempty"` // IPs/CIDRs allowed to use admin token; empty = unrestricted
	MasterPass       string   `yaml:"master_password"`
	DataDir          string   `yaml:"data_dir"`     // default ~/.wall-vault/data
	ServicesDir      string   `yaml:"services_dir"` // YAML service plugin folder
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
			// 5-minute upstream timeout — local Ollama cold-starts on 27B+ models
			// (qwen3.6:27b ≈ 80s, gemma4:26b ≈ 6m) blow past anything shorter,
			// causing every minute-cron caller to disconnect mid-load and trigger
			// the cold-start loop seen on mini.
			Timeout:    300 * time.Second,
		},
		Vault: VaultConfig{
			Port:        56243,
			Host:        "",
			DataDir:     filepath.Join(home, ".wall-vault", "data"),
			ServicesDir: filepath.Join(home, ".wall-vault", "services"),
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
