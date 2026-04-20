package proxy

import (
	"encoding/json"
	"strings"
	"time"
)

// ─── Current Config ───────────────────────────────────────────────────────────

type RuntimeConfig struct {
	Service   string
	Model     string
	ClientID  string
	VaultURL  string
	Token     string
	UpdatedAt time.Time
}

// ─── Gemini Request/Response Structures ──────────────────────────────────────

type GeminiRequest struct {
	Contents         []GeminiContent    `json:"contents"`
	SystemInstruction *GeminiContent    `json:"systemInstruction,omitempty"`
	GenerationConfig  *GenerationConfig  `json:"generationConfig,omitempty"`
	Tools             []interface{}      `json:"tools,omitempty"`
	ToolConfig        interface{}        `json:"toolConfig,omitempty"`
	SafetySettings    []interface{}      `json:"safetySettings,omitempty"`
	// RawOAI carries the original OpenAI request for OAI-native backends (not serialized).
	RawOAI *OpenAIRequest `json:"-"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text             string      `json:"text,omitempty"`
	InlineData       *BlobData   `json:"inlineData,omitempty"`
	FunctionCall     interface{} `json:"functionCall,omitempty"`
	FunctionResponse interface{} `json:"functionResponse,omitempty"`
}

// BlobData carries inline binary content (audio / image) for Gemini multimodal
// requests. Data is base64-encoded; MimeType is the IANA media type
// (e.g. "audio/wav", "image/png").
type BlobData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GenerationConfig struct {
	Temperature      *float64       `json:"temperature,omitempty"`
	MaxOutputTokens  *int           `json:"maxOutputTokens,omitempty"`
	TopP             *float64       `json:"topP,omitempty"`
	TopK             *int           `json:"topK,omitempty"`
	StopSequences    []string       `json:"stopSequences,omitempty"`
	ThinkingConfig   *ThinkingConfig `json:"thinkingConfig,omitempty"`
}

type ThinkingConfig struct {
	ThinkingBudget int  `json:"thinkingBudget,omitempty"`
	IncludeThoughts bool `json:"includeThoughts,omitempty"`
}

type GeminiResponse struct {
	Candidates    []GeminiCandidate `json:"candidates,omitempty"`
	UsageMetadata *GeminiUsage      `json:"usageMetadata,omitempty"`
	Error         *GeminiError      `json:"error,omitempty"`
}

type GeminiCandidate struct {
	Content      GeminiContent   `json:"content"`
	FinishReason string          `json:"finishReason,omitempty"`
	Index        int             `json:"index,omitempty"`
	// RawToolCalls carries tool_calls from OAI-format backend responses (not serialized).
	RawToolCalls json.RawMessage `json:"-"`
}

type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ─── OpenAI Request/Response Structures ──────────────────────────────────────

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Tools       []interface{}   `json:"tools,omitempty"`
	ToolChoice  interface{}     `json:"tool_choice,omitempty"`
	ServiceTier string          `json:"service_tier,omitempty"` // OpenClaw fast mode (strip before forwarding)
	Reasoning   bool            `json:"reasoning,omitempty"`    // local services: request reasoning/thinking output
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	// RawContent preserves the original parts array (text + input_audio +
	// image_url) when the client sent content in OpenAI's multi-part form.
	// Not serialized — only used inbound for OpenAIToGemini multimodal mapping.
	RawContent json.RawMessage `json:"-"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
}

// MarshalJSON emits the multi-part `content` array when RawContent is set
// (preserving image_url / input_audio / input_image / input_file parts on
// outbound requests to OpenAI-compatible upstreams), and otherwise falls back
// to the flat `content` string. It also guarantees that assistant messages
// always include the `content` field — some clients (e.g. Claude Code) call
// .trim() on content and crash with "Cannot read properties of undefined" if
// it's missing, which can happen when thinking models (gemini-3.1-pro)
// exhaust max_tokens on reasoning before producing visible output.
func (m OpenAIMessage) MarshalJSON() ([]byte, error) {
	type alias OpenAIMessage
	// Multi-part form takes priority so multimodal content is preserved
	// end-to-end: UnmarshalJSON parks the original array in RawContent (and
	// extracts a text-only Content for legacy consumers), and this branch
	// re-emits that array verbatim when the proxy forwards to any OpenAI
	// chat-completions endpoint (Ollama, lmstudio, vllm, llamacpp, OpenAI
	// direct, OpenRouter).
	if len(m.RawContent) > 0 {
		out := map[string]interface{}{
			"role":    m.Role,
			"content": m.RawContent,
		}
		if len(m.ToolCalls) > 0 {
			out["tool_calls"] = m.ToolCalls
		}
		if m.ToolCallID != "" {
			out["tool_call_id"] = m.ToolCallID
		}
		if m.Name != "" {
			out["name"] = m.Name
		}
		return json.Marshal(out)
	}
	if m.Role == "assistant" && m.Content == "" {
		// Build JSON manually to include content:"" explicitly
		out := map[string]interface{}{
			"role":    m.Role,
			"content": "",
		}
		if len(m.ToolCalls) > 0 {
			out["tool_calls"] = m.ToolCalls
		}
		if m.ToolCallID != "" {
			out["tool_call_id"] = m.ToolCallID
		}
		if m.Name != "" {
			out["name"] = m.Name
		}
		return json.Marshal(out)
	}
	return json.Marshal(alias(m))
}

// UnmarshalJSON handles content as either a string or an array of content parts.
// OpenAI spec allows both; many clients (including OpenClaw) use the array form.
func (m *OpenAIMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Role       string          `json:"role"`
		Content    json.RawMessage `json:"content"`
		ToolCalls  json.RawMessage `json:"tool_calls"`
		ToolCallID string          `json:"tool_call_id"`
		Name       string          `json:"name"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Role = raw.Role
	m.Content = rawContentToString(raw.Content)
	// Preserve the original raw payload only when it's the multi-part array
	// form — that's the only case OpenAIToGemini needs for inlineData mapping.
	if isJSONArray(raw.Content) {
		m.RawContent = raw.Content
	}
	m.ToolCalls = raw.ToolCalls
	m.ToolCallID = raw.ToolCallID
	m.Name = raw.Name
	return nil
}

// isJSONArray reports whether the first non-whitespace byte of raw is '['.
func isJSONArray(raw json.RawMessage) bool {
	for _, b := range raw {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		return b == '['
	}
	return false
}

// audioFormatToMime maps OpenAI input_audio.format values to IANA mime types.
// Unknown formats fall through as "audio/<format>" so callers can still try.
func audioFormatToMime(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	switch f {
	case "wav":
		return "audio/wav"
	case "mp3":
		return "audio/mpeg"
	case "ogg":
		return "audio/ogg"
	case "flac":
		return "audio/flac"
	case "webm":
		return "audio/webm"
	case "m4a", "mp4":
		return "audio/mp4"
	case "":
		return "audio/wav"
	}
	if strings.HasPrefix(f, "audio/") {
		return f
	}
	return "audio/" + f
}

// videoFormatToMime maps short video format names to IANA mime types. Used by
// the input_video multimodal part. Unknown formats fall through as "video/<f>".
func videoFormatToMime(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	switch f {
	case "mp4", "m4v":
		return "video/mp4"
	case "mov", "qt":
		return "video/quicktime"
	case "webm":
		return "video/webm"
	case "mkv":
		return "video/x-matroska"
	case "avi":
		return "video/x-msvideo"
	case "":
		return "video/mp4"
	}
	if strings.HasPrefix(f, "video/") {
		return f
	}
	return "video/" + f
}

// imageFormatToMime maps short image format names to IANA mime types.
func imageFormatToMime(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	switch f {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "heic":
		return "image/heic"
	case "":
		return "image/png"
	}
	if strings.HasPrefix(f, "image/") {
		return f
	}
	return "image/" + f
}

// parseDataURI splits a "data:<mime>;base64,<data>" URI into (mime, data, true).
// Returns (empty, empty, false) for non-data or non-base64 URIs.
func parseDataURI(uri string) (mime, data string, ok bool) {
	if !strings.HasPrefix(uri, "data:") {
		return "", "", false
	}
	rest := uri[5:]
	semi := strings.Index(rest, ";base64,")
	if semi < 0 {
		return "", "", false
	}
	return rest[:semi], rest[semi+len(";base64,"):], true
}

// rawContentToString converts OpenAI content (string or parts array) to plain text.
// Handles: string, text parts, image_url parts (v2026.3.7+), and tool_result parts.
func rawContentToString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// string form: "hello"
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	// array form: [{"type":"text","text":"..."}, {"type":"image_url","image_url":{...}}, ...]
	var parts []struct {
		Type     string          `json:"type"`
		Text     string          `json:"text"`
		ImageURL json.RawMessage `json:"image_url"` // image_url — skip (text-only proxy)
	}
	if json.Unmarshal(raw, &parts) == nil {
		var sb strings.Builder
		for _, p := range parts {
			switch p.Type {
			case "text":
				sb.WriteString(p.Text)
			case "image_url":
				sb.WriteString("[이미지]") // placeholder — proxy는 텍스트 전달만 지원
			}
		}
		return sb.String()
	}
	return string(raw)
}

type OpenAIResponse struct {
	ID      string         `json:"id,omitempty"`
	Object  string         `json:"object,omitempty"`
	Model   string         `json:"model,omitempty"`
	Choices []OpenAIChoice `json:"choices,omitempty"`
	Usage   *OpenAIUsage   `json:"usage,omitempty"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

type OpenAIChoice struct {
	Message      OpenAIMessage `json:"message"`
	Delta        *OpenAIMessage `json:"delta,omitempty"`
	FinishReason string        `json:"finish_reason,omitempty"`
	Index        int           `json:"index,omitempty"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// ─── Ollama Request/Response Structures ──────────────────────────────────────

type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Tools    []interface{}   `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
	Options  *OllamaOptions  `json:"options,omitempty"`
}

type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

type OllamaResponse struct {
	Model   string        `json:"model,omitempty"`
	Message OpenAIMessage `json:"message"`
	Done    bool          `json:"done"`
}

// ─── Anthropic Request/Response Structures ────────────────────────────────────

// AnthropicRequest: POST /v1/messages request (Claude API format)
type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	System      json.RawMessage    `json:"system,omitempty"` // string or [{type,text}] array (Claude Code ≥2026.3)
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Tools       []interface{}      `json:"tools,omitempty"`
	ServiceTier string             `json:"service_tier,omitempty"` // OpenClaw v2026.3.12 fast mode (strip before forwarding)
	Thinking    interface{}        `json:"thinking,omitempty"`     // Claude extended thinking (ignore)
}

// SystemText extracts the system prompt text from either a JSON string or a
// [{type:"text", text:"..."}] content-block array.
func (r *AnthropicRequest) SystemText() string {
	if len(r.System) == 0 {
		return ""
	}
	// try plain string first
	var s string
	if json.Unmarshal(r.System, &s) == nil {
		return s
	}
	// try content-block array
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(r.System, &parts) == nil {
		var sb strings.Builder
		for _, p := range parts {
			if p.Type == "text" {
				sb.WriteString(p.Text)
			}
		}
		return sb.String()
	}
	return ""
}

type AnthropicMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // string or []AnthropicContent
}

// ContentText extracts plain text from Anthropic content (string or parts array)
func (m AnthropicMessage) ContentText() string {
	var s string
	if json.Unmarshal(m.Content, &s) == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(m.Content, &parts) == nil {
		var sb strings.Builder
		for _, p := range parts {
			if p.Type == "text" {
				sb.WriteString(p.Text)
			}
		}
		return sb.String()
	}
	return string(m.Content)
}

type AnthropicContent struct {
	Type string `json:"type"`
	// No omitempty: Text must always be emitted when Type is "text",
	// otherwise clients like Claude Code crash with "Cannot read properties of
	// undefined (reading 'trim')" when thinking models exhaust max_tokens on
	// reasoning before producing visible output.
	Text string `json:"text"`
	// Source carries the upstream blob for image / document blocks so
	// multimodal responses from Gemini image-generation models don't
	// collapse to text-only on the Anthropic wire.
	Source *AnthropicSource `json:"source,omitempty"`
}

// AnthropicSource mirrors the Anthropic content-block `source` object used by
// image and document blocks in both requests and responses. Inbound base64
// form fills (Type="base64", MediaType, Data); url form fills (Type="url",
// URL). The proxy uses this struct only on the response path — inbound user
// blocks are parsed ad-hoc in anthropicToOpenAIReq.
type AnthropicSource struct {
	Type      string `json:"type"`                 // "base64" or "url"
	MediaType string `json:"media_type,omitempty"` // e.g. "image/png"
	Data      string `json:"data,omitempty"`       // base64 bytes for Type="base64"
	URL       string `json:"url,omitempty"`        // HTTP URL for Type="url"
}

// AnthropicResponse: /v1/messages response
type AnthropicResponse struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Role       string             `json:"role"`
	Content    []AnthropicContent `json:"content"`
	Model      string             `json:"model"`
	StopReason string             `json:"stop_reason"`
	Usage      AnthropicUsage     `json:"usage"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicStreamEvent: SSE event for streaming /v1/messages
type AnthropicStreamEvent struct {
	Type  string      `json:"type"`
	Index int         `json:"index,omitempty"`
	Delta interface{} `json:"delta,omitempty"`
	Usage interface{} `json:"usage,omitempty"`
}
