package proxy

import "time"

// ─── 현재 설정 ────────────────────────────────────────────────────────────────

type RuntimeConfig struct {
	Service   string
	Model     string
	ClientID  string
	VaultURL  string
	Token     string
	UpdatedAt time.Time
}

// ─── Gemini 요청/응답 구조 ───────────────────────────────────────────────────

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

// ─── OpenAI 요청/응답 구조 ───────────────────────────────────────────────────

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Tools       []interface{}   `json:"tools,omitempty"`
	ToolChoice  interface{}     `json:"tool_choice,omitempty"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

// ─── Ollama 요청/응답 구조 ───────────────────────────────────────────────────

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
