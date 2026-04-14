package proxy

import (
	"errors"
	"testing"

	"github.com/sookmook/wall-vault/internal/vault"
)

func TestResolveModel_DefaultWhenNoOverride(t *testing.T) {
	svc := vault.Service{ID: "google", DefaultModel: "gemini-3.1-pro-preview"}
	c := vault.Client{ID: "bot-a", ModelOverride: ""}
	m, err := ResolveModel(c, svc)
	if err != nil || m != "gemini-3.1-pro-preview" {
		t.Fatalf("got (%q, %v)", m, err)
	}
}

func TestResolveModel_OverrideWithEmptyAllowed(t *testing.T) {
	svc := vault.Service{ID: "google", DefaultModel: "gemini-3.1-pro-preview"}
	c := vault.Client{ModelOverride: "gemini-2.5-flash-lite"}
	m, err := ResolveModel(c, svc)
	if err != nil || m != "gemini-2.5-flash-lite" {
		t.Fatalf("got (%q, %v)", m, err)
	}
}

func TestResolveModel_OverrideInWhitelist(t *testing.T) {
	svc := vault.Service{
		ID: "google", DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview", "gemini-3.1-flash-lite-preview"},
	}
	c := vault.Client{ModelOverride: "gemini-3.1-flash-lite-preview"}
	m, err := ResolveModel(c, svc)
	if err != nil || m != "gemini-3.1-flash-lite-preview" {
		t.Fatalf("got (%q, %v)", m, err)
	}
}

func TestResolveModel_OverrideNotInWhitelist(t *testing.T) {
	svc := vault.Service{
		ID: "google", DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview"},
	}
	c := vault.Client{ModelOverride: "gemini-2.5-flash-lite"}
	_, err := ResolveModel(c, svc)
	if !errors.Is(err, ErrModelNotAllowed) {
		t.Fatalf("want ErrModelNotAllowed, got %v", err)
	}
}
