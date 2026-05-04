package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
)

// newTestServerWithBackend builds a Server with the minimum fields
// streamLocalService consults: serviceURLs[svc] points at the mock,
// localSems has a slot, pluginByID is nil-mapped (no plugin = no
// hub auth, no TLS internal CA, no PreserveModelID), serviceReasoning
// is empty (= reasoning off but no inline-tag triggers because we
// don't pass a qwen3 prefix in this test).
func newTestServerWithBackend(t *testing.T, svc, baseURL string) *Server {
	t.Helper()
	cfg := &config.Config{}
	s := &Server{
		cfg:              cfg,
		serviceURLs:      map[string]string{svc: baseURL},
		serviceReasoning: map[string]bool{},
		localSems:        map[string]chan struct{}{svc: make(chan struct{}, 1)},
		pluginByID:       map[string]*config.ServicePlugin{},
	}
	return s
}

func TestWriteOpenAIErrorChunk_EmitsChunkPlusDONE(t *testing.T) {
	var buf bytes.Buffer
	writeOpenAIErrorChunk(&buf, nil, errors.New("backend exploded"))

	got := buf.String()
	if !strings.Contains(got, "data: [DONE]\n\n") {
		t.Fatalf("missing DONE terminator: %q", got)
	}

	// First SSE event must be a parseable chat.completion.chunk
	// with the error string in choices[0].delta.content.
	parts := strings.SplitN(got, "\n\n", 2)
	if len(parts) < 2 {
		t.Fatalf("expected at least one event before DONE, got %q", got)
	}
	payload := strings.TrimPrefix(parts[0], "data: ")
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		t.Fatalf("first event not JSON: %v (payload=%q)", err, payload)
	}
	if chunk["object"] != "chat.completion.chunk" {
		t.Errorf("object=%q, want chat.completion.chunk", chunk["object"])
	}
	choices, _ := chunk["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("choices len=%d, want 1", len(choices))
	}
	c0 := choices[0].(map[string]interface{})
	delta := c0["delta"].(map[string]interface{})
	content, _ := delta["content"].(string)
	if !strings.Contains(content, "backend exploded") {
		t.Errorf("error string not in content: %q", content)
	}
	if c0["finish_reason"] != "stop" {
		t.Errorf("finish_reason=%v, want stop", c0["finish_reason"])
	}
}

func TestRewriteOpenAIChunkModel_RewritesParseableChunk(t *testing.T) {
	in := `data: {"id":"x","object":"chat.completion.chunk","model":"backend-internal-id","choices":[{"index":0,"delta":{"content":"hi"}}]}`
	out := rewriteOpenAIChunkModel(in, "qwen/qwen3.6-27b")
	if !strings.Contains(out, `"model":"qwen/qwen3.6-27b"`) {
		t.Errorf("model not rewritten: %q", out)
	}
	if !strings.Contains(out, `"content":"hi"`) {
		t.Errorf("other fields lost: %q", out)
	}
}

func TestRewriteOpenAIChunkModel_PassesThroughDoneAndUnparseable(t *testing.T) {
	cases := []string{
		`data: [DONE]`,
		`data: not-json-at-all`,
		``,            // empty line
		`: keepalive`, // SSE comment
		`event: ping`, // non-data SSE field
	}
	for _, in := range cases {
		out := rewriteOpenAIChunkModel(in, "anything")
		if out != in {
			t.Errorf("verbatim pass-through expected for %q, got %q", in, out)
		}
	}
}

func TestRewriteOpenAIChunkModel_RewritesGenericObjectsWithObjectKey(t *testing.T) {
	in := `data: {"object":"some.other.thing","model":"x"}`
	out := rewriteOpenAIChunkModel(in, "qwen")
	// Non chat.completion.chunk objects should also be rewritten — we
	// rewrite any payload that has a top-level "object" key. If a
	// future-introduced tighter check is wanted, narrow there.
	if !strings.Contains(out, `"model":"qwen"`) {
		t.Errorf("model not rewritten on generic object: %q", out)
	}
}

func TestStreamLocalService_HappyPath_PipesChunksAndRewritesModel(t *testing.T) {
	// Mock backend that emits 3 content chunks + DONE in OpenAI SSE shape.
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the proxy sent stream:true to the backend.
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode req body: %v", err)
		}
		if body["stream"] != true {
			t.Errorf("backend received stream=%v, want true", body["stream"])
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f, _ := w.(http.Flusher)
		for _, c := range []string{
			`data: {"id":"x","object":"chat.completion.chunk","model":"backend-id","choices":[{"index":0,"delta":{"role":"assistant"}}]}`,
			`data: {"id":"x","object":"chat.completion.chunk","model":"backend-id","choices":[{"index":0,"delta":{"content":"Hello"}}]}`,
			`data: {"id":"x","object":"chat.completion.chunk","model":"backend-id","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		} {
			fmt.Fprintln(w, c)
			fmt.Fprintln(w) // SSE event delimiter
			if f != nil {
				f.Flush()
			}
		}
	}))
	defer mock.Close()

	s := newTestServerWithBackend(t, "lmstudio", mock.URL)

	// Caller-side ResponseWriter
	rec := httptest.NewRecorder()
	flusher := http.Flusher(nil) // recorder isn't a Flusher — that's fine for this test
	oaiReq := &OpenAIRequest{
		Model:    "qwen/qwen3.6-27b",
		Stream:   true,
		Messages: []OpenAIMessage{{Role: "user", Content: "say hi"}},
	}
	err := s.streamLocalService(context.Background(), rec, flusher, "lmstudio", "qwen/qwen3.6-27b", oaiReq)
	if err != nil {
		t.Fatalf("streamLocalService: %v", err)
	}

	got := rec.Body.String()
	if !strings.Contains(got, `"content":"Hello"`) {
		t.Errorf("first content chunk missing: %q", got)
	}
	if !strings.Contains(got, `"content":" world"`) {
		t.Errorf("second content chunk missing: %q", got)
	}
	if !strings.Contains(got, `"model":"qwen/qwen3.6-27b"`) {
		t.Errorf("model rewrite missing — caller should see resolved id, not backend id: %q", got)
	}
	if strings.Contains(got, `"model":"backend-id"`) {
		t.Errorf("backend-id leaked through: %q", got)
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "data: [DONE]") {
		t.Errorf("missing trailing DONE: %q", got)
	}
}

func TestStreamLocalService_Non200_ReturnsError(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"bad model"}`)
	}))
	defer mock.Close()

	s := newTestServerWithBackend(t, "lmstudio", mock.URL)
	rec := httptest.NewRecorder()
	oaiReq := &OpenAIRequest{
		Model:    "anything",
		Stream:   true,
		Messages: []OpenAIMessage{{Role: "user", Content: "x"}},
	}
	err := s.streamLocalService(context.Background(), rec, nil, "lmstudio", "anything", oaiReq)
	if err == nil {
		t.Fatal("expected error for backend 400, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should mention status 400: %v", err)
	}
	if !strings.Contains(err.Error(), "bad model") {
		t.Errorf("error should include backend body: %v", err)
	}

	// The function must NOT have written anything to the caller's
	// ResponseWriter before the failure was detected — the caller
	// expects to handle the error itself (typically via writeOpenAIErrorChunk).
	if rec.Body.Len() != 0 {
		t.Errorf("expected no body writes on pre-200 failure, got %q", rec.Body.String())
	}
}

func TestStreamLocalService_MidStreamAbort_EmitsSyntheticFinishPlusDONE(t *testing.T) {
	// Simulate backend that emits one chunk then silently closes the stream
	// (no hijack needed; just close the response body early).
	// The scanner will see EOF as the end of valid events, and since
	// we never sent [DONE], the function should emit synthetic finish + DONE.
	responseChan := make(chan string, 10)
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f, _ := w.(http.Flusher)
		// Write one chunk.
		fmt.Fprintln(w, `data: {"id":"x","object":"chat.completion.chunk","model":"backend-id","choices":[{"index":0,"delta":{"content":"partial"}}]}`)
		fmt.Fprintln(w)
		if f != nil {
			f.Flush()
		}
		// Close the body by ending the handler without sending [DONE].
		// This causes the scanner on the client side to receive EOF.
	}))
	defer mock.Close()

	s := newTestServerWithBackend(t, "lmstudio", mock.URL)
	rec := httptest.NewRecorder()
	oaiReq := &OpenAIRequest{
		Model:    "qwen/qwen3.6-27b",
		Stream:   true,
		Messages: []OpenAIMessage{{Role: "user", Content: "x"}},
	}
	err := s.streamLocalService(context.Background(), rec, nil, "lmstudio", "qwen/qwen3.6-27b", oaiReq)
	// Mid-stream close-without-DONE: the scanner hits EOF, which is not an
	// error condition (sawDone==false triggers synthetic finish logic).
	// The function emits a synthetic finish_reason="stop" chunk + DONE,
	// then returns nil.
	if err != nil {
		t.Fatalf("streamLocalService unexpected error: %v", err)
	}

	got := rec.Body.String()
	if !strings.Contains(got, `"content":"partial"`) {
		t.Errorf("partial content lost: %q", got)
	}
	if !strings.Contains(got, `"finish_reason":"stop"`) {
		t.Errorf("synthetic finish chunk missing: %q", got)
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "data: [DONE]") {
		t.Errorf("missing DONE: %q", got)
	}
	_ = responseChan // suppress unused warning
}

// Header-idempotence: the handler is allowed to set SSE headers
// before calling streamLocalService; the function must not emit
// duplicate Content-Type entries.
func TestStreamLocalService_HeaderIdempotence(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f, _ := w.(http.Flusher)
		fmt.Fprintln(w, `data: [DONE]`)
		fmt.Fprintln(w)
		if f != nil {
			f.Flush()
		}
	}))
	defer mock.Close()

	s := newTestServerWithBackend(t, "lmstudio", mock.URL)
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "text/event-stream") // pre-set by hypothetical handler
	rec.Header().Set("Cache-Control", "no-cache")

	oaiReq := &OpenAIRequest{
		Model: "x", Stream: true,
		Messages: []OpenAIMessage{{Role: "user", Content: "x"}},
	}
	if err := s.streamLocalService(context.Background(), rec, nil, "lmstudio", "x", oaiReq); err != nil {
		t.Fatalf("streamLocalService: %v", err)
	}
	got := rec.Header().Values("Content-Type")
	if len(got) != 1 {
		t.Errorf("Content-Type set %d times, want 1: %v", len(got), got)
	}
}

func TestStreamLocalService_CallerDisconnect_ReturnsNil(t *testing.T) {
	// Backend sends chunks continuously, blocking with no final [DONE].
	// The caller context is cancelled while streamLocalService is blocked
	// reading. This causes scanner.Err() to return an error, and since
	// ctx.Err() != nil, the function returns nil (caller disconnect).
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f, _ := w.(http.Flusher)
		// Send one chunk.
		fmt.Fprintln(w, `data: {"id":"x","object":"chat.completion.chunk","model":"backend-id","choices":[{"index":0,"delta":{"content":"first"}}]}`)
		fmt.Fprintln(w)
		if f != nil {
			f.Flush()
		}
		// Block until the request context is cancelled.
		<-r.Context().Done()
	}))
	defer mock.Close()

	s := newTestServerWithBackend(t, "lmstudio", mock.URL)
	rec := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	oaiReq := &OpenAIRequest{
		Model:    "qwen/qwen3.6-27b",
		Stream:   true,
		Messages: []OpenAIMessage{{Role: "user", Content: "x"}},
	}

	// Cancel after a delay long enough for the HTTP handshake and first chunk
	// to be processed. streamLocalService will block in scanner.Scan()
	// waiting for the next chunk, which never comes because the context
	// is cancelled.
	go func() {
		// This delay is long enough that by the time we cancel, the function
		// has definitely started the scanner loop and is reading chunks.
		select {
		case <-time.After(100 * time.Millisecond):
			cancel()
		case <-ctx.Done():
		}
	}()

	err := s.streamLocalService(ctx, rec, nil, "lmstudio", "qwen/qwen3.6-27b", oaiReq)

	// When the request context is cancelled mid-stream, scanner.Err()
	// will be non-nil, but ctx.Err() will also be non-nil, so the
	// function returns nil (line 541-542 of stream.go).
	if err != nil {
		t.Fatalf("caller disconnect should return nil, got %v", err)
	}
	if !strings.Contains(rec.Body.String(), `"content":"first"`) {
		t.Errorf("first chunk should have been delivered before cancel: %q", rec.Body.String())
	}
}
