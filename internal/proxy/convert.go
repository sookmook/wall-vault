package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
					if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
						// Upstream (or our own converter upstream of this) sent a
						// tool_call whose arguments field isn't valid JSON. Log with
						// the function name so a misbehaving agent can be spotted,
						// and fall back to an empty object so the conversion doesn't
						// drop the call entirely (Gemini requires an args map).
						log.Printf("[convert] tool_call %q arguments parse failed: %v", name, err)
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

		// Multi-part content (text + audio + image + video) — convert each part
		// into the appropriate GeminiPart (Text or InlineData) so Gemini can
		// process multimodal inputs end-to-end.
		if len(msg.RawContent) > 0 {
			parts := openaiPartsToGemini(msg.RawContent)
			if len(parts) > 0 {
				gemini.Contents = append(gemini.Contents, GeminiContent{Role: role, Parts: parts})
				continue
			}
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

// openaiPartsToGemini maps OpenAI's multi-part content array to Gemini parts.
// Recognised types:
//   - "text"         → GeminiPart{Text}
//   - "input_audio"  → GeminiPart{InlineData} (audio/<format>)
//   - "input_video"  → GeminiPart{InlineData} (video/<format>)
//   - "input_image"  → GeminiPart{InlineData} (image/<format>)
//   - "input_file"   → GeminiPart{InlineData} (mime taken verbatim)
//   - "image_url"    → GeminiPart{InlineData} when url is a data: URI
//                      (mime is whatever the URI declares — image/png, video/mp4, …)
//
// External http(s) URLs in image_url are intentionally skipped this round —
// fetch + size limits + caching require their own design.
func openaiPartsToGemini(raw json.RawMessage) []GeminiPart {
	var arr []map[string]interface{}
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil
	}
	out := make([]GeminiPart, 0, len(arr))
	for _, p := range arr {
		t, _ := p["type"].(string)
		switch t {
		case "text":
			if s, ok := p["text"].(string); ok && s != "" {
				out = append(out, GeminiPart{Text: s})
			}
		case "input_audio":
			if a, ok := p["input_audio"].(map[string]interface{}); ok {
				data, _ := a["data"].(string)
				format, _ := a["format"].(string)
				if data != "" {
					out = append(out, GeminiPart{
						InlineData: &BlobData{MimeType: audioFormatToMime(format), Data: data},
					})
				}
			}
		case "input_video":
			if v, ok := p["input_video"].(map[string]interface{}); ok {
				data, _ := v["data"].(string)
				format, _ := v["format"].(string)
				if data != "" {
					out = append(out, GeminiPart{
						InlineData: &BlobData{MimeType: videoFormatToMime(format), Data: data},
					})
				}
			}
		case "input_image":
			if im, ok := p["input_image"].(map[string]interface{}); ok {
				data, _ := im["data"].(string)
				format, _ := im["format"].(string)
				if data != "" {
					out = append(out, GeminiPart{
						InlineData: &BlobData{MimeType: imageFormatToMime(format), Data: data},
					})
				}
			}
		case "input_file":
			if f, ok := p["input_file"].(map[string]interface{}); ok {
				data, _ := f["data"].(string)
				mime, _ := f["mime"].(string)
				if mime == "" {
					mime, _ = f["mime_type"].(string)
				}
				if data != "" && mime != "" {
					out = append(out, GeminiPart{
						InlineData: &BlobData{MimeType: mime, Data: data},
					})
				}
			}
		case "image_url":
			if iu, ok := p["image_url"].(map[string]interface{}); ok {
				url, _ := iu["url"].(string)
				if mime, data, ok := parseDataURI(url); ok && data != "" {
					out = append(out, GeminiPart{
						InlineData: &BlobData{MimeType: mime, Data: data},
					})
				} else if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
					if mime, data, ok := fetchAsBase64(url, 5<<20); ok {
						out = append(out, GeminiPart{
							InlineData: &BlobData{MimeType: mime, Data: data},
						})
					}
				}
			}
		}
	}
	return out
}

// fetchAsBase64 downloads an http(s) URL up to maxBytes and returns
// (mimeType, base64Body, true). Bodies larger than maxBytes, non-2xx
// responses, and network errors all return ok=false. Used by
// openaiPartsToGemini for image_url entries that point to external URLs.
func fetchAsBase64(url string, maxBytes int64) (mime, data string, ok bool) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[multimodal] fetch failed: %s: %v", url, err)
		return "", "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[multimodal] fetch %s: HTTP %d", url, resp.StatusCode)
		return "", "", false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		log.Printf("[multimodal] read %s: %v", url, err)
		return "", "", false
	}
	if int64(len(body)) > maxBytes {
		log.Printf("[multimodal] %s exceeds %d bytes", url, maxBytes)
		return "", "", false
	}
	mime = resp.Header.Get("Content-Type")
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	return mime, base64.StdEncoding.EncodeToString(body), true
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

// AnthropicToGemini converts an Anthropic /v1/messages request into a GeminiRequest
// while preserving tool_use / tool_result content blocks. Naive text-only extraction
// (the previous implementation) collapsed tool blocks to empty-text parts, which
// upstream Google Gemini / OpenRouter / Anthropic / Ollama then rejected as
// "contents is not specified" or "messages is too short".
//
// We route through the OpenAI-intermediate representation because
// anthropicToOpenAIReq already handles content-block arrays, tool_use → tool_calls,
// and tool_result → role=tool mapping correctly; OpenAIToGemini then rebuilds a
// Gemini-native structure with functionCall / functionResponse parts.
func AnthropicToGemini(req *AnthropicRequest) *GeminiRequest {
	// Build rich OAI representation (preserves tool_use / tool_result blocks).
	oai := anthropicToOpenAIReq(req, req.Model)
	// Convert to Gemini form — OpenAIToGemini drops empty-content messages that
	// would cause "contents is not specified" and materializes functionCall /
	// functionResponse parts for tool turns.
	gemini := OpenAIToGemini(oai)

	// Carry generation params the OAI intermediate may have dropped.
	if req.Temperature != nil || req.MaxTokens > 0 {
		if gemini.GenerationConfig == nil {
			gemini.GenerationConfig = &GenerationConfig{}
		}
		if req.Temperature != nil {
			gemini.GenerationConfig.Temperature = req.Temperature
		}
		if req.MaxTokens > 0 {
			gemini.GenerationConfig.MaxOutputTokens = &req.MaxTokens
		}
	}

	return gemini
}

// ─── Anthropic → OpenAI ──────────────────────────────────────────────────────
// anthropicToOpenAIReq converts an Anthropic /v1/messages request to an OpenAI
// chat-completions request, preserving tool_use / tool_result content blocks.
// This is used by handleAnthropic's dispatch path (via RawOAI) so that
// OpenRouter-based backends receive tools and function calls correctly.

func anthropicToOpenAIReq(req *AnthropicRequest, model string) *OpenAIRequest {
	oai := &OpenAIRequest{
		Model:       model,
		Stream:      req.Stream,
		Temperature: req.Temperature,
	}
	if req.MaxTokens > 0 {
		mt := req.MaxTokens
		oai.MaxTokens = &mt
	}

	// System instruction
	if sys := req.SystemText(); sys != "" {
		oai.Messages = append(oai.Messages, OpenAIMessage{Role: "system", Content: sys})
	}

	// Convert Anthropic messages to OpenAI messages
	for _, msg := range req.Messages {
		// Try to parse as content blocks array
		var blocks []struct {
			Type      string          `json:"type"`
			Text      string          `json:"text"`
			ID        string          `json:"id"`
			Name      string          `json:"name"`
			Input     json.RawMessage `json:"input"`
			ToolUseID string          `json:"tool_use_id"`
			Content   json.RawMessage `json:"content"`
		}
		if json.Unmarshal(msg.Content, &blocks) == nil && len(blocks) > 0 {
			if msg.Role == "assistant" {
				// Extract text and tool_use blocks
				text := ""
				var toolCalls []map[string]interface{}
				for _, b := range blocks {
					switch b.Type {
					case "text":
						text += b.Text
					case "tool_use":
						argsJSON, _ := json.Marshal(b.Input)
						toolCalls = append(toolCalls, map[string]interface{}{
							"id":   b.ID,
							"type": "function",
							"function": map[string]interface{}{
								"name":      b.Name,
								"arguments": string(argsJSON),
							},
						})
					}
				}
				oaiMsg := OpenAIMessage{Role: "assistant", Content: text}
				if len(toolCalls) > 0 {
					tc, _ := json.Marshal(toolCalls)
					oaiMsg.ToolCalls = tc
				}
				oai.Messages = append(oai.Messages, oaiMsg)
			} else {
				// User messages: handle tool_result blocks
				for _, b := range blocks {
					switch b.Type {
					case "text":
						oai.Messages = append(oai.Messages, OpenAIMessage{Role: "user", Content: b.Text})
					case "tool_result":
						content := ""
						// tool_result content can be string or content blocks
						if err := json.Unmarshal(b.Content, &content); err != nil {
							// Try as content blocks
							var parts []struct {
								Type string `json:"type"`
								Text string `json:"text"`
							}
							if err2 := json.Unmarshal(b.Content, &parts); err2 == nil {
								for _, p := range parts {
									if p.Type == "text" {
										content += p.Text
									}
								}
							} else {
								// Neither string nor content-block array — this is
								// unexpected from a spec-conformant Anthropic client.
								// Log once so the malformed tool result can be traced
								// back to the caller; still forward the raw bytes as
								// a string so the turn isn't silently dropped.
								log.Printf("[convert] tool_result content parse failed (string: %v, blocks: %v), forwarding raw", err, err2)
								content = string(b.Content)
							}
						}
						oai.Messages = append(oai.Messages, OpenAIMessage{
							Role:       "tool",
							ToolCallID: b.ToolUseID,
							Content:    content,
						})
					}
				}
			}
			continue
		}

		// Simple string content
		oai.Messages = append(oai.Messages, OpenAIMessage{
			Role:    msg.Role,
			Content: msg.ContentText(),
		})
	}

	// Convert Anthropic tools to OpenAI tool format
	for _, t := range req.Tools {
		toolMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		oaiTool := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        toolMap["name"],
				"description": toolMap["description"],
				"parameters":  toolMap["input_schema"],
			},
		}
		oai.Tools = append(oai.Tools, oaiTool)
	}

	return oai
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
