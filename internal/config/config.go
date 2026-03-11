// Package config: wall-vault 설정 로드·저장
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ─── 최상위 설정 ────────────────────────────────────────────────────────────

type Config struct {
	Mode    string       `yaml:"mode"`    // standalone | distributed
	Lang    string       `yaml:"lang"`    // ko | en | ja
	Theme   string       `yaml:"theme"`   // sakura | dark | light | ocean
	Proxy   ProxyConfig  `yaml:"proxy"`
	Vault   VaultConfig  `yaml:"vault"`
	Doctor  DoctorConfig `yaml:"doctor"`
	Hooks   HooksConfig  `yaml:"hooks"`

	// 런타임 전용 — YAML 직렬화 제외 (LoadPlugins로 채워짐)
	Plugins []ServicePlugin `yaml:"-"`
}

// ─── 프록시 설정 ─────────────────────────────────────────────────────────────

type ProxyConfig struct {
	Port         int           `yaml:"port"`          // 기본 56244
	Host         string        `yaml:"host"`          // 기본 0.0.0.0
	ClientID     string        `yaml:"client_id"`     // bot-a | mini | bot-c | 자유
	VaultURL     string        `yaml:"vault_url"`     // distributed 모드
	VaultToken   string        `yaml:"vault_token"`
	ToolFilter   string        `yaml:"tool_filter"`   // strip_all | whitelist | passthrough
	AllowedTools []string      `yaml:"allowed_tools"` // whitelist 모드용
	Services     []string      `yaml:"services"`      // 활성 서비스 목록
	Timeout      time.Duration `yaml:"timeout"`       // API 타임아웃
}

// ─── 키 금고 설정 ─────────────────────────────────────────────────────────────

type VaultConfig struct {
	Port        int    `yaml:"port"`         // 기본 56243
	Host        string `yaml:"host"`         // 기본 0.0.0.0
	AdminToken  string `yaml:"admin_token"`
	MasterPass  string `yaml:"master_password"`
	DataDir     string `yaml:"data_dir"`     // 기본 ~/.wall-vault/data
	ServicesDir string `yaml:"services_dir"` // YAML 서비스 플러그인 폴더
}

// ─── 주치의 설정 ─────────────────────────────────────────────────────────────

type DoctorConfig struct {
	Interval  time.Duration `yaml:"interval"`   // 기본 5분
	AutoFix   bool          `yaml:"auto_fix"`   // 기본 true
	LogFile   string        `yaml:"log_file"`   // 기본 /tmp/wall-vault-doctor.log
}

// ─── 훅 설정 ────────────────────────────────────────────────────────────────

type HooksConfig struct {
	OnModelChange    string `yaml:"on_model_change"`    // 셸 명령
	OnKeyExhausted   string `yaml:"on_key_exhausted"`
	OnServiceDown    string `yaml:"on_service_down"`
	OnDoctorFix      string `yaml:"on_doctor_fix"`
	OpenClawSocket   string `yaml:"openclaw_socket"`    // OpenClaw TUI 소켓 경로
}

// ─── 기본값 ──────────────────────────────────────────────────────────────────

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Mode:  "standalone",
		Lang:  "ko",
		Theme: "sakura",
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

// ─── 로드 ────────────────────────────────────────────────────────────────────

// Load: 설정 파일 탐색 순서
//  1. 인수로 전달된 경로
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
		// 환경변수 덮어쓰기
		applyEnv(cfg)
		// 서비스 플러그인 로드
		cfg.Plugins, _ = LoadPlugins(cfg.Vault.ServicesDir)
		return cfg, nil
	}

	// 설정 파일 없음 — 기본값
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
	// 레거시 호환 (OpenClaw 환경변수)
	if v := os.Getenv("VAULT_URL"); v != "" && cfg.Proxy.VaultURL == "" {
		cfg.Proxy.VaultURL = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" && cfg.Proxy.VaultToken == "" {
		cfg.Proxy.VaultToken = v
	}
	if v := os.Getenv("VAULT_CLIENT_ID"); v != "" {
		cfg.Proxy.ClientID = v
	}
	// Windows: APPDATA 기반 데이터 경로 자동 설정
	if cfg.Vault.DataDir == "" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			cfg.Vault.DataDir = filepath.Join(appdata, "wall-vault", "data")
		}
	}
}

// Save: 설정 파일 저장
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
