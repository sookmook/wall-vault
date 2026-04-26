// Package config: service plugin loader
package config

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ─── Service Plugin Structure ─────────────────────────────────────────────────

// ServicePlugin: configs/services/*.yaml format
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
	Param string `yaml:"param"` // query_param name (e.g. "key")
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

// ─── Loader ───────────────────────────────────────────────────────────────────

// LoadPlugins: load *.yaml from ServicesDir, return only those with enabled=true
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
		path := filepath.Join(servicesDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[plugins] skip %s: read failed: %v", e.Name(), err)
			continue
		}
		var p ServicePlugin
		if err := yaml.Unmarshal(data, &p); err != nil {
			log.Printf("[plugins] skip %s: yaml parse failed: %v", e.Name(), err)
			continue
		}
		if p.ID == "" {
			log.Printf("[plugins] skip %s: missing id field", e.Name())
			continue
		}
		if !p.Enabled {
			continue
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

// PluginByID: find plugin by ID
func PluginByID(plugins []ServicePlugin, id string) *ServicePlugin {
	for i := range plugins {
		if plugins[i].ID == id {
			return &plugins[i]
		}
	}
	return nil
}
