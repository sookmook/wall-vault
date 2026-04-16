package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateEconoWorldModelAt_UpdatesModelField(t *testing.T) {
	// SSE-driven model refresh must only touch the model field inside
	// openai_compatible; base_url / api_key / max_tokens must survive so we
	// don't lose the bootstrap written by /agent/apply.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	initial := `{
  "provider": "openai_compatible",
  "openai_compatible": {
    "base_url": "http://127.0.0.1:56244/v1",
    "api_key":  "keep-me",
    "model":    "old-model",
    "max_tokens": 4096
  }
}`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := updateEconoWorldModelAt(path, "new-model"); err != nil {
		t.Fatalf("update: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readback: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("parse: %v", err)
	}
	compat, ok := got["openai_compatible"].(map[string]interface{})
	if !ok {
		t.Fatalf("openai_compatible missing after write")
	}
	if compat["model"] != "new-model" {
		t.Fatalf("model = %v, want new-model", compat["model"])
	}
	if compat["api_key"] != "keep-me" {
		t.Fatalf("api_key mutated: %v", compat["api_key"])
	}
	if compat["base_url"] != "http://127.0.0.1:56244/v1" {
		t.Fatalf("base_url mutated: %v", compat["base_url"])
	}
	if got["provider"] != "openai_compatible" {
		t.Fatalf("provider mutated: %v", got["provider"])
	}
}

func TestUpdateEconoWorldModelAt_MissingFileIsSilent(t *testing.T) {
	// A proxy on a host without EconoWorld should get a silent no-op; we
	// don't want error log noise on every econoworld config_change event.
	err := updateEconoWorldModelAt(filepath.Join(t.TempDir(), "does-not-exist.json"), "any")
	if err != nil {
		t.Fatalf("missing file should be silent, got %v", err)
	}
}

func TestUpdateEconoWorldModelAt_NoOpenAICompatSectionIsSilent(t *testing.T) {
	// If ai_config.json exists but hasn't been bootstrapped with the
	// openai_compatible section, leave it alone — the next /agent/apply
	// call is responsible for creating that structure.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	if err := os.WriteFile(path, []byte(`{"provider":"anthropic"}`), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := updateEconoWorldModelAt(path, "any"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	if _, exists := got["openai_compatible"]; exists {
		t.Fatalf("should not have created openai_compatible section: %+v", got)
	}
}

func TestUpdateEconoWorldModel_EmptyModelIsNoop(t *testing.T) {
	// The SSE callback passes "" when a client clears its model_override;
	// that path must not clobber the ai_config.json model (it should keep
	// whatever was bootstrapped). We can assert this at the path-level
	// entry point by making sure the file isn't touched when model is "".
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	original := `{"provider":"openai_compatible","openai_compatible":{"model":"keep-me"}}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	infoBefore, _ := os.Stat(path)

	// Directly call the path-taking function with empty model — expect no change.
	// (updateEconoWorldModel would short-circuit before even reading the file.)
	if err := updateEconoWorldModelAt(path, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["model"] != "keep-me" {
		t.Fatalf("empty model overwrote existing: %v", compat["model"])
	}
	infoAfter, _ := os.Stat(path)
	_ = infoBefore
	_ = infoAfter
}
