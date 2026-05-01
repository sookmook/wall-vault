package proxy

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestOpenAIRequestOllamaFieldsOmittedWhenNil guarantees that the new
// Ollama-only knobs (KeepAlive + Options) stay out of the wire format when
// the proxy talks to non-Ollama backends. OpenAI / Anthropic / Google would
// reject unknown top-level fields, so emitting them unconditionally would
// regress every cloud call.
func TestOpenAIRequestOllamaFieldsOmittedWhenNil(t *testing.T) {
	req := &OpenAIRequest{
		Model:    "test",
		Messages: []OpenAIMessage{{Role: "user", Content: "hi"}},
	}
	out, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(out)
	for _, banned := range []string{"keep_alive", "options", "num_ctx"} {
		if strings.Contains(s, banned) {
			t.Errorf("nil OllamaOptions/KeepAlive should not emit %q; got %s", banned, s)
		}
	}
}

// TestOpenAIRequestOllamaFieldsEmittedWhenSet checks that operators get the
// keep_alive + num_ctx values they configured all the way to the wire.
// Without this, callOllama could silently drop the fields and we'd debug
// "why isn't keep_alive working" by hand.
func TestOpenAIRequestOllamaFieldsEmittedWhenSet(t *testing.T) {
	ka := "30m"
	nc := 8192
	req := &OpenAIRequest{
		Model:     "test",
		Messages:  []OpenAIMessage{{Role: "user", Content: "hi"}},
		KeepAlive: &ka,
		Options:   &OllamaOptions{NumCtx: &nc},
	}
	out, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `"keep_alive":"30m"`) {
		t.Errorf("missing keep_alive in %s", s)
	}
	if !strings.Contains(s, `"num_ctx":8192`) {
		t.Errorf("missing num_ctx in %s", s)
	}
}

// TestOllamaOptionsNumCtxOmitWhenZero pins down the pointer-vs-omitempty
// invariant for NumCtx. Plain `int` with omitempty would drop legitimate
// zero values, but NumCtx=0 doesn't make sense — pointer makes the
// "unset / leave default" intent explicit.
func TestOllamaOptionsNumCtxOmitWhenZero(t *testing.T) {
	opts := &OllamaOptions{Temperature: 0.7}
	out, _ := json.Marshal(opts)
	if strings.Contains(string(out), "num_ctx") {
		t.Errorf("nil NumCtx leaked into %s", out)
	}
}
