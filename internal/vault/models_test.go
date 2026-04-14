package vault

import (
	"encoding/json"
	"testing"
)

func TestServiceRoundTrip(t *testing.T) {
	in := Service{
		ID: "google", Name: "Google Gemini",
		DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview", "gemini-3.1-flash-lite-preview"},
		Enabled: true, ProxyEnabled: true, SortOrder: 2,
	}
	buf, err := json.Marshal(in)
	if err != nil { t.Fatal(err) }
	var out Service
	if err := json.Unmarshal(buf, &out); err != nil { t.Fatal(err) }
	if out.DefaultModel != in.DefaultModel || len(out.AllowedModels) != 2 {
		t.Fatalf("roundtrip mismatch: %+v", out)
	}
}

func TestClientNewFields(t *testing.T) {
	c := Client{ID: "bot-a", PreferredService: "google", ModelOverride: ""}
	if c.PreferredService != "google" {
		t.Fatal("PreferredService missing")
	}
}
