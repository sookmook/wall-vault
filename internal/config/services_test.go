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

	// google.yaml 작성
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

func TestLoadPlugins_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	// 유효한 플러그인
	os.WriteFile(filepath.Join(dir, "good.yaml"), []byte("id: good\nname: Good\nenabled: true\nrequest_format: openai\n"), 0644)
	// 깨진 YAML (파싱 실패 → 건너뜀)
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("id: [broken: yaml: :::"), 0644)

	plugins, err := LoadPlugins(dir)
	if err != nil {
		t.Fatalf("깨진 YAML 있어도 오류 없어야 함: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("유효한 플러그인 1개 기대, got %d", len(plugins))
	}
}
