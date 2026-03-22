package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ─── Gemini → OpenAI ─────────────────────────────────────────────────────────

func GeminiToOpenAI(model string, req *GeminiRequest) *OpenAIRequest {
	oai := &OpenAIRequest{
		Model: model,
	}

	// Prefer original OAI messages when available — they faithfully preserve tool_calls,
	// role=tool, and tool_call_id fields that cannot be reconstructed from Gemini contents.
	// (The Gemini round-trip converts FunctionCall/FunctionResponse parts; using RawOAI
	// avoids the lossy re-conversion that strips those parts down to empty text.)
	if req.RawOAI != nil {
		oai.Messages = req.RawOAI.Messages
		oai.Temperature = req.RawOAI.Temperature
		oai.MaxTokens = req.RawOAI.MaxTokens
		if len(req.RawOAI.Tools) > 0 {
			oai.Tools = req.RawOAI.Tools
			oai.ToolChoice = req.RawOAI.ToolChoice
		}
		return oai
	}

	// RawOAI not available (e.g. request originated from handleGemini or handleAnthropic):
	// reconstruct from Gemini contents.

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
	gemini := &GeminiRequest{RawOAI: req}

	// Track tool_call_id → function_name for functionResponse name lookup.
	toolCallNames := make(map[string]string)

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

		// Handle assistant messages with tool_calls (model function calls).
		if role == "model" && len(msg.ToolCalls) > 0 {
			var tcList []map[string]interface{}
			if json.Unmarshal(msg.ToolCalls, &tcList) == nil {
				var parts []GeminiPart
				for _, tc := range tcList {
					fn, _ := tc["function"].(map[string]interface{})
					if fn == nil {
						continue
					}
					name, _ := fn["name"].(string)
					// Record tool_call_id → function_name for later tool result lookup.
					if id, _ := tc["id"].(string); id != "" && name != "" {
						toolCallNames[id] = name
					}
					argsStr, _ := fn["arguments"].(string)
					var args interface{}
					if json.Unmarshal([]byte(argsStr), &args) != nil {
						args = map[string]interface{}{}
					}
					parts = append(parts, GeminiPart{
						FunctionCall: map[string]interface{}{"name": name, "args": args},
					})
				}
				if len(parts) > 0 {
					gemini.Contents = append(gemini.Contents, GeminiContent{Role: role, Parts: parts})
					continue
				}
			}
		}

		// Handle tool result messages (role=tool → Gemini functionResponse).
		if msg.Role == "tool" {
			// OAI tool results carry tool_call_id (not name); look up the function name.
			fname := toolCallNames[msg.ToolCallID]
			if fname == "" {
				fname = msg.Name // fallback if name was populated directly
			}
			gemini.Contents = append(gemini.Contents, GeminiContent{
				Role: "user",
				Parts: []GeminiPart{{
					FunctionResponse: map[string]interface{}{
						"name":     fname,
						"response": map[string]interface{}{"content": msg.Content},
					},
				}},
			})
			continue
		}

		// Skip empty text parts — Gemini rejects parts with no data.
		if msg.Content == "" {
			continue
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

	// Convert OpenAI tools to Gemini functionDeclarations format.
	if len(req.Tools) > 0 {
		var funcDecls []interface{}
		for _, t := range req.Tools {
			toolMap, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			fn, ok := toolMap["function"].(map[string]interface{})
			if !ok {
				continue
			}
			decl := map[string]interface{}{}
			if name, ok := fn["name"].(string); ok {
				decl["name"] = name
			}
			if desc, ok := fn["description"].(string); ok && desc != "" {
				decl["description"] = desc
			}
			if params := fn["parameters"]; params != nil {
				// Strip JSON Schema fields unsupported by Gemini.
				decl["parameters"] = stripGeminiUnsupported(params)
			}
			funcDecls = append(funcDecls, decl)
		}
		if len(funcDecls) > 0 {
			gemini.Tools = []interface{}{
				map[string]interface{}{"functionDeclarations": funcDecls},
			}
		}
	}

	// Convert tool_choice to Gemini toolConfig.functionCallingConfig.
	if req.ToolChoice != nil {
		mode := "AUTO"
		switch v := req.ToolChoice.(type) {
		case string:
			switch v {
			case "required":
				mode = "ANY"
			case "none":
				mode = "NONE"
			}
		case map[string]interface{}:
			// {"type": "function", "function": {"name": "..."}} → forced specific function
			if fn, ok := v["function"].(map[string]interface{}); ok {
				if name, ok := fn["name"].(string); ok && name != "" {
					gemini.ToolConfig = map[string]interface{}{
						"functionCallingConfig": map[string]interface{}{
							"mode":                 "ANY",
							"allowedFunctionNames": []string{name},
						},
					}
					return gemini
				}
			}
		}
		gemini.ToolConfig = map[string]interface{}{
			"functionCallingConfig": map[string]interface{}{"mode": mode},
		}
	}

	return gemini
}

// stripGeminiUnsupported recursively removes JSON Schema fields that Gemini's
// function declaration API rejects (additionalProperties, patternProperties, etc.).
func stripGeminiUnsupported(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, child := range val {
			switch k {
			case "additionalProperties", "patternProperties",
				"$schema", "$ref", "$defs", "definitions",
				"unevaluatedProperties", "strict":
				// drop — Gemini doesn't support these
			default:
				out[k] = stripGeminiUnsupported(child)
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, elem := range val {
			out[i] = stripGeminiUnsupported(elem)
		}
		return out
	default:
		return v
	}
}

// ─── OpenAI Response → Gemini Response ───────────────────────────────────────

func OpenAIRespToGemini(resp *OpenAIResponse) *GeminiResponse {
	gr := &GeminiResponse{}
	for i, c := range resp.Choices {
		reason := strings.ToUpper(c.FinishReason)
		if reason == "" {
			reason = "STOP"
		}
		cand := GeminiCandidate{
			Content: GeminiContent{
				Role:  "model",
				Parts: []GeminiPart{{Text: c.Message.Content}},
			},
			FinishReason: reason,
			Index:        i,
		}
		// Carry tool_calls through so handleOpenAI can return them to the client.
		if len(c.Message.ToolCalls) > 0 {
			cand.RawToolCalls = c.Message.ToolCalls
		}
		gr.Candidates = append(gr.Candidates, cand)
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
	cand := GeminiCandidate{
		Content: GeminiContent{
			Role:  "model",
			Parts: []GeminiPart{{Text: resp.Message.Content}},
		},
		FinishReason: "STOP",
	}
	// Carry tool_calls through so handleOpenAI can return them to the client.
	if len(resp.Message.ToolCalls) > 0 {
		cand.RawToolCalls = resp.Message.ToolCalls
	}
	return &GeminiResponse{Candidates: []GeminiCandidate{cand}}
}

// ─── Gemini → Ollama ─────────────────────────────────────────────────────────

func GeminiToOllama(model string, req *GeminiRequest) *OllamaRequest {
	oai := GeminiToOpenAI(model, req)
	ollama := &OllamaRequest{
		Model:    model,
		Messages: oai.Messages,
		Tools:    oai.Tools,
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
