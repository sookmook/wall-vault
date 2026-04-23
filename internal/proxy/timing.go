package proxy

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"
)

// Local-inference timing budgets, shared across all local backends
// (ollama / llamacpp / lmstudio / vllm). Four nodes with 500ms bucket
// spacing average 125ms apart; 200ms of additive jitter smooths
// residual collisions.
const (
	localAgentOffsetMs    = 500
	localFallbackJitterMs = 200
)

// AgentOffset returns a deterministic per-agent delay in [0, maxMs) ms
// derived from clientID. Same clientID always maps to the same offset,
// so two fleet members land on different phase positions structurally.
// Empty clientID (standalone mode) or non-positive maxMs → 0 duration.
func AgentOffset(clientID string, maxMs int) time.Duration {
	if clientID == "" || maxMs <= 0 {
		return 0
	}
	h := sha256.Sum256([]byte(clientID))
	n := binary.BigEndian.Uint64(h[:8]) % uint64(maxMs)
	return time.Duration(n) * time.Millisecond
}

// FallbackJitter returns a uniform random delay in [0, maxMs) ms. Used
// as an additive component on top of AgentOffset so even agents whose
// hash buckets happen to collide still diverge across successive
// requests. Non-positive maxMs or rand failure → 0 duration.
func FallbackJitter(maxMs int) time.Duration {
	if maxMs <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxMs)))
	if err != nil {
		return 0
	}
	return time.Duration(n.Int64()) * time.Millisecond
}
