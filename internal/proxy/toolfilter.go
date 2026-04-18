package proxy

import "encoding/json"

// FilterMode: tool filter mode
type FilterMode string

const (
	FilterStripAll   FilterMode = "strip_all"   // block all external tools
	FilterWhitelist  FilterMode = "whitelist"   // only allowed list passes through
	FilterPassthrough FilterMode = "passthrough" // no filter
)

// ToolFilter: remove tools from Gemini / OpenAI requests
type ToolFilter struct {
	mode         FilterMode
	allowedTools map[string]bool
}

func NewToolFilter(mode FilterMode, allowed []string) *ToolFilter {
	m := make(map[string]bool, len(allowed))
	for _, t := range allowed {
		m[t] = true
	}
	return &ToolFilter{mode: mode, allowedTools: m}
}

// FilterGemini: handle tools/toolConfig in Gemini requests
func (f *ToolFilter) FilterGemini(req *GeminiRequest) int {
	if f.mode == FilterPassthrough {
		return 0
	}
	stripped := 0
	if req.Tools != nil {
		if f.mode == FilterStripAll {
			stripped = len(req.Tools)
			req.Tools = nil
			req.ToolConfig = nil
		} else if f.mode == FilterWhitelist {
			var kept []interface{}
			for _, t := range req.Tools {
				if f.toolAllowed(t) {
					kept = append(kept, t)
				} else {
					stripped++
				}
			}
			req.Tools = kept
			if len(kept) == 0 {
				req.ToolConfig = nil
			}
		}
	}
	return stripped
}

// FilterOpenAI: handle tools/tool_choice in OpenAI requests
func (f *ToolFilter) FilterOpenAI(req *OpenAIRequest) int {
	if f.mode == FilterPassthrough {
		return 0
	}
	stripped := 0
	if req.Tools != nil {
		if f.mode == FilterStripAll {
			stripped = len(req.Tools)
			req.Tools = nil
			req.ToolChoice = nil
		} else if f.mode == FilterWhitelist {
			var kept []interface{}
			for _, t := range req.Tools {
				if f.toolAllowed(t) {
					kept = append(kept, t)
				} else {
					stripped++
				}
			}
			req.Tools = kept
			if len(kept) == 0 {
				req.ToolChoice = nil
			}
		}
	}
	return stripped
}

// FilterAnthropic: handle tools in Anthropic (/v1/messages) requests.
// Anthropic tool shape: {"name": "...", "description": "...", "input_schema": {...}}
func (f *ToolFilter) FilterAnthropic(req *AnthropicRequest) int {
	if f.mode == FilterPassthrough {
		return 0
	}
	stripped := 0
	if req.Tools != nil {
		if f.mode == FilterStripAll {
			stripped = len(req.Tools)
			req.Tools = nil
		} else if f.mode == FilterWhitelist {
			var kept []interface{}
			for _, t := range req.Tools {
				if f.toolAllowed(t) {
					kept = append(kept, t)
				} else {
					stripped++
				}
			}
			req.Tools = kept
		}
	}
	return stripped
}

func (f *ToolFilter) toolAllowed(t interface{}) bool {
	if len(f.allowedTools) == 0 {
		return false
	}
	// serialize to JSON then extract name field
	data, err := json.Marshal(t)
	if err != nil {
		return false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}
	// Gemini: {"functionDeclarations": [{"name": "..."}]}
	if decls, ok := m["functionDeclarations"].([]interface{}); ok {
		for _, d := range decls {
			if dm, ok := d.(map[string]interface{}); ok {
				if name, ok := dm["name"].(string); ok && f.allowedTools[name] {
					return true
				}
			}
		}
	}
	// OpenAI: {"type": "function", "function": {"name": "..."}}
	if fn, ok := m["function"].(map[string]interface{}); ok {
		if name, ok := fn["name"].(string); ok && f.allowedTools[name] {
			return true
		}
	}
	// Anthropic: {"name": "...", "description": "...", "input_schema": {...}}
	if name, ok := m["name"].(string); ok && f.allowedTools[name] {
		return true
	}
	return false
}
