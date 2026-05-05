package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sookmook/wall-vault/internal/config"
)

func TestIsSelfManagedSentinel(t *testing.T) {
	cases := []struct {
		token string
		want  bool
	}{
		{"proxy-managed", true},
		{"dummy", true},
		{"", false},
		{"real-token-abc", false},
		{"PROXY-MANAGED", false}, // case-sensitive — sentinels are exact
	}
	for _, c := range cases {
		if got := isSelfManagedSentinel(c.token); got != c.want {
			t.Errorf("isSelfManagedSentinel(%q) = %v, want %v", c.token, got, c.want)
		}
	}
}

func TestIsLoopbackHostPort(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"127.0.0.1:54321", true},
		{"[::1]:54321", true},
		{"localhost:54321", true},
		{"192.168.0.20:54321", false},
		{"10.0.0.1:80", false},
		{"127.0.0.1", true},
		{"::1", true},
		{"example.com:443", false},
	}
	for _, c := range cases {
		if got := isLoopbackHostPort(c.addr); got != c.want {
			t.Errorf("isLoopbackHostPort(%q) = %v, want %v", c.addr, got, c.want)
		}
	}
}

func TestSubstituteSelfManagedSentinel_Disabled(t *testing.T) {
	s := &Server{cfg: &config.Config{Proxy: config.ProxyConfig{
		TokenSentinelFallback: false,
		VaultToken:            "real-vault-token",
	}}}
	r := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	r.Header.Set("Authorization", "Bearer proxy-managed")
	r.RemoteAddr = "127.0.0.1:55555"
	if s.substituteSelfManagedSentinel(r) {
		t.Fatal("expected no substitution when fallback flag is off")
	}
	if got := r.Header.Get("Authorization"); got != "Bearer proxy-managed" {
		t.Errorf("header changed unexpectedly: %q", got)
	}
}

func TestSubstituteSelfManagedSentinel_Loopback(t *testing.T) {
	s := &Server{cfg: &config.Config{Proxy: config.ProxyConfig{
		TokenSentinelFallback: true,
		VaultToken:            "real-vault-token",
	}}}
	r := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	r.Header.Set("Authorization", "Bearer proxy-managed")
	r.RemoteAddr = "127.0.0.1:55555"
	if !s.substituteSelfManagedSentinel(r) {
		t.Fatal("expected substitution for loopback + sentinel + flag-on")
	}
	if got := r.Header.Get("Authorization"); got != "Bearer real-vault-token" {
		t.Errorf("Authorization not rewritten: got %q", got)
	}
}

func TestSubstituteSelfManagedSentinel_NonLoopbackBlocked(t *testing.T) {
	s := &Server{cfg: &config.Config{Proxy: config.ProxyConfig{
		TokenSentinelFallback: true,
		VaultToken:            "real-vault-token",
	}}}
	r := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	r.Header.Set("Authorization", "Bearer proxy-managed")
	r.RemoteAddr = "192.168.0.20:55555"
	if s.substituteSelfManagedSentinel(r) {
		t.Fatal("non-loopback caller must not get substitution")
	}
	if got := r.Header.Get("Authorization"); got != "Bearer proxy-managed" {
		t.Errorf("header changed unexpectedly: %q", got)
	}
}

func TestSubstituteSelfManagedSentinel_NonSentinelUntouched(t *testing.T) {
	s := &Server{cfg: &config.Config{Proxy: config.ProxyConfig{
		TokenSentinelFallback: true,
		VaultToken:            "real-vault-token",
	}}}
	r := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	r.Header.Set("Authorization", "Bearer some-real-client-token")
	r.RemoteAddr = "127.0.0.1:55555"
	if s.substituteSelfManagedSentinel(r) {
		t.Fatal("non-sentinel token must not be substituted")
	}
	if got := r.Header.Get("Authorization"); got != "Bearer some-real-client-token" {
		t.Errorf("header changed unexpectedly: %q", got)
	}
}

func TestSubstituteSelfManagedSentinel_AnthropicXAPIKey(t *testing.T) {
	s := &Server{cfg: &config.Config{Proxy: config.ProxyConfig{
		TokenSentinelFallback: true,
		VaultToken:            "real-vault-token",
	}}}
	r := httptest.NewRequest("POST", "/v1/messages", nil)
	r.Header.Set("x-api-key", "proxy-managed")
	r.RemoteAddr = "127.0.0.1:55555"
	if !s.substituteSelfManagedSentinel(r) {
		t.Fatal("expected x-api-key sentinel substitution on loopback")
	}
	if got := r.Header.Get("x-api-key"); got != "real-vault-token" {
		t.Errorf("x-api-key not rewritten: got %q", got)
	}
}

func TestSubstituteSelfManagedSentinel_NoVaultTokenConfigured(t *testing.T) {
	s := &Server{cfg: &config.Config{Proxy: config.ProxyConfig{
		TokenSentinelFallback: true,
		VaultToken:            "", // proxy has no vault token to substitute with
	}}}
	r := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	r.Header.Set("Authorization", "Bearer proxy-managed")
	r.RemoteAddr = "127.0.0.1:55555"
	if s.substituteSelfManagedSentinel(r) {
		t.Fatal("substitution must not happen when proxy has no VaultToken")
	}
}

// Compile-time: substituteSelfManagedSentinel signature stable.
var _ = func() bool {
	var s *Server
	return s.substituteSelfManagedSentinel(&http.Request{})
}
