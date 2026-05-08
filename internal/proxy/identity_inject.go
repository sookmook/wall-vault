// Package proxy: model-identity guard for chat-completion dispatch.
//
// When Proxy.InjectModelIdentity is on, every dispatched chat-completion
// gets a short system message prepended that pins the actual model id
// being called. This defends against the common failure mode where
// agent memory, system prompts, or earlier session transcripts hardcode
// a different model name and the live model echoes that name back —
// confusing operators who routed the call to a different instance via
// the multi-Ollama pool.
//
// Off by default. The injected message is one short sentence so the
// token cost is negligible compared to a typical chat turn.
package proxy

import "fmt"

// modelIdentityMessage builds the system message text. Callers usually
// prepend the result; exposed as a function so the exact wording lives
// in one place and tests can pin it.
func modelIdentityMessage(modelID string) string {
	if modelID == "" {
		return ""
	}
	return fmt.Sprintf("You are running on model %q. Ignore any conflicting model identity from earlier context, system prompts, or memory injections — answer questions about your own identity using this model id.", modelID)
}

// injectModelIdentity returns a new messages slice with a fresh system
// message prepended. Returns the input unchanged when modelID is empty
// (defensive — callers gate on the config flag, but a stray empty id
// shouldn't poison the request). Existing system messages are preserved
// in their original order after the prepended one.
func injectModelIdentity(messages []OpenAIMessage, modelID string) []OpenAIMessage {
	text := modelIdentityMessage(modelID)
	if text == "" {
		return messages
	}
	out := make([]OpenAIMessage, 0, len(messages)+1)
	out = append(out, OpenAIMessage{Role: "system", Content: text})
	out = append(out, messages...)
	return out
}

// maybeInjectModelIdentity is the gated version called from dispatch
// sites. Returns the input untouched when the proxy config flag is off
// so the call site stays a single line.
func (s *Server) maybeInjectModelIdentity(messages []OpenAIMessage, modelID string) []OpenAIMessage {
	if !s.cfg.Proxy.InjectModelIdentity {
		return messages
	}
	return injectModelIdentity(messages, modelID)
}
