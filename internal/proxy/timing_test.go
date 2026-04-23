package proxy

import (
	"testing"
	"time"
)

func TestAgentOffset_Determinism(t *testing.T) {
	const maxMs = 500
	want := AgentOffset("agent-a", maxMs)
	for i := 0; i < 100; i++ {
		if got := AgentOffset("agent-a", maxMs); got != want {
			t.Fatalf("iteration %d: got %v, want %v", i, got, want)
		}
	}
}

func TestAgentOffset_Range(t *testing.T) {
	const maxMs = 500
	for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
		got := AgentOffset(id, maxMs)
		if got < 0 || got >= time.Duration(maxMs)*time.Millisecond {
			t.Fatalf("id=%q: offset %v out of [0, %dms)", id, got, maxMs)
		}
	}
}

func TestAgentOffset_BucketDiversity(t *testing.T) {
	const maxMs = 500
	ids := []string{"a", "b", "c", "d"}
	seen := make(map[time.Duration]string, len(ids))
	for _, id := range ids {
		off := AgentOffset(id, maxMs)
		if prev, ok := seen[off]; ok {
			t.Fatalf("ids %q and %q collide on offset %v", prev, id, off)
		}
		seen[off] = id
	}
}

func TestAgentOffset_EmptyID(t *testing.T) {
	if got := AgentOffset("", 500); got != 0 {
		t.Fatalf("empty id: got %v, want 0", got)
	}
}

func TestAgentOffset_ZeroMax(t *testing.T) {
	if got := AgentOffset("agent-a", 0); got != 0 {
		t.Fatalf("zero max: got %v, want 0", got)
	}
	if got := AgentOffset("agent-a", -10); got != 0 {
		t.Fatalf("negative max: got %v, want 0", got)
	}
}

func TestFallbackJitter_Range(t *testing.T) {
	const maxMs = 200
	for i := 0; i < 1000; i++ {
		got := FallbackJitter(maxMs)
		if got < 0 || got >= time.Duration(maxMs)*time.Millisecond {
			t.Fatalf("iter %d: jitter %v out of [0, %dms)", i, got, maxMs)
		}
	}
}

func TestFallbackJitter_NonConstant(t *testing.T) {
	seen := make(map[time.Duration]struct{})
	for i := 0; i < 1000; i++ {
		seen[FallbackJitter(200)] = struct{}{}
	}
	if len(seen) < 100 {
		t.Fatalf("only %d distinct jitter values out of 1000 — too clumpy", len(seen))
	}
}

func TestFallbackJitter_ZeroMax(t *testing.T) {
	if got := FallbackJitter(0); got != 0 {
		t.Fatalf("zero max: got %v, want 0", got)
	}
	if got := FallbackJitter(-1); got != 0 {
		t.Fatalf("negative max: got %v, want 0", got)
	}
}
