package proxy

import (
	"fmt"
	"strings"
	"time"
)

// ─── Gemini → OpenAI ─────────────────────────────────────────────────────────

func GeminiToOpenAI(model string, req *GeminiRequest) *OpenAIRequest {
	oai := &OpenAIRequest{
		Model: model,
	}

	// system prompt
	if req.SystemInstruction != nil {
		text := extractText(req.SystemInstruction.Parts)
		if text != "" {
			oai.Messages = append(oai.Messages, OpenAIMessage{
				Role:    "system",
				Content: text,
			})
		}
	}

	// conversation contents
	for _, c := range req.Contents {
		role := c.Role
		if role == "model" {
			role = "assistant"
		}
		text := extractText(c.Parts)
		oai.Messages = append(oai.Messages, OpenAIMessage{
			Role:    role,
			Content: text,
		})
	}

	// generation parameters
	if req.GenerationConfig != nil {
		oai.Temperature = req.GenerationConfig.Temperature
		oai.MaxTokens = req.GenerationConfig.MaxOutputTokens
	}

	return oai
}

// ─── OpenAI → Gemini ─────────────────────────────────────────────────────────

func OpenAIToGemini(req *OpenAIRequest) *GeminiRequest {
	gemini := &GeminiRequest{}

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			gemini.SystemInstruction = &GeminiContent{
				Parts: []GeminiPart{{Text: msg.Content}},
			}
			continue
		}
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		gemini.Contents = append(gemini.Contents, GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: msg.Content}},
		})
	}

	if req.Temperature != nil || req.MaxTokens != nil {
		gemini.GenerationConfig = &GenerationConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		}
	}

	return gemini
}

// ─── OpenAI Response → Gemini Response ───────────────────────────────────────

func OpenAIRespToGemini(resp *OpenAIResponse) *GeminiResponse {
	gr := &GeminiResponse{}
	for i, c := range resp.Choices {
		reason := strings.ToUpper(c.FinishReason)
		if reason == "" {
			reason = "STOP"
		}
		gr.Candidates = append(gr.Candidates, GeminiCandidate{
			Content: GeminiContent{
				Role:  "model",
				Parts: []GeminiPart{{Text: c.Message.Content}},
			},
			FinishReason: reason,
			Index:        i,
		})
	}
	if resp.Usage != nil {
		gr.UsageMetadata = &GeminiUsage{
			PromptTokenCount:     resp.Usage.PromptTokens,
			CandidatesTokenCount: resp.Usage.CompletionTokens,
			TotalTokenCount:      resp.Usage.TotalTokens,
		}
	}
	return gr
}

// ─── Ollama Response → Gemini Response ───────────────────────────────────────

func OllamaRespToGemini(resp *OllamaResponse) *GeminiResponse {
	return &GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Content: GeminiContent{
					Role:  "model",
					Parts: []GeminiPart{{Text: resp.Message.Content}},
				},
				FinishReason: "STOP",
			},
		},
	}
}

// ─── Gemini → Ollama ─────────────────────────────────────────────────────────

func GeminiToOllama(model string, req *GeminiRequest) *OllamaRequest {
	oai := GeminiToOpenAI(model, req)
	ollama := &OllamaRequest{
		Model:    model,
		Messages: oai.Messages,
		Stream:   false,
	}
	if req.GenerationConfig != nil {
		opts := &OllamaOptions{}
		if req.GenerationConfig.Temperature != nil {
			opts.Temperature = *req.GenerationConfig.Temperature
		}
		if req.GenerationConfig.MaxOutputTokens != nil {
			opts.NumPredict = *req.GenerationConfig.MaxOutputTokens
		}
		ollama.Options = opts
	}
	return ollama
}

// ─── Anthropic → Gemini ───────────────────────────────────────────────────────

func AnthropicToGemini(req *AnthropicRequest) *GeminiRequest {
	gemini := &GeminiRequest{}

	if sys := req.SystemText(); sys != "" {
		gemini.SystemInstruction = &GeminiContent{
			Parts: []GeminiPart{{Text: sys}},
		}
	}

	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		text := msg.ContentText()
		gemini.Contents = append(gemini.Contents, GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: text}},
		})
	}

	if req.Temperature != nil || req.MaxTokens > 0 {
		cfg := &GenerationConfig{}
		if req.Temperature != nil {
			cfg.Temperature = req.Temperature
		}
		if req.MaxTokens > 0 {
			cfg.MaxOutputTokens = &req.MaxTokens
		}
		gemini.GenerationConfig = cfg
	}

	return gemini
}

// ─── Gemini → Anthropic ───────────────────────────────────────────────────────

func GeminiRespToAnthropic(model string, resp *GeminiResponse) *AnthropicResponse {
	ar := &AnthropicResponse{
		ID:         fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Type:       "message",
		Role:       "assistant",
		Model:      model,
		StopReason: "end_turn",
	}

	for _, c := range resp.Candidates {
		text := extractText(c.Content.Parts)
		ar.Content = append(ar.Content, AnthropicContent{
			Type: "text",
			Text: text,
		})
	}
	if len(ar.Content) == 0 {
		ar.Content = []AnthropicContent{{Type: "text", Text: ""}}
	}

	if resp.UsageMetadata != nil {
		ar.Usage = AnthropicUsage{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
		}
	}
	return ar
}

// ─── Util ─────────────────────────────────────────────────────────────────────

func extractText(parts []GeminiPart) string {
	var sb strings.Builder
	for _, p := range parts {
		sb.WriteString(p.Text)
	}
	return sb.String()
}
