package vault

import (
	"testing"
)

// v1Envelope mirrors the legacy shape just enough to exercise migration.
type v1Envelope struct {
	Services []struct {
		ID, Name, LocalURL string
		Enabled, ProxyEnabled bool
	} `json:"services"`
	Clients []struct {
		ID, Name, Token string
		DefaultService, DefaultModel string
		AllowedServices []string
		AgentType string
		Enabled bool
		SortOrder int
	} `json:"clients"`
	APIKeys []*APIKey `json:"api_keys"`
}

func TestMigrateV1ToV2(t *testing.T) {
	v1Raw := []byte(`{
		"services": [
			{"id":"google","name":"Google","enabled":true,"proxy_enabled":true},
			{"id":"ollama","name":"Ollama","local_url":"http://192.168.1.20:11434","enabled":true,"proxy_enabled":true}
		],
		"clients": [
			{"id":"bot-a","name":"Delta","token":"t","default_service":"google","default_model":"gemini-3.1-pro-preview","agent_type":"nanoclaw","enabled":true,"sort_order":4},
			{"id":"bot-b","name":"Bravo","token":"t2","default_service":"google","default_model":"gemini-3.1-pro-preview","agent_type":"openclaw","enabled":true,"sort_order":1}
		],
		"api_keys": []
	}`)
	out, err := MigrateV1ToV2(v1Raw)
	if err != nil { t.Fatal(err) }
	if out.SchemaVersion != CurrentSchemaVersion { t.Fatalf("want version %d, got %d", CurrentSchemaVersion, out.SchemaVersion) }
	// google service inherits the most-common default_model from clients
	var googleDM string
	for _, s := range out.Services {
		if s.ID == "google" { googleDM = s.DefaultModel }
	}
	if googleDM != "gemini-3.1-pro-preview" {
		t.Fatalf("google default_model = %q, want gemini-3.1-pro-preview", googleDM)
	}
	// clients renamed
	var bot-a Client
	for _, c := range out.Clients {
		if c.ID == "bot-a" { bot-a = *c }
	}
	if bot-a.PreferredService != "google" {
		t.Fatalf("preferred_service = %q", bot-a.PreferredService)
	}
	if bot-a.ModelOverride != "gemini-3.1-pro-preview" {
		t.Fatalf("model_override = %q", bot-a.ModelOverride)
	}
}
