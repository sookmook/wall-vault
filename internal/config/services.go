// Package config: 서비스 플러그인 로더
package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ─── 서비스 플러그인 구조 ─────────────────────────────────────────────────────

// ServicePlugin: configs/services/*.yaml 형식
type ServicePlugin struct {
	ID            string                 `yaml:"id"`
	Name          string                 `yaml:"name"`
	Enabled       bool                   `yaml:"enabled"`
	Endpoints     ServiceEndpoints       `yaml:"endpoints"`
	Auth          ServiceAuth            `yaml:"auth"`
	RequestFormat string                 `yaml:"request_format"` // gemini | openai | ollama | raw
	ModelFetch    ServiceModelFetch      `yaml:"model_fetch"`
	ErrorCodes    map[int]ErrorCodeRule  `yaml:"error_codes"`
	UsageThreshold int                   `yaml:"usage_threshold"`
	Headers       ServiceHeaders         `yaml:"headers"`
	Concurrency   ServiceConcurrency     `yaml:"concurrency"`
	FallbackOnly  bool                   `yaml:"fallback_only"`
}

type ServiceEndpoints struct {
	Generate   string `yaml:"generate"`
	ListModels string `yaml:"list_models"`
}

type ServiceAuth struct {
	Type  string `yaml:"type"`  // bearer | query_param | header | none
	Param string `yaml:"param"` // query_param 이름 (e.g. "key")
}

type ServiceModelFetch struct {
	Enabled        bool     `yaml:"enabled"`
	URL            string   `yaml:"url"`
	AuthRequired   bool     `yaml:"auth_required"`
	Dynamic        bool     `yaml:"dynamic"`
	AutoDetectURL  bool     `yaml:"auto_detect_url"`
	FallbackModels []FallbackModel `yaml:"fallback_models"`
}

type FallbackModel struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type ErrorCodeRule struct {
	Cooldown time.Duration `yaml:"cooldown"`
	Message  string        `yaml:"message"`
}

type ServiceHeaders struct {
	Strip []string          `yaml:"strip"`
	Add   map[string]string `yaml:"add"`
}

type ServiceConcurrency struct {
	Max         int  `yaml:"max"`
	QueueSize   int  `yaml:"queue_size"`
	WaitNotify  bool `yaml:"wait_notify"`
}

// ─── 로더 ─────────────────────────────────────────────────────────────────────

// LoadPlugins: ServicesDir에서 *.yaml 로드, enabled=true인 것만 반환
func LoadPlugins(servicesDir string) ([]ServicePlugin, error) {
	if servicesDir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []ServicePlugin
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(servicesDir, e.Name()))
		if err != nil {
			continue
		}
		var p ServicePlugin
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}
		if p.Enabled && p.ID != "" {
			plugins = append(plugins, p)
		}
	}
	return plugins, nil
}

// PluginByID: ID로 플러그인 검색
func PluginByID(plugins []ServicePlugin, id string) *ServicePlugin {
	for i := range plugins {
		if plugins[i].ID == id {
			return &plugins[i]
		}
	}
	return nil
}
