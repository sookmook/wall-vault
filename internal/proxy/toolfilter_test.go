package proxy

import (
	"testing"
)

func TestFilterGemini_StripAll(t *testing.T) {
	f := NewToolFilter(FilterStripAll, nil)
	req := &GeminiRequest{
		Contents: []GeminiContent{{Role: "user", Parts: []GeminiPart{{Text: "hello"}}}},
		Tools:    []interface{}{"tool1", "tool2"},
		ToolConfig: map[string]string{"mode": "any"},
	}

	stripped := f.FilterGemini(req)
	if stripped != 2 {
		t.Fatalf("strip_all: 2개 차단 기대, got %d", stripped)
	}
	if req.Tools != nil {
		t.Fatal("strip_all 후 Tools가 nil이 아님")
	}
	if req.ToolConfig != nil {
		t.Fatal("strip_all 후 ToolConfig가 nil이 아님")
	}
}

func TestFilterGemini_Passthrough(t *testing.T) {
	f := NewToolFilter(FilterPassthrough, nil)
	tools := []interface{}{"tool1", "tool2"}
	req := &GeminiRequest{
		Tools: tools,
	}

	stripped := f.FilterGemini(req)
	if stripped != 0 {
		t.Fatalf("passthrough: 0개 차단 기대, got %d", stripped)
	}
	if len(req.Tools) != 2 {
		t.Fatal("passthrough 후 Tools 변경됨")
	}
}

func TestFilterGemini_Whitelist(t *testing.T) {
	f := NewToolFilter(FilterWhitelist, []string{"allowed_func"})

	// Gemini 도구 형식: functionDeclarations
	allowedTool := map[string]interface{}{
		"functionDeclarations": []interface{}{
			map[string]interface{}{"name": "allowed_func"},
		},
	}
	blockedTool := map[string]interface{}{
		"functionDeclarations": []interface{}{
			map[string]interface{}{"name": "blocked_func"},
		},
	}

	req := &GeminiRequest{
		Tools: []interface{}{allowedTool, blockedTool},
	}

	stripped := f.FilterGemini(req)
	if stripped != 1 {
		t.Fatalf("whitelist: 1개 차단 기대, got %d", stripped)
	}
	if len(req.Tools) != 1 {
		t.Fatalf("허용 도구 1개 기대, got %d", len(req.Tools))
	}
}

func TestFilterOpenAI_StripAll(t *testing.T) {
	f := NewToolFilter(FilterStripAll, nil)
	req := &OpenAIRequest{
		Tools:      []interface{}{"t1", "t2", "t3"},
		ToolChoice: "auto",
	}

	stripped := f.FilterOpenAI(req)
	if stripped != 3 {
		t.Fatalf("strip_all: 3개 차단 기대, got %d", stripped)
	}
	if req.Tools != nil {
		t.Fatal("Tools가 nil이 아님")
	}
	if req.ToolChoice != nil {
		t.Fatal("ToolChoice가 nil이 아님")
	}
}

func TestFilterGemini_NoTools(t *testing.T) {
	f := NewToolFilter(FilterStripAll, nil)
	req := &GeminiRequest{
		Contents: []GeminiContent{{Role: "user", Parts: []GeminiPart{{Text: "hi"}}}},
	}
	stripped := f.FilterGemini(req)
	if stripped != 0 {
		t.Fatalf("도구 없는 요청: 0 기대, got %d", stripped)
	}
}
