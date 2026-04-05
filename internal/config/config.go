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
	ClientID     string        `yaml:"client_id"`     // bot-a | mini | bot-c | free
	VaultURL     string        `yaml:"vault_url"`     // distributed mode
	VaultToken   string        `yaml:"vault_token"`
	ToolFilter   string        `yaml:"tool_filter"`   // strip_all | whitelist | passthrough
	AllowedTools []string      `yaml:"allowed_tools"` // for whitelist mode
	Services     []string      `yaml:"services"`      // active service list
	Timeout      time.Duration `yaml:"timeout"`       // API timeout
	Avatar       string        `yaml:"avatar"`        // relative path under ~/.openclaw/ (e.g. workspace/avatars/bot-a.png)
}

// ─── Key Vault Config ─────────────────────────────────────────────────────────

type VaultConfig struct {
	Port        int    `yaml:"port"`         // default 56243
	Host        string `yaml:"host"`         // default 0.0.0.0
	AdminToken  string `yaml:"admin_token"`
	MasterPass  string `yaml:"master_password"`
	DataDir     string `yaml:"data_dir"`     // default ~/.wall-vault/data
	ServicesDir string `yaml:"services_dir"` // YAML service plugin folder
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
		Mode:  "standalone",
		Lang:  "ko",
		Theme: "cherry",
		Proxy: ProxyConfig{
			Port:       56244,
			Host:       "0.0.0.0",
			ClientID:   "local",
			ToolFilter: "strip_all",
			Services:   []string{"google", "openrouter", "ollama"},
			Timeout:    60 * time.Second,
		},
		Vault: VaultConfig{
			Port:        56243,
			Host:        "0.0.0.0",
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
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("설정 파싱 오류 (%s): %w", p, err)
		}
		// override with env vars
		applyEnv(cfg)
		// load service plugins
		cfg.Plugins, _ = LoadPlugins(cfg.Vault.ServicesDir)
		return cfg, nil
	}

	// no config file found — use defaults
	applyEnv(cfg)
	cfg.Plugins, _ = LoadPlugins(cfg.Vault.ServicesDir)
	return cfg, nil
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
	// Windows: auto-set data path based on APPDATA
	if cfg.Vault.DataDir == "" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			cfg.Vault.DataDir = filepath.Join(appdata, "wall-vault", "data")
		}
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
	return os.WriteFile(path, data, 0600)
}
