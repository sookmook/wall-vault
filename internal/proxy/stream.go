package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ─── Gemini Streaming Handler ────────────────────────────────────────────────

// handleGeminiStream: handle the streamGenerateContent endpoint
func (s *Server) handleGeminiStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireProxyToken(w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAIBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "body read error", http.StatusBadRequest)
		return
	}

	var req GeminiRequest
	if err := json.Unmarshal(body, &req); err != nil {
		jsonError(w, "invalid gemini request", http.StatusBadRequest)
		return
	}

	stripped := s.filter.FilterGemini(&req)
	if stripped > 0 {
		log.Printf("[Security] blocked %d tools from stream request", stripped)
	}

	s.mu.RLock()
	svc := s.service
	mdl := s.model
	s.mu.RUnlock()

	if urlModel := extractModelFromPath(r.URL.Path); urlModel != "" {
		if strings.HasPrefix(urlModel, "gemini-") || strings.HasPrefix(urlModel, "gemma-") {
			svc = "google"
			mdl = urlModel
		}
	}

	// set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, _ := w.(http.Flusher)

	switch svc {
	case "google":
		s.streamGoogle(w, flusher, mdl, &req)
	case "openrouter":
		s.streamOpenRouter(w, flusher, mdl, &req)
	case "ollama":
		s.streamOllama(r.Context(), w, flusher, mdl, &req)
	default:
		// Plugin-defined OAI-compat backends (lmstudio / vllm / llamacpp /
		// jan / kobold / tabbyapi / etc.) get streamed via the same
		// OAI-SSE→Gemini-SSE bridge as openrouter, so a Gemini-style
		// caller (streamGenerateContent) can reach any OAI-compat plugin
		// without needing a Go-side switch case per backend.
		if plugin := s.pluginByID[svc]; plugin != nil {
			switch plugin.RequestFormat {
			case "", "openai":
				s.streamPluginAsGemini(r.Context(), w, flusher, svc, mdl, &req)
				return
			}
		}
		writeGeminiErrorChunk(w, flusher, fmt.Errorf("서비스 미지원: %s", svc))
	}
}

// ─── Google Streaming Pass-Through ───────────────────────────────────────────

func (s *Server) streamGoogle(w http.ResponseWriter, f http.Flusher, model string, req *GeminiRequest) {
	key, plainKey, err := s.getKey("google")
	if err != nil {
		writeGeminiErrorChunk(w, f, err)
		return
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", model, plainKey)
	data, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: s.cfg.Proxy.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		s.keyMgr.RecordError(key, 0)
		writeGeminiErrorChunk(w, f, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.keyMgr.RecordError(key, resp.StatusCode)
		writeGeminiErrorChunk(w, f, fmt.Errorf("Google API 오류: HTTP %d", resp.StatusCode))
		return
	}

	// Google SSE pass-through: forward as-is to client
	tokens := 0
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		if f != nil {
			f.Flush()
		}
		// estimate token count (parse usageMetadata)
		if strings.HasPrefix(line, "data: ") {
			var chunk GeminiResponse
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chunk); err == nil {
				if chunk.UsageMetadata != nil {
					tokens = chunk.UsageMetadata.TotalTokenCount
				}
			}
		}
	}
	if tokens == 0 {
		tokens = 1 // minimum 1 per request when API does not report usage
	}
	s.keyMgr.RecordSuccess(key, tokens)
}

// ─── OpenRouter Streaming Conversion ─────────────────────────────────────────

func (s *Server) streamOpenRouter(w http.ResponseWriter, f http.Flusher, model string, req *GeminiRequest) {
	key, plainKey, err := s.getKey("openrouter")
	if err != nil {
		writeGeminiErrorChunk(w, f, err)
		return
	}

	oaiReq := GeminiToOpenAI(model, req)
	oaiReq.Stream = true
	data, _ := json.Marshal(oaiReq)

	httpReq, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+plainKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/sookmook/wall-vault")
	httpReq.Header.Set("X-Title", "wall-vault")

	client := &http.Client{Timeout: s.cfg.Proxy.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		s.keyMgr.RecordError(key, 0)
		writeGeminiErrorChunk(w, f, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.keyMgr.RecordError(key, resp.StatusCode)
		writeGeminiErrorChunk(w, f, fmt.Errorf("OpenRouter 오류: HTTP %d", resp.StatusCode))
		return
	}

	// convert OpenAI SSE → Gemini SSE
	totalTokens := 0
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var chunk OpenAIResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		// capture token usage from any chunk — the final usage chunk often has
		// empty choices (choices:[]) or empty delta content, so check usage first
		if chunk.Usage != nil && chunk.Usage.TotalTokens > 0 {
			totalTokens = chunk.Usage.TotalTokens
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		text := ""
		if delta != nil {
			text = delta.Content
			// Reasoning-model fallback (see OpenAIRespToGemini): some
			// backends only emit reasoning_content per chunk while content
			// stays empty until the very last delta. Surface those chunks
			// so the caller sees progress instead of a long silent gap.
			if text == "" && delta.ReasoningContent != "" {
				text = delta.ReasoningContent
			}
		}
		if text == "" {
			continue
		}

		geminiChunk := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Role:  "model",
						Parts: []GeminiPart{{Text: text}},
					},
					FinishReason: strings.ToUpper(chunk.Choices[0].FinishReason),
				},
			},
		}

		chunkData, _ := json.Marshal(geminiChunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkData)
		if f != nil {
			f.Flush()
		}
	}
	if totalTokens == 0 {
		totalTokens = 1 // minimum 1 per request when API does not report usage
	}
	s.keyMgr.RecordSuccess(key, totalTokens)
}

// ─── Plugin OAI-compat → Gemini SSE Bridge ───────────────────────────────────

// streamPluginAsGemini streams from a plugin-defined OAI-compat backend
// (lmstudio / vllm / llamacpp / jan / etc.) and reformats each chunk into
// Gemini's SSE shape. Mirrors streamOpenRouter's transcoding logic but
// substitutes baseURL/auth/TLS from the plugin yaml so any drop-in plugin
// participates without a Go edit.
//
// Returns silently on success and emits a single error chunk on any
// failure path (matches streamOpenRouter's error contract).
func (s *Server) streamPluginAsGemini(ctx context.Context, w http.ResponseWriter, f http.Flusher, serviceID, model string, req *GeminiRequest) {
	plugin := s.pluginByID[serviceID]
	if plugin == nil {
		writeGeminiErrorChunk(w, f, fmt.Errorf("plugin %s not registered", serviceID))
		return
	}

	// URL resolution shared with callLocalService.
	baseURL := s.resolveLocalServiceURL(serviceID)
	if baseURL == "" {
		writeGeminiErrorChunk(w, f, fmt.Errorf("%s: URL 미설정", serviceID))
		return
	}

	// Per-service semaphore — same shape as streamOllama / callLocalService.
	if sem, ok := s.localSems[serviceID]; ok {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			writeGeminiErrorChunk(w, f, ctx.Err())
			return
		}
	}

	// Strip "publisher/" prefix unless the hub-style plugin says preserve.
	sentModel := model
	if !plugin.PreserveModelID {
		if i := strings.Index(sentModel, "/"); i >= 0 {
			sentModel = sentModel[i+1:]
		}
	}

	oaiReq := GeminiToOpenAI(sentModel, req)
	oaiReq.Stream = true
	s.mu.RLock()
	reasoning := s.serviceReasoning[serviceID]
	s.mu.RUnlock()
	oaiReq.Reasoning = reasoning
	oaiReq.Think = &reasoning
	s.applyQwen3NoThinkSuffix(oaiReq, serviceID, sentModel, reasoning)

	data, err := json.Marshal(oaiReq)
	if err != nil {
		writeGeminiErrorChunk(w, f, fmt.Errorf("%s: marshal: %w", serviceID, err))
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		writeGeminiErrorChunk(w, f, fmt.Errorf("%s: stream request build: %w", serviceID, err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if plugin.Auth.Type == "bearer" && s.cfg.Proxy.VaultToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	}

	var client *http.Client
	if plugin.TLSInternalCA {
		client = internalHTTPClient(10 * time.Minute)
	} else {
		client = &http.Client{Timeout: 10 * time.Minute}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		writeGeminiErrorChunk(w, f, fmt.Errorf("%s: stream connect: %w", serviceID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeGeminiErrorChunk(w, f, fmt.Errorf("%s: stream status %d", serviceID, resp.StatusCode))
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}
		var chunk OpenAIResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		text := ""
		if delta != nil {
			text = delta.Content
			if text == "" && delta.ReasoningContent != "" {
				text = delta.ReasoningContent
			}
		}
		if text == "" {
			continue
		}
		geminiChunk := GeminiResponse{
			Candidates: []GeminiCandidate{{
				Content: GeminiContent{
					Role:  "model",
					Parts: []GeminiPart{{Text: text}},
				},
				FinishReason: strings.ToUpper(chunk.Choices[0].FinishReason),
			}},
		}
		chunkData, _ := json.Marshal(geminiChunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkData)
		if f != nil {
			f.Flush()
		}
	}
}

// ─── Ollama Streaming ─────────────────────────────────────────────────────────

func (s *Server) streamOllama(ctx context.Context, w http.ResponseWriter, f http.Flusher, model string, req *GeminiRequest) {
	if model == "" {
		// Prefer the proxy's configured default model when callers leave it blank.
		// The previous hardcoded "qwen3.5:35b" was wrong on hosts whose Ollama did
		// not have that exact tag pulled. If no default is configured either,
		// fail loudly rather than silently invoking a model the server cannot serve.
		s.mu.RLock()
		model = s.model
		s.mu.RUnlock()
		if model == "" {
			writeGeminiErrorChunk(w, f, fmt.Errorf("ollama: no model specified and no default configured"))
			return
		}
	}
	ollamaURL := s.ollamaURL()

	// Fleet time distribution — same AgentOffset + FallbackJitter as the
	// non-streaming path. Restored in v0.2.27. See timing.go.
	if d := AgentOffset(s.cfg.Proxy.ClientID, localAgentOffsetMs) +
		FallbackJitter(localFallbackJitterMs); d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			writeGeminiErrorChunk(w, f, ctx.Err())
			return
		}
	}

	// Bounded by the per-service local semaphore. Abort the acquire if
	// the caller's context was cancelled (e.g. client disconnect while
	// we're queued) so we don't leak a goroutine waiting on a slot whose
	// response will be dropped anyway.
	sem := s.localSems["ollama"]
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		writeGeminiErrorChunk(w, f, ctx.Err())
		return
	}

	ollamaReq := GeminiToOllama(model, req)
	ollamaReq.Stream = true
	// Pin per-request think to avoid silent reasoning blow-ups on
	// thinking-capable models. See OpenAIRequest.Think rationale.
	s.mu.RLock()
	reasoning := s.serviceReasoning["ollama"]
	s.mu.RUnlock()
	ollamaReq.Think = &reasoning
	data, _ := json.Marshal(ollamaReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", ollamaURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		writeGeminiErrorChunk(w, f, fmt.Errorf("Ollama 요청 생성 실패: %w", err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Match callOllama's budget — local inference with cold model reload
	// (OLLAMA_KEEP_ALIVE can unload large models between calls) easily
	// exceeds cfg.Proxy.Timeout's default 60s.
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		writeGeminiErrorChunk(w, f, fmt.Errorf("Ollama 연결 실패: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeGeminiErrorChunk(w, f, fmt.Errorf("Ollama 오류: HTTP %d", resp.StatusCode))
		return
	}

	// Ollama streaming: parse JSON line-by-line → convert to Gemini SSE
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var chunk OllamaResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		geminiChunk := &GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Role:  "model",
						Parts: []GeminiPart{{Text: chunk.Message.Content}},
					},
				},
			},
		}
		if chunk.Done {
			geminiChunk.Candidates[0].FinishReason = "STOP"
		}

		chunkData, _ := json.Marshal(geminiChunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkData)
		if f != nil {
			f.Flush()
		}

		if chunk.Done {
			break
		}
	}
}

// ─── Util ─────────────────────────────────────────────────────────────────────

func writeGeminiChunk(w io.Writer, f http.Flusher, resp *GeminiResponse, final bool) {
	if final && len(resp.Candidates) > 0 {
		resp.Candidates[0].FinishReason = "STOP"
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f != nil {
		f.Flush()
	}
}

func writeGeminiErrorChunk(w io.Writer, f http.Flusher, err error) {
	errResp := GeminiResponse{
		Error: &GeminiError{Code: 500, Message: err.Error(), Status: "INTERNAL"},
	}
	data, _ := json.Marshal(errResp)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f != nil {
		f.Flush()
	}
}

// writeOpenAIErrorChunk emits a single OpenAI chat.completion.chunk
// SSE event carrying the error message in choices[0].delta.content,
// then a DONE terminator. Mirrors writeGeminiErrorChunk for the
// OpenAI-compatible response shape. Used when an OpenAI-compatible
// streaming dispatch aborts after SSE headers have been committed,
// so the caller still sees a valid SSE termination instead of a
// half-open connection.
func writeOpenAIErrorChunk(w io.Writer, f http.Flusher, err error) {
	chunk := map[string]interface{}{
		"id":     "chatcmpl-proxy",
		"object": "chat.completion.chunk",
		"model":  "",
		"choices": []map[string]interface{}{{
			"index":         0,
			"delta":         map[string]interface{}{"role": "assistant", "content": "[wall-vault: " + err.Error() + "]"},
			"finish_reason": "stop",
		}},
	}
	if b, mErr := json.Marshal(chunk); mErr == nil {
		fmt.Fprintf(w, "data: %s\n\n", b)
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
	if f != nil {
		f.Flush()
	}
}

// streamLocalService dispatches an OpenAI-compatible chat completion
// request to a local backend in oaiCompatServices with stream:true,
// and pipes each SSE line straight back to the caller's ResponseWriter
// without buffering. Returns nil on a clean stream termination (DONE
// seen) or the error that aborted it.
//
// Mirrors callLocalService for everything except the buffering
// strategy: same baseURL/plugin resolution, same auth header decision
// (plugin.Auth.Type == "bearer"), same TLS trust decision
// (plugin.TLSInternalCA → internalHTTPClient), same per-service
// semaphore (s.localSems[serviceID]), same request-body mutations
// (Reasoning/Think and the qwen3 inline /no_think tag).
func (s *Server) streamLocalService(
	ctx context.Context,
	w http.ResponseWriter,
	flusher http.Flusher,
	serviceID string,
	model string,
	oaiReq *OpenAIRequest,
) error {
	plugin := s.pluginByID[serviceID]
	// URL resolution shared with callLocalService.
	baseURL := s.resolveLocalServiceURL(serviceID)
	if baseURL == "" {
		return fmt.Errorf("%s: URL 미설정", serviceID)
	}

	// Fleet time distribution — same AgentOffset + FallbackJitter as
	// callLocalService and streamOllama. Restored in v0.2.27. See timing.go.
	if d := AgentOffset(s.cfg.Proxy.ClientID, localAgentOffsetMs) +
		FallbackJitter(localFallbackJitterMs); d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Per-service semaphore — same fleet-time-distribution shape as
	// callLocalService and streamOllama. Streaming callers queue
	// behind in-flight non-stream callers; that's intentional.
	if sem, ok := s.localSems[serviceID]; ok {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Strip publisher prefix when the plugin is path-style hub mode
	// without preserve_model_id — same rule callLocalService uses.
	sentModel := model
	if plugin == nil || !plugin.PreserveModelID {
		if i := strings.Index(sentModel, "/"); i >= 0 {
			sentModel = sentModel[i+1:]
		}
	}

	// Apply the same per-request mutations as callLocalService so the
	// only field that diverges between paths is Stream.
	s.mu.RLock()
	reasoning := s.serviceReasoning[serviceID]
	s.mu.RUnlock()
	oaiReq.Model = sentModel
	oaiReq.Stream = true
	oaiReq.Reasoning = reasoning
	oaiReq.Think = &reasoning
	s.applyQwen3NoThinkSuffix(oaiReq, serviceID, sentModel, reasoning)

	data, _ := json.Marshal(oaiReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s: stream request build: %w", serviceID, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if plugin != nil && plugin.Auth.Type == "bearer" && s.cfg.Proxy.VaultToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	}

	var client *http.Client
	if plugin != nil && plugin.TLSInternalCA {
		client = internalHTTPClient(10 * time.Minute)
	} else {
		client = &http.Client{Timeout: 10 * time.Minute}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%s: stream connect: %w", serviceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%s: stream status %d: %s", serviceID, resp.StatusCode, string(body))
	}

	// Caller-side SSE headers (idempotent — handler may have set them).
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
	}

	scanner := bufio.NewScanner(resp.Body)
	// Larger token buffer — tool-call chunks and reasoning emits can
	// exceed Scanner's default 64 KB.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	sawDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			fmt.Fprint(w, "\n")
			if flusher != nil {
				flusher.Flush()
			}
			continue
		}
		if strings.HasPrefix(line, "data: [DONE]") {
			fmt.Fprint(w, "data: [DONE]\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			sawDone = true
			break
		}
		rewritten := rewriteOpenAIChunkModel(line, model)
		fmt.Fprint(w, rewritten+"\n")
		if flusher != nil {
			flusher.Flush()
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		if ctx.Err() != nil {
			return nil // caller disconnect; not an error from our POV
		}
		return fmt.Errorf("%s: stream read: %w", serviceID, scanErr)
	}
	if !sawDone {
		// Backend closed without DONE — emit synthetic finish chunk +
		// DONE so caller's parser terminates cleanly.
		finishChunk := map[string]interface{}{
			"id":     "chatcmpl-proxy",
			"object": "chat.completion.chunk",
			"model":  model,
			"choices": []map[string]interface{}{{
				"index":         0,
				"delta":         map[string]interface{}{},
				"finish_reason": "stop",
			}},
		}
		if b, mErr := json.Marshal(finishChunk); mErr == nil {
			fmt.Fprintf(w, "data: %s\n\n", b)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}
	return nil
}

// rewriteOpenAIChunkModel parses an SSE data line, replaces the chunk's
// model field with mdl, and returns the re-marshalled line. Lines
// that are not parseable JSON, are the [DONE] terminator, or do not
// have an "object" key are returned verbatim — defensive for backends
// that emit non-standard event types we don't want to silently drop.
func rewriteOpenAIChunkModel(line, mdl string) string {
	const prefix = "data: "
	if !strings.HasPrefix(line, prefix) {
		return line
	}
	payload := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if payload == "[DONE]" || payload == "" {
		return line
	}
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		return line
	}
	if _, hasObject := chunk["object"]; !hasObject {
		return line
	}
	chunk["model"] = mdl
	// Note: Marshal of a map alphabetises keys; OpenAI SSE clients
	// parse JSON shape, not byte order, so this is benign.
	b, err := json.Marshal(chunk)
	if err != nil {
		return line
	}
	return prefix + string(b)
}
