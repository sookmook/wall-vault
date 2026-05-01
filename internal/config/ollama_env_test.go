package config

import "testing"

// TestOllamaTuningDefaults — Default() bakes in the keep_alive=30m / num_ctx=8192
// pair. Operators who never touch the config or env still get the warm-cache
// behaviour out of the box.
func TestOllamaTuningDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Proxy.OllamaKeepAlive != "30m" {
		t.Errorf("OllamaKeepAlive default = %q, want 30m", cfg.Proxy.OllamaKeepAlive)
	}
	if cfg.Proxy.OllamaNumCtx != 8192 {
		t.Errorf("OllamaNumCtx default = %d, want 8192", cfg.Proxy.OllamaNumCtx)
	}
}

// TestOllamaTuningEnvOverride — env vars must win over Default(), letting
// operators tune RAM/latency without rewriting YAML.
func TestOllamaTuningEnvOverride(t *testing.T) {
	t.Setenv("WV_OLLAMA_KEEP_ALIVE", "10m")
	t.Setenv("WV_OLLAMA_NUM_CTX", "4096")

	cfg := Default()
	applyEnv(cfg)

	if cfg.Proxy.OllamaKeepAlive != "10m" {
		t.Errorf("OllamaKeepAlive = %q, want 10m", cfg.Proxy.OllamaKeepAlive)
	}
	if cfg.Proxy.OllamaNumCtx != 4096 {
		t.Errorf("OllamaNumCtx = %d, want 4096", cfg.Proxy.OllamaNumCtx)
	}
}

// TestOllamaTuningEnvIgnoresInvalid — bad numeric value should leave the
// existing setting alone instead of zeroing it out and silently killing the
// num_ctx hint.
func TestOllamaTuningEnvIgnoresInvalid(t *testing.T) {
	t.Setenv("WV_OLLAMA_NUM_CTX", "notanumber")

	cfg := Default()
	applyEnv(cfg)

	if cfg.Proxy.OllamaNumCtx != 8192 {
		t.Errorf("invalid NUM_CTX should preserve default; got %d", cfg.Proxy.OllamaNumCtx)
	}
}
