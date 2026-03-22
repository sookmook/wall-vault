package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// callAnthropicPassthrough: forward the original AnthropicRequest directly to
// Anthropic without any format conversion. Preserves tools, multi-block content,
// and all other fields that would be lost in a GeminiRequest round-trip.
// Returns the raw Anthropic response body, or an error.
func (s *Server) callAnthropicPassthrough(req *AnthropicRequest, model string) ([]byte, *localKey, error) {
	if !strings.HasPrefix(model, "claude-") {
		model = "claude-haiku-4-5-20251001"
	}

	// Build a sanitised copy: strip fields the upstream doesn't accept.
	type passthroughReq struct {
		Model       string             `json:"model"`
		Messages    []AnthropicMessage `json:"messages"`
		System      json.RawMessage    `json:"system,omitempty"`
		MaxTokens   int                `json:"max_tokens,omitempty"`
		Temperature *float64           `json:"temperature,omitempty"`
		Stream      bool               `json:"stream,omitempty"`
		Tools       []interface{}      `json:"tools,omitempty"`
		// service_tier and thinking intentionally omitted
	}
	mt := req.MaxTokens
	if mt == 0 {
		mt = 8192
	}
	payload := passthroughReq{
		Model:       model,
		Messages:    req.Messages,
		System:      req.System,
		MaxTokens:   mt,
		Temperature: req.Temperature,
		Stream:      req.Stream,
		Tools:       req.Tools,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("Anthropic passthrough 직렬화 오류: %w", err)
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		key, plainKey, err := s.getKey("anthropic")
		if err != nil {
			lastErr = err
			break
		}
		resp, err := s.doRequest("POST", "https://api.anthropic.com/v1/messages", data, map[string]string{
			"x-api-key":         plainKey,
			"anthropic-version": "2023-06-01",
		})
		if err != nil {
			s.keyMgr.RecordError(key, 0)
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPaymentRequired {
			s.keyMgr.RecordError(key, resp.StatusCode)
			lastErr = fmt.Errorf("Anthropic passthrough: HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			s.keyMgr.RecordError(key, resp.StatusCode)
			return nil, key, fmt.Errorf("Anthropic passthrough: HTTP %d: %s", resp.StatusCode, body)
		}

		// record token usage
		var usage struct {
			Usage *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal(body, &usage) == nil && usage.Usage != nil {
			tokens := usage.Usage.InputTokens + usage.Usage.OutputTokens
			s.keyMgr.RecordSuccess(key, tokens)
			log.Printf("[proxy] anthropic passthrough ok: model=%s tokens=%d", model, tokens)
		} else {
			s.keyMgr.RecordSuccess(key, 0)
		}
		return body, key, nil
	}
	return nil, nil, lastErr
}

// callAnthropic: call Anthropic API directly using vault anthropic keys.
// If `model` is not a Claude model, falls back to claude-haiku-4-5-20251001.
func (s *Server) callAnthropic(model string, req *GeminiRequest) (*GeminiResponse, error) {
	if !strings.HasPrefix(model, "claude-") {
		model = "claude-haiku-4-5-20251001"
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		key, plainKey, err := s.getKey("anthropic")
		if err != nil {
			lastErr = err
			break
		}

		resp, err := s.doAnthropicRequest(plainKey, model, req)
		if err != nil {
			s.keyMgr.RecordError(key, 0)
			lastErr = err
			break
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("Anthropic: 모델 없음 (%s)", model)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPaymentRequired {
			resp.Body.Close()
			s.keyMgr.RecordError(key, resp.StatusCode)
			lastErr = fmt.Errorf("Anthropic 오류: HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode == http.StatusBadRequest {
			// 400: request error (wrong model name, unsupported params) —
			// not a key fault; skip without cooldown so dispatch falls through.
			resp.Body.Close()
			return nil, fmt.Errorf("Anthropic 오류: HTTP %d", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			s.keyMgr.RecordError(key, resp.StatusCode)
			return nil, fmt.Errorf("Anthropic 오류: HTTP %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var antResp struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			Usage *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &antResp); err != nil {
			return nil, fmt.Errorf("Anthropic 응답 파싱 오류: %w", err)
		}
		if antResp.Error != nil {
			return nil, fmt.Errorf("Anthropic: %s", antResp.Error.Message)
		}

		var sb strings.Builder
		for _, c := range antResp.Content {
			if c.Type == "text" {
				sb.WriteString(c.Text)
			}
		}
		tokens := 0
		if antResp.Usage != nil {
			tokens = antResp.Usage.InputTokens + antResp.Usage.OutputTokens
		}
		s.keyMgr.RecordSuccess(key, tokens)
		log.Printf("[proxy] anthropic ok: model=%s tokens=%d", model, tokens)

		return &GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Role:  "model",
						Parts: []GeminiPart{{Text: sb.String()}},
					},
					FinishReason: "STOP",
				},
			},
		}, nil
	}
	return nil, lastErr
}

func (s *Server) doAnthropicRequest(apiKey, model string, req *GeminiRequest) (*http.Response, error) {
	type antMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type antReq struct {
		Model     string   `json:"model"`
		MaxTokens int      `json:"max_tokens"`
		System    string   `json:"system,omitempty"`
		Messages  []antMsg `json:"messages"`
	}

	ar := antReq{
		Model:     model,
		MaxTokens: 8192,
	}

	if req.SystemInstruction != nil {
		ar.System = extractText(req.SystemInstruction.Parts)
	}

	for _, turn := range req.Contents {
		role := turn.Role
		if role == "model" {
			role = "assistant"
		}
		if role != "user" && role != "assistant" {
			continue
		}
		ar.Messages = append(ar.Messages, antMsg{
			Role:    role,
			Content: extractText(turn.Parts),
		})
	}

	if len(ar.Messages) == 0 {
		return nil, fmt.Errorf("Anthropic: 변환할 메시지 없음")
	}

	data, _ := json.Marshal(ar)
	return s.doRequest("POST", "https://api.anthropic.com/v1/messages", data, map[string]string{
		"x-api-key":         apiKey,
		"anthropic-version": "2023-06-01",
	})
}
