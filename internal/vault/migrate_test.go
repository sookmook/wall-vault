package vault

import (
	"testing"
)

func TestMigrateV1ToV2(t *testing.T) {
	v1Raw := []byte(`{
		"services": [
			{"id":"google","name":"Google","enabled":true,"proxy_enabled":true},
			{"id":"ollama","name":"Ollama","local_url":"http://192.168.1.20:11434","enabled":true,"proxy_enabled":true}
		],
		"clients": [
			{"id":"bot-a","name":"Alpha","token":"t","default_service":"google","default_model":"gemini-3.1-pro-preview","agent_type":"nanoclaw","enabled":true,"sort_order":4},
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
	var botA Client
	for _, c := range out.Clients {
		if c.ID == "bot-a" { botA = *c }
	}
	if botA.PreferredService != "google" {
		t.Fatalf("preferred_service = %q", botA.PreferredService)
	}
	if botA.ModelOverride != "gemini-3.1-pro-preview" {
		t.Fatalf("model_override = %q", botA.ModelOverride)
	}
}

// TestMigrateV1ToV2_ServiceWithNoClients verifies that a service with no
// referencing clients produces DefaultModel == "".
func TestMigrateV1ToV2_ServiceWithNoClients(t *testing.T) {
	v1Raw := []byte(`{
		"services": [
			{"id":"openrouter","name":"OpenRouter","enabled":true,"proxy_enabled":true}
		],
		"clients": [],
		"api_keys": []
	}`)
	out, err := MigrateV1ToV2(v1Raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(out.Services))
	}
	if out.Services[0].DefaultModel != "" {
		t.Fatalf("expected empty DefaultModel, got %q", out.Services[0].DefaultModel)
	}
}

// TestMigrateV1ToV2_MajorityVoteTieBreak verifies that when two clients use
// different models on the same service with equal vote counts, the client with
// the lowest sort_order wins. Two services are used so neither service
// coincidentally picks the same model.
func TestMigrateV1ToV2_MajorityVoteTieBreak(t *testing.T) {
	v1Raw := []byte(`{
		"services": [
			{"id":"google","name":"Google","enabled":true,"proxy_enabled":true,"sort_order":1},
			{"id":"openrouter","name":"OpenRouter","enabled":true,"proxy_enabled":true,"sort_order":2}
		],
		"clients": [
			{"id":"c1","name":"C1","token":"t1","default_service":"google","default_model":"model-alpha","enabled":true,"sort_order":1},
			{"id":"c2","name":"C2","token":"t2","default_service":"google","default_model":"model-beta","enabled":true,"sort_order":2},
			{"id":"c3","name":"C3","token":"t3","default_service":"openrouter","default_model":"model-gamma","enabled":true,"sort_order":1},
			{"id":"c4","name":"C4","token":"t4","default_service":"openrouter","default_model":"model-delta","enabled":true,"sort_order":2}
		],
		"api_keys": []
	}`)
	out, err := MigrateV1ToV2(v1Raw)
	if err != nil {
		t.Fatal(err)
	}
	var googleDM, openrouterDM string
	for _, s := range out.Services {
		switch s.ID {
		case "google":
			googleDM = s.DefaultModel
		case "openrouter":
			openrouterDM = s.DefaultModel
		}
	}
	// Each service has a 1-1 tie; lowest sort_order client's model must win.
	if googleDM != "model-alpha" {
		t.Fatalf("google tiebreak: want model-alpha, got %q", googleDM)
	}
	if openrouterDM != "model-gamma" {
		t.Fatalf("openrouter tiebreak: want model-gamma, got %q", openrouterDM)
	}
}

// TestMigrateV1ToV2_SkipEmptyFields verifies that clients with an empty
// default_service or empty default_model do not contribute to the vote.
func TestMigrateV1ToV2_SkipEmptyFields(t *testing.T) {
	v1Raw := []byte(`{
		"services": [
			{"id":"google","name":"Google","enabled":true,"proxy_enabled":true}
		],
		"clients": [
			{"id":"no-service","name":"NoSvc","token":"t1","default_service":"","default_model":"gemini-3.1-pro","enabled":true,"sort_order":1},
			{"id":"no-model","name":"NoMdl","token":"t2","default_service":"google","default_model":"","enabled":true,"sort_order":2},
			{"id":"valid","name":"Valid","token":"t3","default_service":"google","default_model":"gemini-exp-1206","enabled":true,"sort_order":3}
		],
		"api_keys": []
	}`)
	out, err := MigrateV1ToV2(v1Raw)
	if err != nil {
		t.Fatal(err)
	}
	var googleDM string
	for _, s := range out.Services {
		if s.ID == "google" {
			googleDM = s.DefaultModel
		}
	}
	// Only the "valid" client should have voted; its model must win.
	if googleDM != "gemini-exp-1206" {
		t.Fatalf("want gemini-exp-1206, got %q", googleDM)
	}
}
