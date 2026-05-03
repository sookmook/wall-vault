package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/hooks"
	"github.com/sookmook/wall-vault/internal/models"
)

// captureRequestServer responds with a tiny OpenAI-compat completion and
// records the headers/path the request arrived with. Lets the test verify
// that auth/TLS settings from a plugin yaml actually shape the outbound
// HTTP call.
type captureRequestServer struct {
	mu      sync.Mutex
	headers http.Header
	path    string
	body    []byte
}

func (c *captureRequestServer) handler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	c.headers = r.Header.Clone()
	c.path = r.URL.Path
	c.body, _ = io.ReadAll(r.Body)
	c.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"id":      "test-1",
		"object":  "chat.completion",
		"choices": []map[string]any{{"index": 0, "message": map[string]string{"role": "assistant", "content": "ok"}, "finish_reason": "stop"}},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// minimalServer builds the smallest *Server that callLocalService needs.
// Avoids NewServer to keep the test free of real key managers, vault SSE,
// systemd glue, etc. Every field touched by callLocalService is set here.
func minimalServer(t *testing.T, plugin *config.ServicePlugin) *Server {
	t.Helper()
	cfg := &config.Config{}
	cfg.Proxy.ClientID = "test"
	cfg.Proxy.VaultToken = "test-vault-token"
	cfg.Proxy.Services = []string{"lmstudio"}
	if plugin != nil {
		cfg.Plugins = []config.ServicePlugin{*plugin}
	}
	s := &Server{
		cfg:              cfg,
		registry:         models.NewRegistry(time.Minute),
		hooksMgr:         hooks.NewManager(nil, ""),
		serviceURLs:      map[string]string{},
		serviceDefaults:  map[string]string{},
		serviceReasoning: map[string]bool{},
		localSems:        map[string]chan struct{}{"lmstudio": make(chan struct{}, 1)},
		stopCh:           make(chan struct{}),
	}
	if plugin != nil {
		s.pluginByID = map[string]*config.ServicePlugin{plugin.ID: &cfg.Plugins[0]}
	}
	return s
}

func TestCallLocalService_BearerAuthFromPlugin(t *testing.T) {
	cap := &captureRequestServer{}
	srv := httptest.NewServer(http.HandlerFunc(cap.handler))
	defer srv.Close()

	plugin := &config.ServicePlugin{
		ID:         "lmstudio",
		Enabled:    true,
		DefaultURL: srv.URL,
		Auth:       config.ServiceAuth{Type: "bearer"},
	}
	s := minimalServer(t, plugin)

	req := &GeminiRequest{}
	if _, err := s.callLocalService(context.Background(), "lmstudio", "qwen3.6:27b", req); err != nil {
		t.Fatalf("callLocalService: %v", err)
	}
	got := cap.headers.Get("Authorization")
	if got != "Bearer test-vault-token" {
		t.Errorf("Authorization = %q, want Bearer test-vault-token", got)
	}
	if !strings.HasSuffix(cap.path, "/v1/chat/completions") {
		t.Errorf("path = %q, want /v1/chat/completions suffix", cap.path)
	}
}

func TestCallLocalService_NoAuthByDefault(t *testing.T) {
	cap := &captureRequestServer{}
	srv := httptest.NewServer(http.HandlerFunc(cap.handler))
	defer srv.Close()

	plugin := &config.ServicePlugin{
		ID:         "lmstudio",
		Enabled:    true,
		DefaultURL: srv.URL,
		// Auth omitted → none
	}
	s := minimalServer(t, plugin)

	req := &GeminiRequest{}
	if _, err := s.callLocalService(context.Background(), "lmstudio", "x", req); err != nil {
		t.Fatalf("callLocalService: %v", err)
	}
	if got := cap.headers.Get("Authorization"); got != "" {
		t.Errorf("expected no Authorization header, got %q", got)
	}
}

func TestCallLocalService_DefaultURLFallback(t *testing.T) {
	cap := &captureRequestServer{}
	srv := httptest.NewServer(http.HandlerFunc(cap.handler))
	defer srv.Close()

	// No plugin URL via serviceURLs map; only via plugin.DefaultURL.
	plugin := &config.ServicePlugin{
		ID:         "lmstudio",
		Enabled:    true,
		DefaultURL: srv.URL,
	}
	s := minimalServer(t, plugin)
	// serviceURLs is intentionally empty so the dispatcher has to fall
	// through to plugin.DefaultURL — this is the path a fresh raspi takes
	// before the SSE config_change has populated serviceURLs.

	req := &GeminiRequest{}
	if _, err := s.callLocalService(context.Background(), "lmstudio", "x", req); err != nil {
		t.Fatalf("callLocalService: %v", err)
	}
	if cap.path == "" {
		t.Fatal("expected request to reach test server via plugin DefaultURL")
	}
}

func TestCallLocalService_PluginURLBeatsVaultSSE(t *testing.T) {
	// The plugin yaml is the operator's explicit override. When both a
	// vault-distributed serviceURLs entry AND a plugin DefaultURL exist,
	// the plugin must win. Without this, the hub-topology pattern (a
	// client wall-vault forwarding to a remote hub wall-vault) is dead
	// on arrival because vault's default URL — pointing at the hub's
	// localhost LLM, unreachable from the client — silently masks the
	// plugin's hub URL.
	cap := &captureRequestServer{}
	srv := httptest.NewServer(http.HandlerFunc(cap.handler))
	defer srv.Close()

	plugin := &config.ServicePlugin{
		ID:         "lmstudio",
		Enabled:    true,
		DefaultURL: srv.URL, // plugin says: dispatch to test server
	}
	s := minimalServer(t, plugin)
	// Vault SSE pretends to know a (deliberately wrong) URL — should be
	// ignored because the plugin override is present.
	s.serviceURLs["lmstudio"] = "http://192.0.2.1:9999"

	req := &GeminiRequest{}
	if _, err := s.callLocalService(context.Background(), "lmstudio", "x", req); err != nil {
		t.Fatalf("callLocalService: %v — should have used plugin URL, not vault SSE", err)
	}
	if cap.path == "" {
		t.Fatal("plugin DefaultURL did not win over serviceURLs[id]")
	}
}

func TestCallLocalService_LegacyNoPlugin(t *testing.T) {
	// No plugin at all — verify the pre-v0.2.44 behaviour still works:
	// serviceURLs[svc] picks the URL, no auth header, default http client.
	cap := &captureRequestServer{}
	srv := httptest.NewServer(http.HandlerFunc(cap.handler))
	defer srv.Close()

	s := minimalServer(t, nil)
	s.serviceURLs["lmstudio"] = srv.URL

	req := &GeminiRequest{}
	if _, err := s.callLocalService(context.Background(), "lmstudio", "x", req); err != nil {
		t.Fatalf("callLocalService: %v", err)
	}
	if got := cap.headers.Get("Authorization"); got != "" {
		t.Errorf("legacy path should not set Authorization, got %q", got)
	}
}
