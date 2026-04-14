package proxy

// dispatch_v2.go — v0.2 dispatch core helpers.
//
// The legacy Server.dispatch (dispatch with *ServiceConfig / string service+model
// parameters) remains untouched in server.go and continues to serve all existing
// HTTP handlers until Stage 4 replaces them.
//
// dispatchWith and dispatchForTest live here so they can be unit-tested without
// importing or touching the legacy path.  They use the v0.2 canonical types
// vault.Service and vault.Client exclusively.

import (
	"fmt"
	"log"

	"github.com/sookmook/wall-vault/internal/vault"
)

// dispatchWith is the testable v0.2 dispatch core.  It iterates the provided
// services in order, resolving each step's model via ResolveModel.  The actual
// upstream call is delegated to `call`, allowing real handlers to supply a
// (svc, model) → real upstream invocation, and unit tests to inject a stub.
//
// services is expected to be ordered: primary first, then fallback chain by
// Service.SortOrder ascending.  The caller is responsible for that ordering.
//
// Cooldown skipping is also the caller's responsibility; the real call site
// should filter services through keymgr before passing them in.  Tests may
// pass an already-filtered slice.
//
// TODO (v0.2 Stage 4): wire Server.dispatch to call dispatchWith internally
// and delete the legacy dispatch implementation.
func dispatchWith(
	client vault.Client,
	services []vault.Service,
	call func(vault.Service, string) (*GeminiResponse, error),
) (*GeminiResponse, error) {
	var lastErr error
	for _, svc := range services {
		mdl, err := ResolveModel(client, svc)
		if err != nil {
			log.Printf("[dispatch] skip %s: %v", svc.ID, err)
			lastErr = err
			continue
		}
		resp, err := call(svc, mdl)
		if err == nil {
			return resp, nil
		}
		log.Printf("[dispatch] %s failed: %v", svc.ID, err)
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no enabled services")
	}
	return nil, fmt.Errorf("모든 서비스 실패: %w", lastErr)
}

// dispatchForTest is a thin wrapper exposing dispatchWith to unit tests via a
// (svc, mdl) string-pair caller.  Tests don't need to construct a GeminiResponse;
// they just observe the call sequence and return errors where desired.
func dispatchForTest(
	client vault.Client,
	services []vault.Service,
	caller func(svc, mdl string) error,
) error {
	_, err := dispatchWith(client, services, func(s vault.Service, m string) (*GeminiResponse, error) {
		return &GeminiResponse{}, caller(s.ID, m)
	})
	return err
}
