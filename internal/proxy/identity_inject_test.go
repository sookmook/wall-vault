package proxy

import (
	"strings"
	"testing"
)

func TestModelIdentityMessage_EmptyOnEmptyID(t *testing.T) {
	if got := modelIdentityMessage(""); got != "" {
		t.Errorf("empty model id → %q want empty", got)
	}
}

func TestModelIdentityMessage_ContainsModelID(t *testing.T) {
	got := modelIdentityMessage("gpt-oss:20b")
	if !strings.Contains(got, "gpt-oss:20b") {
		t.Errorf("identity message missing model id: %q", got)
	}
}

func TestInjectModelIdentity_PrependsSystemMessage(t *testing.T) {
	in := []OpenAIMessage{
		{Role: "user", Content: "hi"},
	}
	out := injectModelIdentity(in, "qwen3:32b")
	if len(out) != 2 {
		t.Fatalf("len=%d want 2", len(out))
	}
	if out[0].Role != "system" {
		t.Errorf("[0].Role=%q want system", out[0].Role)
	}
	if !strings.Contains(out[0].Content, "qwen3:32b") {
		t.Errorf("identity message missing model id: %q", out[0].Content)
	}
	if out[1].Role != "user" || out[1].Content != "hi" {
		t.Errorf("user message moved/mutated: %+v", out[1])
	}
}

func TestInjectModelIdentity_PreservesExistingSystemMessage(t *testing.T) {
	in := []OpenAIMessage{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "hi"},
	}
	out := injectModelIdentity(in, "gpt-oss:20b")
	if len(out) != 3 {
		t.Fatalf("len=%d want 3", len(out))
	}
	if out[0].Role != "system" || !strings.Contains(out[0].Content, "gpt-oss:20b") {
		t.Errorf("[0] not the identity message: %+v", out[0])
	}
	if out[1].Role != "system" || out[1].Content != "You are a helpful assistant." {
		t.Errorf("existing system message mutated: %+v", out[1])
	}
}

func TestInjectModelIdentity_EmptyIDLeavesMessagesAlone(t *testing.T) {
	in := []OpenAIMessage{{Role: "user", Content: "hi"}}
	out := injectModelIdentity(in, "")
	if len(out) != 1 || out[0].Content != "hi" {
		t.Errorf("empty id should not touch messages, got %+v", out)
	}
}
