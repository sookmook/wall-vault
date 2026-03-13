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
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text         string      `json:"text,omitempty"`
	FunctionCall interface{} `json:"functionCall,omitempty"`
	FunctionResponse interface{} `json:"functionResponse,omitempty"`
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
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
	Index        int           `json:"index,omitempty"`
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
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// UnmarshalJSON handles content as either a string or an array of content parts.
// OpenAI spec allows both; many clients (including OpenClaw) use the array form.
func (m *OpenAIMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Role = raw.Role
	m.Content = rawContentToString(raw.Content)
	return nil
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
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Tools       []interface{}      `json:"tools,omitempty"`
	ServiceTier string             `json:"service_tier,omitempty"` // OpenClaw v2026.3.12 fast mode (strip before forwarding)
	Thinking    interface{}        `json:"thinking,omitempty"`     // Claude extended thinking (ignore)
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
	Text string `json:"text,omitempty"`
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
