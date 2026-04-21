package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
)

// newBYOServer sets up a minimal proxy.Server suitable for exercising the BYO
// branch of callAnthropicPassthrough. Only the config fields doRequest touches
// are populated — vault key / rotation paths must not be reachable via BYO.
func newBYOServer() *Server {
	cfg := &config.Config{}
	cfg.Proxy.Timeout = 5 * time.Second
	return &Server{cfg: cfg}
}

// TestAnthropicPassthrough_BYOBearer verifies that when a caller supplies a
// Claude-Code-style OAuth token via Authorization: Bearer sk-ant-oat..., the
// passthrough forwards it verbatim and does NOT overwrite it with a vault
// x-api-key. Regression guard for the Anthropic-credit-exhaustion hang.
func TestAnthropicPassthrough_BYOBearer(t *testing.T) {
	var capturedAuth, capturedXKey, capturedVersion string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		capturedXKey = r.Header.Get("x-api-key")
		capturedVersion = r.Header.Get("anthropic-version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg_x","type":"message","content":[{"type":"text","text":"ok"}]}`)
	}))
	defer upstream.Close()

	prev := anthropicMessagesURL
	anthropicMessagesURL = upstream.URL
	defer func() { anthropicMessagesURL = prev }()

	s := newBYOServer()
	req := &AnthropicRequest{
		Model:    "claude-opus-4-7",
		Messages: []AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hi"`)}},
	}
	body, ct, key, err := s.callAnthropicPassthrough(
		context.Background(), req, "claude-opus-4-7", "", "sk-ant-oat-test-12345",
	)
	if err != nil {
		t.Fatalf("BYO bearer passthrough: %v", err)
	}
	if key != nil {
		t.Fatal("BYO path must return nil vault key")
	}
	if capturedAuth != "Bearer sk-ant-oat-test-12345" {
		t.Errorf("Authorization header = %q, want Bearer sk-ant-oat-test-12345", capturedAuth)
	}
	if capturedXKey != "" {
		t.Errorf("x-api-key must NOT be injected on BYO bearer path, got %q", capturedXKey)
	}
	if capturedVersion != "2023-06-01" {
		t.Errorf("anthropic-version header = %q, want 2023-06-01", capturedVersion)
	}
	if ct != "application/json" {
		t.Errorf("Content-Type returned = %q, want application/json", ct)
	}
	if !strings.Contains(string(body), "ok") {
		t.Errorf("response body missing upstream content: %s", body)
	}
}

// TestAnthropicPassthrough_BYOAPIKey verifies that a caller-supplied personal
// API key on the x-api-key header is forwarded as-is and no Authorization
// header is invented.
func TestAnthropicPassthrough_BYOAPIKey(t *testing.T) {
	var capturedAuth, capturedXKey string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		capturedXKey = r.Header.Get("x-api-key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg_x","type":"message","content":[{"type":"text","text":"ok"}]}`)
	}))
	defer upstream.Close()

	prev := anthropicMessagesURL
	anthropicMessagesURL = upstream.URL
	defer func() { anthropicMessagesURL = prev }()

	s := newBYOServer()
	req := &AnthropicRequest{
		Model:    "claude-sonnet-4-6",
		Messages: []AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hi"`)}},
	}
	_, _, key, err := s.callAnthropicPassthrough(
		context.Background(), req, "claude-sonnet-4-6", "sk-ant-api03-byo-fake", "",
	)
	if err != nil {
		t.Fatalf("BYO api-key passthrough: %v", err)
	}
	if key != nil {
		t.Fatal("BYO path must return nil vault key")
	}
	if capturedXKey != "sk-ant-api03-byo-fake" {
		t.Errorf("x-api-key forwarded = %q, want sk-ant-api03-byo-fake", capturedXKey)
	}
	if capturedAuth != "" {
		t.Errorf("Authorization must NOT be set on BYO api-key path, got %q", capturedAuth)
	}
}

// TestAnthropicPassthrough_BYO_UpstreamContentTypePreserved verifies that the
// passthrough returns whatever Content-Type the upstream used (e.g. a stream=
// true request will come back as text/event-stream, not JSON).
func TestAnthropicPassthrough_BYO_UpstreamContentTypePreserved(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "event: message_start\ndata: {}\n\n")
	}))
	defer upstream.Close()

	prev := anthropicMessagesURL
	anthropicMessagesURL = upstream.URL
	defer func() { anthropicMessagesURL = prev }()

	s := newBYOServer()
	req := &AnthropicRequest{
		Model:    "claude-opus-4-7",
		Stream:   true,
		Messages: []AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hi"`)}},
	}
	_, ct, _, err := s.callAnthropicPassthrough(
		context.Background(), req, "claude-opus-4-7", "", "sk-ant-oat-stream",
	)
	if err != nil {
		t.Fatalf("BYO stream passthrough: %v", err)
	}
	if ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

// TestAnthropicPassthrough_BYO_UpstreamErrorPropagated verifies that a non-200
// upstream response surfaces as an error so handleAnthropic falls back to
// dispatch instead of happily writing the error body as a success payload.
func TestAnthropicPassthrough_BYO_UpstreamErrorPropagated(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"type":"authentication_error","message":"invalid bearer"}}`)
	}))
	defer upstream.Close()

	prev := anthropicMessagesURL
	anthropicMessagesURL = upstream.URL
	defer func() { anthropicMessagesURL = prev }()

	s := newBYOServer()
	req := &AnthropicRequest{
		Model:    "claude-opus-4-7",
		Messages: []AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hi"`)}},
	}
	body, _, _, err := s.callAnthropicPassthrough(
		context.Background(), req, "claude-opus-4-7", "", "sk-ant-oat-bad",
	)
	if err == nil {
		t.Fatal("upstream 401 must surface as error")
	}
	if body != nil {
		t.Errorf("body must be nil on error, got %q", body)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error must mention HTTP 401, got %v", err)
	}
}
