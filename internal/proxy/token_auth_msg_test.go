package proxy

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// tokenAuthFail used to fall through to the same generic "invalid token" 401
// regardless of why the lookup failed. An earlier reviewer report documented
// real ops cost — operators spent time chasing IP whitelist when the actual
// cause was vault-side. Each branch below pins the wire-format message a
// caller sees so a future refactor that conflates them again fails the
// test instead of silently regressing the diagnostics.

func TestTokenAuthFail_NotRegistered(t *testing.T) {
	w := httptest.NewRecorder()
	tokenAuthFail(w, tokenLookupNotRegistered, true)
	if w.Code != 401 {
		t.Errorf("not-registered status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "not registered") {
		t.Errorf("missing 'not registered' hint: %s", w.Body.String())
	}
}

func TestTokenAuthFail_VaultUnreachable(t *testing.T) {
	w := httptest.NewRecorder()
	tokenAuthFail(w, tokenLookupVaultUnreachable, true)
	if w.Code != 503 {
		t.Errorf("unreachable status = %d, want 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), "vault unreachable") {
		t.Errorf("missing 'vault unreachable' hint: %s", w.Body.String())
	}
}

func TestTokenAuthFail_VaultError(t *testing.T) {
	w := httptest.NewRecorder()
	tokenAuthFail(w, tokenLookupVaultError, true)
	if w.Code != 502 {
		t.Errorf("vault error status = %d, want 502", w.Code)
	}
}

// SkippedWithProxyToken — proxy.vault_token configured but mismatched, no
// vault fallback. Caller needs to know it's their proxy token, not vault.
func TestTokenAuthFail_SkippedWithProxyToken(t *testing.T) {
	w := httptest.NewRecorder()
	tokenAuthFail(w, tokenLookupSkipped, true)
	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "proxy.vault_token mismatch") {
		t.Errorf("missing 'proxy.vault_token mismatch' hint: %s", w.Body.String())
	}
}

// SkippedNoProxyToken — neither proxy.vault_token nor vault_url set; the
// proxy is essentially un-auth-configured. 503 because the operator's
// expectation (auth) can't be served at all.
func TestTokenAuthFail_SkippedNoProxyToken(t *testing.T) {
	w := httptest.NewRecorder()
	tokenAuthFail(w, tokenLookupSkipped, false)
	if w.Code != 503 {
		t.Errorf("status = %d, want 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), "no auth configured") {
		t.Errorf("missing 'no auth configured' hint: %s", w.Body.String())
	}
}
