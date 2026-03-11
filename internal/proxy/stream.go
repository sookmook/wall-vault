package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// ─── Gemini 스트리밍 핸들러 ──────────────────────────────────────────────────

// handleGeminiStream: streamGenerateContent 엔드포인트 처리
func (s *Server) handleGeminiStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

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
		log.Printf("[Security] 스트림 요청에서 %d개 도구 차단", stripped)
	}

	s.mu.RLock()
	svc := s.service
	mdl := s.model
	s.mu.RUnlock()

	if urlModel := extractModelFromPath(r.URL.Path); urlModel != "" {
		if strings.HasPrefix(urlModel, "gemini-") {
			svc = "google"
			mdl = urlModel
		}
	}

	// SSE 헤더 설정
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
		s.streamOllama(w, flusher, mdl, &req)
	default:
		writeGeminiErrorChunk(w, flusher, fmt.Errorf("서비스 미지원: %s", svc))
	}
}

// ─── Google 스트리밍 패스스루 ────────────────────────────────────────────────

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

	// Google SSE 패스스루: 그대로 클라이언트에 전달
	tokens := 0
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		if f != nil {
			f.Flush()
		}
		// 토큰 수 추정 (usageMetadata 파싱)
		if strings.HasPrefix(line, "data: ") {
			var chunk GeminiResponse
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chunk); err == nil {
				if chunk.UsageMetadata != nil {
					tokens = chunk.UsageMetadata.TotalTokenCount
				}
			}
		}
	}
	s.keyMgr.RecordSuccess(key, tokens)
}

// ─── OpenRouter 스트리밍 변환 ────────────────────────────────────────────────

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

	// OpenAI SSE → Gemini SSE 변환
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
		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		if delta == nil || delta.Content == "" {
			continue
		}

		geminiChunk := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Role:  "model",
						Parts: []GeminiPart{{Text: delta.Content}},
					},
					FinishReason: strings.ToUpper(chunk.Choices[0].FinishReason),
				},
			},
		}
		if chunk.Usage != nil {
			totalTokens = chunk.Usage.TotalTokens
		}

		chunkData, _ := json.Marshal(geminiChunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkData)
		if f != nil {
			f.Flush()
		}
	}
	s.keyMgr.RecordSuccess(key, totalTokens)
}

// ─── Ollama 스트리밍 ──────────────────────────────────────────────────────────

func (s *Server) streamOllama(w http.ResponseWriter, f http.Flusher, model string, req *GeminiRequest) {
	if model == "" {
		model = "qwen3.5:35b"
	}
	ollamaURL := s.ollamaURL()

	// Ollama 동시 요청 제한
	s.ollamaMu.Lock()
	defer s.ollamaMu.Unlock()

	ollamaReq := GeminiToOllama(model, req)
	ollamaReq.Stream = true
	data, _ := json.Marshal(ollamaReq)

	httpReq, _ := http.NewRequest("POST", ollamaURL+"/api/chat", bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: s.cfg.Proxy.Timeout}
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

	// Ollama 스트리밍: 한 줄씩 JSON 파싱 → Gemini SSE 변환
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

// ─── 유틸 ────────────────────────────────────────────────────────────────────

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
