package proxy

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sookmook/wall-vault/internal/vault"
)

func TestDispatch_FallbackUsesEachServicesDefaultModel(t *testing.T) {
	services := []vault.Service{
		{ID: "google", DefaultModel: "gemini-3.1-pro-preview", ProxyEnabled: true, SortOrder: 1},
		{ID: "ollama", DefaultModel: "gemma4:26b", ProxyEnabled: true, SortOrder: 4},
	}
	client := vault.Client{ID: "bot-a", PreferredService: "google"}

	var calls []string
	caller := func(svc, mdl string) error {
		calls = append(calls, svc+"/"+mdl)
		if svc == "google" {
			return fmt.Errorf("google HTTP 400")
		}
		return nil
	}

	err := dispatchForTest(client, services, caller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"google/gemini-3.1-pro-preview", "ollama/gemma4:26b"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%v, want %v", calls, want)
	}
}
