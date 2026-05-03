package proxy

import "testing"

func TestInferServiceFromBareModel(t *testing.T) {
	cases := []struct {
		name string
		mdl  string
		want string
	}{
		// post #38 surfaced: bare gemini name addressed to a client whose
		// preferred_service was ollama. inference must promote it to google.
		{"gemini-flash",      "gemini-2.5-flash",          "google"},
		{"gemini-pro",        "gemini-3.1-pro-preview",    "google"},
		{"gemma small",       "gemma-2b",                  "google"},

		{"claude opus",       "claude-opus-4-7",           "anthropic"},
		{"claude sonnet",     "claude-3-5-sonnet",         "anthropic"},

		{"gpt-4o",            "gpt-4o",                    "openai"},
		{"o1 plain",          "o1",                        "openai"},
		{"o1 mini",           "o1-mini",                   "openai"},
		{"o3 family",         "o3-mini-2025",              "openai"},
		{"o4 family",         "o4-mini",                   "openai"},

		// tag-style — ollama's local-model convention. Even ambiguous root
		// names (gemma4:*) belong to ollama once the tag is present.
		{"ollama qwen",       "qwen3.6:27b",               "ollama"},
		{"ollama gemma",      "gemma4:26b",                "ollama"},
		{"ollama llama",      "llama3:8b",                 "ollama"},

		// :cloud is handled earlier in parseProviderModel — inferer just sees
		// the bare colon-form and routes to ollama. parseProviderModel itself
		// strips :cloud and routes to openrouter before calling us.
		{"ollama latest tag", "deepseek-r1:latest",        "ollama"},

		// genuinely ambiguous / unknown — leave caller's choice in place.
		{"empty",             "",                          ""},
		{"qwen sans tag",     "qwen3.5-32b",               ""},
		{"deepseek bare",     "deepseek-r1",               ""},
		{"unknown",           "some-private-model-name",   ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := inferServiceFromBareModel(tc.mdl)
			if got != tc.want {
				t.Errorf("inferServiceFromBareModel(%q) = %q, want %q", tc.mdl, got, tc.want)
			}
		})
	}
}

func TestParseProviderModel_BarePromotion(t *testing.T) {
	cases := []struct {
		name    string
		svc     string
		mdl     string
		wantSvc string
		wantMdl string
	}{
		// Post #38 reproduction: bare gemini name to ollama-preferred client
		// must be promoted to google instead of force-routed to ollama.
		{"gemini → google",   "ollama",     "gemini-2.5-flash",   "google",    "gemini-2.5-flash"},
		{"claude → anthropic","ollama",     "claude-opus-4-7",    "anthropic", "claude-opus-4-7"},
		{"gpt → openai",      "ollama",     "gpt-4o",             "openai",    "gpt-4o"},

		// Already-correct service: inferer agrees, no change.
		{"google native",     "google",     "gemini-2.5-flash",   "google",    "gemini-2.5-flash"},
		{"anthropic native",  "anthropic",  "claude-opus-4-7",    "anthropic", "claude-opus-4-7"},

		// Tag-style ollama name — preferred ollama, stays ollama.
		{"ollama tag stays",  "ollama",     "qwen3.6:27b",        "ollama",    "qwen3.6:27b"},

		// Ambiguous bare name — leave caller's preferred untouched.
		{"unknown stays",     "ollama",     "some-private-name",  "ollama",    "some-private-name"},

		// Provider prefix already present — earlier branch wins, inferer not consulted.
		{"google/ prefix",    "ollama",     "google/gemini-flash","google",    "google/gemini-flash"},
		{"anthropic/ prefix", "openrouter", "anthropic/claude",   "openrouter","anthropic/claude"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotSvc, gotMdl := parseProviderModel(tc.svc, tc.mdl)
			if gotSvc != tc.wantSvc || gotMdl != tc.wantMdl {
				t.Errorf("parseProviderModel(%q,%q) = (%q,%q), want (%q,%q)",
					tc.svc, tc.mdl, gotSvc, gotMdl, tc.wantSvc, tc.wantMdl)
			}
		})
	}
}

// TestParseProviderModel_OAICompatMatrix exercises the oaiCompatServices set
// for the two directions the parser uses it: (a) caller chose an OAI-compat
// backend as svc — model prefix must not hijack the route to a native handler;
// (b) caller wrote an OAI-compat backend as the model prefix — route must land
// on that backend regardless of the original svc. Ollama prefix is the one
// documented escape hatch (operator-wired Modelfile aliases).
func TestParseProviderModel_OAICompatMatrix(t *testing.T) {
	backends := []string{
		"lmstudio", "vllm", "llamacpp", "tgwui", "localai",
		"jan", "koboldcpp", "tabbyapi", "mlx-server", "litellm-proxy",
	}
	prefixes := []string{"google", "anthropic", "qwen", "meta-llama", "openai"}

	for _, b := range backends {
		for _, p := range prefixes {
			mdl := p + "/some-model-7b"

			// (a) svc=backend + foreign prefix → svc honoured (no hijack).
			t.Run(b+"/svc_keeps_"+p, func(t *testing.T) {
				gotSvc, gotMdl := parseProviderModel(b, mdl)
				if gotSvc != b || gotMdl != mdl {
					t.Errorf("parseProviderModel(%q,%q) = (%q,%q), want (%q,%q)",
						b, mdl, gotSvc, gotMdl, b, mdl)
				}
			})

			// (b) svc=ollama + backend prefix → routes to backend.
			pmdl := b + "/some-model-7b"
			t.Run(b+"/prefix_routes_from_ollama", func(t *testing.T) {
				gotSvc, gotMdl := parseProviderModel("ollama", pmdl)
				if gotSvc != b || gotMdl != pmdl {
					t.Errorf("parseProviderModel(%q,%q) = (%q,%q), want (%q,%q)",
						"ollama", pmdl, gotSvc, gotMdl, b, pmdl)
				}
			})
		}

		// Documented escape: svc=backend + ollama/* prefix bypasses honour-svc.
		// Caller's "ollama/qwen3.6:27b" still lands on ollama even when svc was
		// pinned to an OAI-compat backend (matches the ollama-bypass clause).
		t.Run(b+"/ollama_prefix_bypass", func(t *testing.T) {
			gotSvc, gotMdl := parseProviderModel(b, "ollama/qwen3.6:27b")
			if gotSvc != "ollama" || gotMdl != "qwen3.6:27b" {
				t.Errorf("parseProviderModel(%q, ollama/qwen3.6:27b) = (%q,%q), want (ollama, qwen3.6:27b)",
					b, gotSvc, gotMdl)
			}
		})
	}
}
