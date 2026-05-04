package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPlugins_Empty(t *testing.T) {
	plugins, err := LoadPlugins("")
	if err != nil {
		t.Fatalf("빈 경로: 오류 기대 안 함, got %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("빈 경로: 0개 기대, got %d", len(plugins))
	}
}

func TestLoadPlugins_NotExist(t *testing.T) {
	plugins, err := LoadPlugins("/nonexistent/path/xyz")
	if err != nil {
		t.Fatalf("없는 경로: 오류 기대 안 함 (nil 반환), got %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("0개 기대, got %d", len(plugins))
	}
}

func TestLoadPlugins_Valid(t *testing.T) {
	dir := t.TempDir()

	// write google.yaml
	googleYAML := `
id: google
name: Google Gemini
enabled: true
request_format: gemini
endpoints:
  generate: https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
auth:
  type: query_param
  param: key
`
	// disabled.yaml (enabled: false)
	disabledYAML := `
id: disabled-svc
name: Disabled Service
enabled: false
request_format: openai
`

	os.WriteFile(filepath.Join(dir, "google.yaml"), []byte(googleYAML), 0644)
	os.WriteFile(filepath.Join(dir, "disabled.yaml"), []byte(disabledYAML), 0644)
	os.WriteFile(filepath.Join(dir, "notayaml.txt"), []byte("ignored"), 0644)

	plugins, err := LoadPlugins(dir)
	if err != nil {
		t.Fatalf("LoadPlugins 오류: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("enabled 플러그인 1개 기대, got %d", len(plugins))
	}
	if plugins[0].ID != "google" {
		t.Fatalf("플러그인 ID 불일치: %q", plugins[0].ID)
	}
	if plugins[0].RequestFormat != "gemini" {
		t.Fatalf("request_format 불일치: %q", plugins[0].RequestFormat)
	}
}

func TestPluginByID(t *testing.T) {
	plugins := []ServicePlugin{
		{ID: "google", Name: "Google"},
		{ID: "openrouter", Name: "OpenRouter"},
	}

	p := PluginByID(plugins, "openrouter")
	if p == nil {
		t.Fatal("openrouter 플러그인 못 찾음")
	}
	if p.Name != "OpenRouter" {
		t.Fatalf("이름 불일치: %q", p.Name)
	}

	notFound := PluginByID(plugins, "nonexistent")
	if notFound != nil {
		t.Fatal("없는 플러그인이 반환됨")
	}
}

func TestLoadPlugins_NewSchemaFields(t *testing.T) {
	dir := t.TempDir()
	hubYAML := `
id: wall-vault-hub
name: Hub Wall-Vault
enabled: true
request_format: openai
default_url: https://hub.example:56244
default_model: qwen3.6:27b
tls_internal_ca: true
preserve_model_id: true
auth:
  type: bearer
endpoints:
  generate: /v1/chat/completions
`
	if err := os.WriteFile(filepath.Join(dir, "hub.yaml"), []byte(hubYAML), 0644); err != nil {
		t.Fatal(err)
	}
	plugins, err := LoadPlugins(dir)
	if err != nil {
		t.Fatalf("LoadPlugins error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	p := plugins[0]
	if p.ID != "wall-vault-hub" {
		t.Errorf("id = %q", p.ID)
	}
	if !p.TLSInternalCA {
		t.Errorf("expected tls_internal_ca=true, got false")
	}
	if p.Auth.Type != "bearer" {
		t.Errorf("auth.type = %q, want bearer", p.Auth.Type)
	}
	if p.DefaultURL != "https://hub.example:56244" {
		t.Errorf("default_url = %q", p.DefaultURL)
	}
	if p.DefaultModel != "qwen3.6:27b" {
		t.Errorf("default_model = %q", p.DefaultModel)
	}
	if !p.PreserveModelID {
		t.Errorf("expected preserve_model_id=true, got false")
	}
}

func TestLoadPlugins_BackwardCompat(t *testing.T) {
	// A plugin yaml without any of the new fields must still load and
	// behave the same as before — implicit Compat=openai_chat_completions,
	// Auth.Type=none, TLSInternalCA=false.
	dir := t.TempDir()
	legacyYAML := `
id: legacy
name: Legacy Service
enabled: true
request_format: openai
endpoints:
  generate: http://localhost:9999/v1/chat/completions
`
	os.WriteFile(filepath.Join(dir, "legacy.yaml"), []byte(legacyYAML), 0644)
	plugins, err := LoadPlugins(dir)
	if err != nil {
		t.Fatalf("LoadPlugins error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	p := plugins[0]
	if p.TLSInternalCA {
		t.Errorf("legacy plugin should default tls_internal_ca=false")
	}
	if p.Auth.Type != "" {
		t.Errorf("legacy auth.type should be empty (=none), got %q", p.Auth.Type)
	}
	if p.DefaultURL != "" {
		t.Errorf("legacy default_url should be empty, got %q", p.DefaultURL)
	}
}

func TestLoadPlugins_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	// valid plugin
	os.WriteFile(filepath.Join(dir, "good.yaml"), []byte("id: good\nname: Good\nenabled: true\nrequest_format: openai\n"), 0644)
	// broken YAML (parse failure → skip)
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("id: [broken: yaml: :::"), 0644)

	plugins, err := LoadPlugins(dir)
	if err != nil {
		t.Fatalf("깨진 YAML 있어도 오류 없어야 함: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("유효한 플러그인 1개 기대, got %d", len(plugins))
	}
}

func TestApplyEnv_OAIStreamForward(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"yes", true},
		{"on", true},
		{" true ", true},    // whitespace-padded
		{"0", false},
		{"false", false},
		{"no", false},
		{"off", false},
		{"", false},      // unset
		{"ture", false},  // typo — unrecognised, leaves default
	}
	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv("WV_OAI_STREAM_FORWARD", tc.env)
			cfg := Default()
			applyEnv(cfg)
			if cfg.Proxy.OAIStreamForward != tc.want {
				t.Errorf("env=%q want=%v got=%v", tc.env, tc.want, cfg.Proxy.OAIStreamForward)
			}
		})
	}
}
