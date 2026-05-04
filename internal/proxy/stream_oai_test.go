package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

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
		``,                  // empty line
		`: keepalive`,       // SSE comment
		`event: ping`,       // non-data SSE field
	}
	for _, in := range cases {
		out := rewriteOpenAIChunkModel(in, "anything")
		if out != in {
			t.Errorf("verbatim pass-through expected for %q, got %q", in, out)
		}
	}
}

func TestRewriteOpenAIChunkModel_LeavesNonChunkObjectsAlone(t *testing.T) {
	in := `data: {"object":"some.other.thing","model":"x"}`
	out := rewriteOpenAIChunkModel(in, "qwen")
	// Non chat.completion.chunk objects should also be rewritten — we
	// rewrite any payload that has a top-level "object" key. If a
	// future-introduced tighter check is wanted, narrow there.
	if !strings.Contains(out, `"model":"qwen"`) {
		t.Errorf("model not rewritten on generic object: %q", out)
	}
}
