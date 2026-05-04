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
