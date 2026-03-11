package proxy

import "encoding/json"

// FilterMode: 도구 필터 모드
type FilterMode string

const (
	FilterStripAll   FilterMode = "strip_all"   // 모든 외부 도구 차단
	FilterWhitelist  FilterMode = "whitelist"   // 허용 목록만 통과
	FilterPassthrough FilterMode = "passthrough" // 필터 없음
)

// ToolFilter: Gemini / OpenAI 요청에서 도구 제거
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

// FilterGemini: Gemini 요청에서 tools/toolConfig 처리
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

// FilterOpenAI: OpenAI 요청에서 tools/tool_choice 처리
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

func (f *ToolFilter) toolAllowed(t interface{}) bool {
	if len(f.allowedTools) == 0 {
		return false
	}
	// JSON 직렬화 후 name 필드 추출
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
	return false
}
