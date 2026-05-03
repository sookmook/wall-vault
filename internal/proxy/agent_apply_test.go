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

func TestApplyEconoWorldConfig_WritesNewRobustnessFields(t *testing.T) {
	// Fresh bootstrap into an empty workdir must surface stream,
	// request_timeout_seconds and the configured max_tokens floor.
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "analyzer"), 0755); err != nil {
		t.Fatalf("seed: %v", err)
	}
	path, err := applyEconoWorldConfig("http://127.0.0.1:56245/v1", "m1", "tok", tmp, 8192, true, 300)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["max_tokens"].(float64) != 8192 {
		t.Fatalf("max_tokens = %v, want 8192", compat["max_tokens"])
	}
	if compat["stream"] != true {
		t.Fatalf("stream = %v, want true", compat["stream"])
	}
	if compat["request_timeout_seconds"].(float64) != 300 {
		t.Fatalf("request_timeout_seconds = %v, want 300", compat["request_timeout_seconds"])
	}
}

func TestApplyEconoWorldConfig_PreservesOperatorTunedValues(t *testing.T) {
	// An operator who set custom values must not have them clobbered when
	// /agent/apply runs again — only newly-missing fields should be filled.
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "analyzer"), 0755); err != nil {
		t.Fatalf("seed: %v", err)
	}
	pre := `{
	  "provider": "openai_compatible",
	  "openai_compatible": {
	    "base_url": "http://old/v1",
	    "max_tokens": 16000,
	    "stream": false,
	    "request_timeout_seconds": 120
	  }
	}`
	path := filepath.Join(tmp, "analyzer", "ai_config.json")
	if err := os.WriteFile(path, []byte(pre), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := applyEconoWorldConfig("http://new/v1", "m2", "tok", tmp, 8192, true, 300); err != nil {
		t.Fatalf("apply: %v", err)
	}
	var got map[string]interface{}
	data, _ := os.ReadFile(path)
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["max_tokens"].(float64) != 16000 {
		t.Fatalf("max_tokens overwritten: %v", compat["max_tokens"])
	}
	if compat["stream"] != false {
		t.Fatalf("stream overwritten: %v", compat["stream"])
	}
	if compat["request_timeout_seconds"].(float64) != 120 {
		t.Fatalf("request_timeout overwritten: %v", compat["request_timeout_seconds"])
	}
	if compat["base_url"] != "http://new/v1" {
		t.Fatalf("base_url should always update: %v", compat["base_url"])
	}
}

func TestHealEconoWorldConfigAt_RewritesLocalProxyURL(t *testing.T) {
	// Same-host base_url (https://localhost:56244/v1) must be rewritten
	// to the loopback companion when localBaseOrigin is set; the path
	// suffix must survive the rewrite.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	pre := `{"provider":"openai_compatible","openai_compatible":{"base_url":"https://localhost:56244/v1"}}`
	if err := os.WriteFile(path, []byte(pre), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := healEconoWorldConfigAt(path, "http://127.0.0.1:56245", 0, false, 0); err != nil {
		t.Fatalf("heal: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["base_url"] != "http://127.0.0.1:56245/v1" {
		t.Fatalf("base_url = %v, want loopback rewrite with /v1 suffix", compat["base_url"])
	}
}

func TestHealEconoWorldConfigAt_LeavesExternalURLAlone(t *testing.T) {
	// A LAN base_url (192.168.x.x or any non-localhost host) must never
	// be rewritten — same-host loopback would be wrong for that caller.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	pre := `{"provider":"openai_compatible","openai_compatible":{"base_url":"https://<internal-host>:56244/v1"}}`
	if err := os.WriteFile(path, []byte(pre), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := healEconoWorldConfigAt(path, "http://127.0.0.1:56245", 0, false, 0); err != nil {
		t.Fatalf("heal: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["base_url"] != "https://<internal-host>:56244/v1" {
		t.Fatalf("external base_url rewritten: %v", compat["base_url"])
	}
}

func TestHealEconoWorldConfigAt_FillsMissingFields(t *testing.T) {
	// Old bootstraps lack stream / request_timeout_seconds and have the
	// legacy max_tokens=4096. Heal should fill missing fields and bump
	// the legacy 4096 floor when configured higher.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	pre := `{"provider":"openai_compatible","openai_compatible":{"base_url":"http://x","max_tokens":4096}}`
	if err := os.WriteFile(path, []byte(pre), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := healEconoWorldConfigAt(path, "", 8192, true, 300); err != nil {
		t.Fatalf("heal: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["stream"] != true {
		t.Fatalf("stream not filled: %v", compat["stream"])
	}
	if compat["request_timeout_seconds"].(float64) != 300 {
		t.Fatalf("request_timeout not filled: %v", compat["request_timeout_seconds"])
	}
	if compat["max_tokens"].(float64) != 8192 {
		t.Fatalf("legacy 4096 not bumped: %v", compat["max_tokens"])
	}
}

func TestHealEconoWorldConfigAt_PreservesNonLegacyMaxTokens(t *testing.T) {
	// max_tokens that is anything other than the legacy 4096 must be
	// left alone — operator may have tuned it deliberately to 16000 etc.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ai_config.json")
	pre := `{"provider":"openai_compatible","openai_compatible":{"base_url":"http://x","max_tokens":16000}}`
	if err := os.WriteFile(path, []byte(pre), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := healEconoWorldConfigAt(path, "", 8192, false, 0); err != nil {
		t.Fatalf("heal: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	compat, _ := got["openai_compatible"].(map[string]interface{})
	if compat["max_tokens"].(float64) != 16000 {
		t.Fatalf("operator value bumped: %v", compat["max_tokens"])
	}
}

func TestHealEconoWorldConfigAt_MissingFileIsSilent(t *testing.T) {
	if err := healEconoWorldConfigAt(filepath.Join(t.TempDir(), "nope.json"), "http://127.0.0.1:56245", 8192, true, 300); err != nil {
		t.Fatalf("missing file should be silent, got %v", err)
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

func TestEconoWorldDirCache_RoundTrip(t *testing.T) {
	// Reset between sub-cases — the cache is package-global.
	defer econoWorldDirCache.Store(nil)
	econoWorldDirCache.Store(nil)
	t.Setenv("WV_ECONOWORLD_WORKDIR_CACHE_DISABLE", "")

	if got := cachedEconoWorldDir(); got != "" {
		t.Fatalf("empty cache should return empty, got %q", got)
	}
	setCachedEconoWorldDir("/tmp/econoworld-x")
	if got := cachedEconoWorldDir(); got != "/tmp/econoworld-x" {
		t.Fatalf("cache round-trip: got %q, want %q", got, "/tmp/econoworld-x")
	}
	// Empty input must be a no-op so a probe call can't blank a real dir.
	setCachedEconoWorldDir("")
	if got := cachedEconoWorldDir(); got != "/tmp/econoworld-x" {
		t.Fatalf("empty set cleared cache: got %q", got)
	}
}

func TestEconoWorldDirCache_DisableEnvSkips(t *testing.T) {
	defer econoWorldDirCache.Store(nil)
	econoWorldDirCache.Store(nil)
	t.Setenv("WV_ECONOWORLD_WORKDIR_CACHE_DISABLE", "1")

	setCachedEconoWorldDir("/tmp/econoworld-y")
	if got := cachedEconoWorldDir(); got != "" {
		t.Fatalf("env disable: cachedEconoWorldDir should return empty, got %q", got)
	}
}

func TestApplyEconoWorldConfig_PopulatesDirCache(t *testing.T) {
	// /agent/apply must seed the cache so a subsequent SSE-driven model
	// refresh hits the same workDir the operator pinned via the dashboard,
	// not the resolveEconoWorldDir("") fallback.
	defer econoWorldDirCache.Store(nil)
	econoWorldDirCache.Store(nil)
	t.Setenv("WV_ECONOWORLD_WORKDIR_CACHE_DISABLE", "")

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "analyzer"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := applyEconoWorldConfig(
		"http://127.0.0.1:56244/v1", "google/gemma-4-26b-a4b", "tok",
		tmp, 8192, true, 300,
	); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got := cachedEconoWorldDir(); got != tmp {
		t.Fatalf("apply did not seed cache: got %q, want %q", got, tmp)
	}
}
