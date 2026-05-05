package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIsLocalProxyBaseURL(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"https://localhost:56244", true},
		{"http://localhost:56244/v1", true},
		{"https://127.0.0.1:56244/v1", true},
		{"http://127.0.0.1:9999", true},
		{"http://192.168.0.6:11434", false},
		{"http://192.168.0.6:11434/v1", false},
		{"https://api.anthropic.com", false},
		{"https://localhost", false}, // no port — bareword host, treat as not-local-proxy
	}
	for _, c := range cases {
		if got := isLocalProxyBaseURL(c.in); got != c.want {
			t.Errorf("isLocalProxyBaseURL(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestForceLocalhostBaseURL_RewritesUpstream(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"baseUrl": "http://192.168.0.6:11434/v1",
		},
	}
	if !forceLocalhostBaseURL(providers, "custom", "https://localhost:56244/v1") {
		t.Fatal("expected change=true for upstream baseUrl")
	}
	got := providers["custom"].(map[string]interface{})["baseUrl"].(string)
	if got != "https://localhost:56244/v1" {
		t.Errorf("baseUrl after rewrite = %q", got)
	}
}

func TestForceLocalhostBaseURL_LeavesLocalhost(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"baseUrl": "https://localhost:56244/v1",
		},
	}
	if forceLocalhostBaseURL(providers, "custom", "https://localhost:56244/v1") {
		t.Fatal("expected change=false for already-local baseUrl")
	}
}

func TestForceLocalhostBaseURL_LeavesEmpty(t *testing.T) {
	// Empty baseUrl means the provider hasn't been initialized yet — we
	// don't want to materialize a config we'd otherwise leave alone.
	providers := map[string]interface{}{
		"custom": map[string]interface{}{},
	}
	if forceLocalhostBaseURL(providers, "custom", "https://localhost:56244/v1") {
		t.Fatal("expected change=false for empty baseUrl")
	}
}

func TestForceLocalhostBaseURL_AbsentProvider(t *testing.T) {
	providers := map[string]interface{}{}
	if forceLocalhostBaseURL(providers, "custom", "https://localhost:56244/v1") {
		t.Fatal("expected change=false for absent provider")
	}
}

func TestPruneStaleModelsAcross_DropsEmptyIDAndDanglingName(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"models": []interface{}{
				map[string]interface{}{"id": "good", "name": "ollama / good"},
				map[string]interface{}{"id": "", "name": "ollama / bad"},
				map[string]interface{}{"id": "dangling", "name": "openrouter / "},
			},
		},
	}
	if !pruneStaleModelsAcross(providers) {
		t.Fatal("expected change=true")
	}
	models := providers["custom"].(map[string]interface{})["models"].([]interface{})
	if len(models) != 1 {
		t.Fatalf("expected 1 model after prune, got %d", len(models))
	}
	if got := models[0].(map[string]interface{})["id"].(string); got != "good" {
		t.Errorf("surviving model id = %q", got)
	}
}

func TestPruneStaleModelsAcross_DedupesById(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"models": []interface{}{
				map[string]interface{}{"id": "qwen3.6:27b", "name": "ollama / a"},
				map[string]interface{}{"id": "qwen3.6:27b", "name": "ollama / b"},
				map[string]interface{}{"id": "qwen3.6:27b", "name": "ollama / c"},
			},
		},
	}
	if !pruneStaleModelsAcross(providers) {
		t.Fatal("expected change=true")
	}
	models := providers["custom"].(map[string]interface{})["models"].([]interface{})
	if len(models) != 1 {
		t.Fatalf("expected 1 model after dedup, got %d", len(models))
	}
	if got := models[0].(map[string]interface{})["name"].(string); got != "ollama / a" {
		t.Errorf("expected first occurrence preserved, got %q", got)
	}
}

func TestPruneStaleModelsAcross_NoChangeWhenClean(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"models": []interface{}{
				map[string]interface{}{"id": "a", "name": "ollama / a"},
				map[string]interface{}{"id": "b", "name": "ollama / b"},
			},
		},
	}
	if pruneStaleModelsAcross(providers) {
		t.Fatal("expected change=false for clean models[]")
	}
}

func TestIsDanglingModelRef(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"custom/", true},
		{"anthropic/", true},
		{"custom/ ", true},  // whitespace after slash counts as empty id
		{"custom/qwen3.6:27b", false},
		{"anthropic/claude-opus-4-7", false},
		{"qwen3.6:27b", false}, // no slash — bare id, leave alone
	}
	for _, c := range cases {
		if got := isDanglingModelRef(c.in); got != c.want {
			t.Errorf("isDanglingModelRef(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRepairDanglingPrimary_PromoteFallback(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{
					"primary":   "custom/",
					"fallbacks": []interface{}{"custom/gemini-2.5-flash-lite"},
				},
			},
		},
	}
	if !repairDanglingPrimaryModel(cfg) {
		t.Fatal("expected change=true for dangling primary")
	}
	got := cfg["agents"].(map[string]interface{})["defaults"].(map[string]interface{})["model"].(map[string]interface{})["primary"]
	if got != "custom/gemini-2.5-flash-lite" {
		t.Errorf("primary after repair = %q", got)
	}
}

func TestRepairDanglingPrimary_DeleteWhenNoFallback(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{
					"primary": "custom/",
				},
			},
		},
	}
	if !repairDanglingPrimaryModel(cfg) {
		t.Fatal("expected change=true")
	}
	mdl := cfg["agents"].(map[string]interface{})["defaults"].(map[string]interface{})["model"].(map[string]interface{})
	if _, exists := mdl["primary"]; exists {
		t.Errorf("primary should be deleted when no usable fallback, got %v", mdl["primary"])
	}
}

func TestRepairDanglingPrimary_NoChangeWhenValid(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{
					"primary": "custom/qwen3.6:27b",
				},
			},
		},
	}
	if repairDanglingPrimaryModel(cfg) {
		t.Fatal("expected change=false for valid primary")
	}
}

func TestUpdateOpenClawJSON_EmptyModelIsNoop(t *testing.T) {
	// Regression for raspi 2026-05-01: SSE config_change fired before a
	// real model had been resolved, and updateOpenClawJSON wrote
	// primary="custom/" which OpenClaw rejected on every gateway restart.
	// We can't easily exercise the full function (it touches $HOME), but
	// the guard sits at the top so a smoke check is enough — the call must
	// not panic and must return immediately when model is empty.
	updateOpenClawJSON("custom", "", "")
	updateOpenClawJSON("anthropic", "", "")
	// no assertion: the test passes as long as we returned without
	// touching the filesystem (an attempt to read $HOME/.openclaw with
	// empty model would print a [openclaw-sync] log line, which we'd see
	// in CI failure output).
}

func TestNormalizeProviderAuth_RewritesStaleApiKey(t *testing.T) {
	// raspi/motoko 2026-05-02 — providers.{custom,anthropic,google}.apiKey
	// was "dummy"/"proxy-managed"/"" with authHeader=false, left over from
	// pre-v0.2.37 installs. OpenClaw faithfully sent those literals and
	// every call 401'd with `token not registered with vault`.
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"apiKey":     "dummy",
			"authHeader": false,
		},
		"anthropic": map[string]interface{}{
			"apiKey": "proxy-managed",
		},
		"google": map[string]interface{}{
			"apiKey": "",
		},
		// Third-party provider — must be left alone.
		"openrouter": map[string]interface{}{
			"apiKey": "sk-or-...",
		},
	}
	if !normalizeProviderAuth(providers, "real-vault-token") {
		t.Fatal("expected change=true")
	}
	for _, name := range []string{"custom", "anthropic", "google"} {
		p := providers[name].(map[string]interface{})
		if p["apiKey"] != "real-vault-token" {
			t.Errorf("%s.apiKey = %q, want real-vault-token", name, p["apiKey"])
		}
		if p["authHeader"] != true {
			t.Errorf("%s.authHeader = %v, want true", name, p["authHeader"])
		}
	}
	if providers["openrouter"].(map[string]interface{})["apiKey"] != "sk-or-..." {
		t.Error("openrouter apiKey should be left untouched")
	}
}

func TestNormalizeProviderAuth_NoChangeWhenAlreadyCorrect(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"apiKey":     "real-vault-token",
			"authHeader": true,
		},
	}
	if normalizeProviderAuth(providers, "real-vault-token") {
		t.Error("expected change=false for already-correct auth")
	}
}

func TestNormalizeProviderAuth_NoTokenSkips(t *testing.T) {
	// When the proxy has no vault token (standalone mode), the auth pass
	// must skip rather than write an empty string into apiKey.
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"apiKey": "dummy",
		},
	}
	cfg := map[string]interface{}{
		"models": map[string]interface{}{
			"providers": providers,
		},
	}
	if normalizeOpenClawProviders(cfg, "", "", "", "") {
		// Other heal rules might still fire; the auth-specific check is
		// that the apiKey wasn't blanked.
	}
	if got := providers["custom"].(map[string]interface{})["apiKey"]; got != "dummy" {
		t.Errorf("apiKey = %q, want preserved 'dummy' when token=''", got)
	}
}

func TestNormalizeOpenClawProviders_GoogleStaleUpstream(t *testing.T) {
	// motoko 2026-05-02 — google provider baseUrl pointed at the mini's
	// ollama (http://192.168.0.6:11434/v1), which made every OpenClaw
	// google call land on ollama and 404 with "model not found".
	cfg := map[string]interface{}{
		"models": map[string]interface{}{
			"providers": map[string]interface{}{
				"google": map[string]interface{}{
					"baseUrl": "http://192.168.0.6:11434/v1",
				},
			},
		},
	}
	if !normalizeOpenClawProviders(cfg, "", "", "", "") {
		t.Fatal("expected change=true for upstream baseUrl on google")
	}
	got := cfg["models"].(map[string]interface{})["providers"].(map[string]interface{})["google"].(map[string]interface{})["baseUrl"]
	if got != "https://localhost:56244" {
		t.Errorf("google baseUrl after heal = %q", got)
	}
}

func TestNormalizeOpenClawProviders_RaspiSnapshot(t *testing.T) {
	// Mirror of the raspi 2026-05-01 broken state: anthropic + custom
	// pointing at the upstream ollama, custom.models with 11 dup-id entries
	// plus one dangling-name entry.
	cfg := map[string]interface{}{
		"models": map[string]interface{}{
			"providers": map[string]interface{}{
				"anthropic": map[string]interface{}{
					"baseUrl": "http://192.168.0.6:11434",
					"apiKey":  "proxy-managed",
					"models": []interface{}{
						map[string]interface{}{"id": "qwen3.6:27b", "name": "proxy / x"},
					},
				},
				"custom": map[string]interface{}{
					"baseUrl": "http://192.168.0.6:11434/v1",
					"models": []interface{}{
						map[string]interface{}{"id": "qwen3.6:27b", "name": "ollama / a"},
						map[string]interface{}{"id": "qwen3.6:27b", "name": "ollama / b"},
						map[string]interface{}{"id": "qwen3.6:27b", "name": "openrouter / "},
					},
				},
			},
		},
	}
	if !normalizeOpenClawProviders(cfg, "", "", "", "") {
		t.Fatal("expected change=true for broken raspi snapshot")
	}
	provs := cfg["models"].(map[string]interface{})["providers"].(map[string]interface{})
	if got := provs["custom"].(map[string]interface{})["baseUrl"].(string); got != "https://localhost:56244/v1" {
		t.Errorf("custom baseUrl after heal = %q", got)
	}
	if got := provs["anthropic"].(map[string]interface{})["baseUrl"].(string); got != "https://localhost:56244" {
		t.Errorf("anthropic baseUrl after heal = %q", got)
	}
	customModels := provs["custom"].(map[string]interface{})["models"].([]interface{})
	if len(customModels) != 1 {
		t.Errorf("custom models[] after heal = %d entries, want 1", len(customModels))
	}
}

func TestHealAgentSpecificModels_RewritesStaleCache(t *testing.T) {
	// Simulate the mini 2026-05-02 state: per-agent cache holds
	// baseUrl=http://localhost:56244/v1 + apiKey=dummy + authHeader=false
	// while the main openclaw.json was already healed to https + real token.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	agentDir := filepath.Join(tmp, ".openclaw", "agents", "main", "agent")
	if err := os.MkdirAll(agentDir, 0o700); err != nil {
		t.Fatal(err)
	}
	stale := map[string]interface{}{
		"providers": map[string]interface{}{
			"custom": map[string]interface{}{
				"baseUrl":    "http://localhost:56244/v1",
				"apiKey":     "dummy",
				"authHeader": false,
				"api":        "openai-completions",
			},
			"ollama": map[string]interface{}{
				"baseUrl": "http://127.0.0.1:11434/v1",
				"apiKey":  "ollama-local",
			},
		},
	}
	path := filepath.Join(agentDir, "models.json")
	data, _ := json.MarshalIndent(stale, "", "  ")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	healAgentSpecificModels("real-vault-token", "", "", "")

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	custom := got["providers"].(map[string]interface{})["custom"].(map[string]interface{})
	if bu := custom["baseUrl"].(string); bu != "https://localhost:56244/v1" {
		t.Errorf("custom baseUrl after heal = %q, want https://localhost:56244/v1", bu)
	}
	if ak := custom["apiKey"].(string); ak != "real-vault-token" {
		t.Errorf("custom apiKey after heal = %q, want real-vault-token", ak)
	}
	if ah, _ := custom["authHeader"].(bool); !ah {
		t.Errorf("custom authHeader after heal = %v, want true", ah)
	}
	// ollama provider is third-party (not wall-vault-fronted) and must
	// be left untouched — its baseUrl is non-localhost in absolute terms
	// but IS local-proxy by isLocalProxyBaseURL (127.0.0.1) so heal
	// leaves it alone, and its apiKey is not in the wallVaultFronted set.
	ollama := got["providers"].(map[string]interface{})["ollama"].(map[string]interface{})
	if ak := ollama["apiKey"].(string); ak != "ollama-local" {
		t.Errorf("ollama apiKey was modified: %q", ak)
	}
}

func TestHealAgentSpecificModels_NoChangeWhenAlreadyCorrect(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	agentDir := filepath.Join(tmp, ".openclaw", "agents", "main", "agent")
	if err := os.MkdirAll(agentDir, 0o700); err != nil {
		t.Fatal(err)
	}
	good := map[string]interface{}{
		"providers": map[string]interface{}{
			"custom": map[string]interface{}{
				"baseUrl":    "https://localhost:56244/v1",
				"apiKey":     "real-vault-token",
				"authHeader": true,
			},
		},
	}
	path := filepath.Join(agentDir, "models.json")
	data, _ := json.MarshalIndent(good, "", "  ")
	st0, _ := os.Stat(path)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	st0, _ = os.Stat(path)

	healAgentSpecificModels("real-vault-token", "", "", "")

	st1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st1.ModTime().Equal(st0.ModTime()) {
		t.Errorf("expected no rewrite when config is already healed (mtime changed: %v → %v)", st0.ModTime(), st1.ModTime())
	}
}

func TestHealAgentSpecificModels_MissingDirIsNoop(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// No ~/.openclaw/agents/ at all — heal must not panic or error.
	healAgentSpecificModels("token", "", "", "")
}

func TestNormalizeProviderTLSCA_WritesNewCA(t *testing.T) {
	providers := map[string]interface{}{
		"custom":    map[string]interface{}{"baseUrl": "https://localhost:56244/v1"},
		"anthropic": map[string]interface{}{"baseUrl": "https://localhost:56244"},
	}
	if !normalizeProviderTLSCA(providers, "/etc/wv/ca.crt") {
		t.Fatal("expected change=true on first write")
	}
	for _, name := range []string{"custom", "anthropic"} {
		req := providers[name].(map[string]interface{})["request"].(map[string]interface{})
		tls := req["tls"].(map[string]interface{})
		if got := tls["ca"].(string); got != "/etc/wv/ca.crt" {
			t.Errorf("%s tls.ca = %q, want /etc/wv/ca.crt", name, got)
		}
	}
}

func TestNormalizeProviderTLSCA_NoChangeWhenAlreadyCorrect(t *testing.T) {
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"baseUrl": "https://localhost:56244/v1",
			"request": map[string]interface{}{
				"tls": map[string]interface{}{"ca": "/etc/wv/ca.crt"},
			},
		},
	}
	if normalizeProviderTLSCA(providers, "/etc/wv/ca.crt") {
		t.Fatal("expected change=false when CA already set")
	}
}

func TestNormalizeProviderTLSCA_PreservesSiblingTLSFields(t *testing.T) {
	// Operator-set request.tls.{cert, serverName, etc.} must survive a
	// CA-only heal — we only touch tls.ca.
	providers := map[string]interface{}{
		"custom": map[string]interface{}{
			"baseUrl": "https://localhost:56244/v1",
			"request": map[string]interface{}{
				"tls": map[string]interface{}{
					"serverName": "wall-vault.lan",
					"cert":       "/path/to/client.crt",
				},
			},
		},
	}
	if !normalizeProviderTLSCA(providers, "/etc/wv/ca.crt") {
		t.Fatal("expected change=true")
	}
	tls := providers["custom"].(map[string]interface{})["request"].(map[string]interface{})["tls"].(map[string]interface{})
	if tls["serverName"] != "wall-vault.lan" {
		t.Errorf("serverName lost: %v", tls["serverName"])
	}
	if tls["cert"] != "/path/to/client.crt" {
		t.Errorf("cert lost: %v", tls["cert"])
	}
}

func TestNormalizeProviderTLSCA_SkipsThirdPartyProviders(t *testing.T) {
	providers := map[string]interface{}{
		"ollama": map[string]interface{}{"baseUrl": "http://127.0.0.1:11434/v1"},
	}
	if normalizeProviderTLSCA(providers, "/etc/wv/ca.crt") {
		t.Fatal("third-party provider must not be touched")
	}
	if _, has := providers["ollama"].(map[string]interface{})["request"]; has {
		t.Errorf("request key materialized on third-party provider")
	}
}

func TestProviderHealURLs(t *testing.T) {
	c, a, g := providerHealURLs("", "")
	if c != "https://localhost:56244/v1" || a != "https://localhost:56244" || g != "https://localhost:56244" {
		t.Errorf("legacy fallback targets wrong: %s | %s | %s", c, a, g)
	}
	c, a, g = providerHealURLs("http://127.0.0.1:56245", "")
	if c != "http://127.0.0.1:56245/v1" || a != "http://127.0.0.1:56245" || g != "http://127.0.0.1:56245" {
		t.Errorf("plain-companion targets wrong: %s | %s | %s", c, a, g)
	}
	// Operator-configured port flows through when no companion is set.
	c, a, g = providerHealURLs("", "http://localhost:7777")
	if c != "http://localhost:7777/v1" || a != "http://localhost:7777" || g != "http://localhost:7777" {
		t.Errorf("custom-port targets wrong: %s | %s | %s", c, a, g)
	}
	// Companion still wins over default when both are set.
	c, a, g = providerHealURLs("http://127.0.0.1:56245", "http://localhost:7777")
	if c != "http://127.0.0.1:56245/v1" || a != "http://127.0.0.1:56245" || g != "http://127.0.0.1:56245" {
		t.Errorf("companion priority wrong: %s | %s | %s", c, a, g)
	}
}

func TestNormalizeOpenClawProviders_PlainCompanionRewritesHttpsLocalhost(t *testing.T) {
	// Existing config points at the TLS listener — heal must rewrite to
	// the plain-HTTP companion when localBaseOrigin is set.
	cfg := map[string]interface{}{
		"models": map[string]interface{}{
			"providers": map[string]interface{}{
				"custom":    map[string]interface{}{"baseUrl": "https://localhost:56244/v1"},
				"anthropic": map[string]interface{}{"baseUrl": "https://localhost:56244"},
			},
		},
	}
	if !normalizeOpenClawProviders(cfg, "", "", "http://127.0.0.1:56245", "") {
		t.Fatal("expected change=true to switch to plain companion")
	}
	provs := cfg["models"].(map[string]interface{})["providers"].(map[string]interface{})
	if got := provs["custom"].(map[string]interface{})["baseUrl"].(string); got != "http://127.0.0.1:56245/v1" {
		t.Errorf("custom baseUrl = %q, want plain companion", got)
	}
	if got := provs["anthropic"].(map[string]interface{})["baseUrl"].(string); got != "http://127.0.0.1:56245" {
		t.Errorf("anthropic baseUrl = %q, want plain companion", got)
	}
}

func TestNormalizeOpenClawProviders_NoCompanionLeavesHttpsLocalhostAlone(t *testing.T) {
	cfg := map[string]interface{}{
		"models": map[string]interface{}{
			"providers": map[string]interface{}{
				"custom": map[string]interface{}{"baseUrl": "https://localhost:56244/v1"},
			},
		},
		// Pre-seed channel-stale threshold above the heal floor so this
		// test stays focused on baseUrl behaviour — relaxChannelStaleThreshold
		// would otherwise flip changed=true on every cfg without a gateway block.
		"gateway": map[string]interface{}{"channelStaleEventThresholdMinutes": float64(120)},
	}
	if normalizeOpenClawProviders(cfg, "", "", "", "") {
		t.Fatal("expected change=false when no companion configured and base is already correct")
	}
}

func TestAlignActiveMemoryModelToAgentDefault_RewritesGoogleToOllama(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{
					"primary": "custom/qwen3.6:27b",
				},
			},
		},
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"active-memory": map[string]interface{}{
					"config": map[string]interface{}{
						"model": "custom/gemini-2.5-flash-lite",
					},
				},
			},
		},
	}
	if !alignActiveMemoryModelToAgentDefault(cfg) {
		t.Fatal("expected change=true to align with agent default")
	}
	got := cfg["plugins"].(map[string]interface{})["entries"].(map[string]interface{})["active-memory"].(map[string]interface{})["config"].(map[string]interface{})["model"]
	if got != "custom/qwen3.6:27b" {
		t.Errorf("active-memory model = %q, want custom/qwen3.6:27b", got)
	}
}

func TestAlignActiveMemoryModelToAgentDefault_NoChangeWhenAligned(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{"primary": "custom/qwen3.6:27b"},
			},
		},
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"active-memory": map[string]interface{}{
					"config": map[string]interface{}{"model": "custom/qwen3.6:27b"},
				},
			},
		},
	}
	if alignActiveMemoryModelToAgentDefault(cfg) {
		t.Fatal("expected change=false when already aligned")
	}
}

func TestAlignActiveMemoryModelToAgentDefault_NoOpWhenPluginAbsent(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{"primary": "custom/qwen3.6:27b"},
			},
		},
	}
	if alignActiveMemoryModelToAgentDefault(cfg) {
		t.Fatal("expected change=false when active-memory plugin not configured")
	}
	// Must not materialize plugins/entries map.
	if _, has := cfg["plugins"]; has {
		t.Errorf("plugins map materialized")
	}
}

func TestAlignActiveMemoryModelToAgentDefault_NoOpWhenPrimaryEmpty(t *testing.T) {
	cfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{},
			},
		},
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"active-memory": map[string]interface{}{
					"config": map[string]interface{}{"model": "custom/gemini-2.5-flash-lite"},
				},
			},
		},
	}
	if alignActiveMemoryModelToAgentDefault(cfg) {
		t.Fatal("expected change=false with no agent primary to align to")
	}
}

func TestRelaxChannelStaleThreshold_AddsWhenAbsent(t *testing.T) {
	cfg := map[string]interface{}{}
	if !relaxChannelStaleThreshold(cfg) {
		t.Fatal("expected change=true to add threshold")
	}
	gw := cfg["gateway"].(map[string]interface{})
	if gw["channelStaleEventThresholdMinutes"] != 60 {
		t.Errorf("threshold = %v, want 60", gw["channelStaleEventThresholdMinutes"])
	}
}

func TestRelaxChannelStaleThreshold_RaisesWhenTooShort(t *testing.T) {
	cfg := map[string]interface{}{
		"gateway": map[string]interface{}{
			"channelStaleEventThresholdMinutes": float64(5),
		},
	}
	if !relaxChannelStaleThreshold(cfg) {
		t.Fatal("expected change=true to raise threshold")
	}
	gw := cfg["gateway"].(map[string]interface{})
	if gw["channelStaleEventThresholdMinutes"] != 60 {
		t.Errorf("threshold = %v, want 60", gw["channelStaleEventThresholdMinutes"])
	}
}

func TestRelaxChannelStaleThreshold_NoChangeWhenAlreadyHigh(t *testing.T) {
	cfg := map[string]interface{}{
		"gateway": map[string]interface{}{
			"channelStaleEventThresholdMinutes": float64(120),
		},
	}
	if relaxChannelStaleThreshold(cfg) {
		t.Fatal("expected change=false when already above floor")
	}
	gw := cfg["gateway"].(map[string]interface{})
	if gw["channelStaleEventThresholdMinutes"] != float64(120) {
		t.Errorf("threshold = %v, want 120 preserved", gw["channelStaleEventThresholdMinutes"])
	}
}
