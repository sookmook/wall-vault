package proxy

import (
	"encoding/json"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestGeminiToOpenAI_Basic(t *testing.T) {
	req := &GeminiRequest{
		Contents: []GeminiContent{
			{Role: "user", Parts: []GeminiPart{{Text: "안녕하세요"}}},
			{Role: "model", Parts: []GeminiPart{{Text: "안녕하세요!"}}},
			{Role: "user", Parts: []GeminiPart{{Text: "잘 지내요?"}}},
		},
	}

	oai := GeminiToOpenAI("gemini-2.5-flash", req)
	if oai.Model != "gemini-2.5-flash" {
		t.Fatalf("모델 불일치: %q", oai.Model)
	}
	if len(oai.Messages) != 3 {
		t.Fatalf("메시지 수 기대 3, got %d", len(oai.Messages))
	}
	if oai.Messages[1].Role != "assistant" {
		t.Fatalf("model→assistant 변환 실패: %q", oai.Messages[1].Role)
	}
}

func TestGeminiToOpenAI_SystemInstruction(t *testing.T) {
	req := &GeminiRequest{
		SystemInstruction: &GeminiContent{
			Parts: []GeminiPart{{Text: "너는 친절한 어시스턴트야"}},
		},
		Contents: []GeminiContent{
			{Role: "user", Parts: []GeminiPart{{Text: "hello"}}},
		},
	}

	oai := GeminiToOpenAI("model", req)
	if len(oai.Messages) != 2 {
		t.Fatalf("시스템+유저 2개 기대, got %d", len(oai.Messages))
	}
	if oai.Messages[0].Role != "system" {
		t.Fatalf("첫 번째 메시지가 system이 아님: %q", oai.Messages[0].Role)
	}
}

func TestOpenAIToGemini_Basic(t *testing.T) {
	req := &OpenAIRequest{
		Messages: []OpenAIMessage{
			{Role: "system", Content: "시스템 프롬프트"},
			{Role: "user", Content: "사용자 메시지"},
			{Role: "assistant", Content: "어시스턴트 응답"},
		},
	}

	gemini := OpenAIToGemini(req)
	if gemini.SystemInstruction == nil {
		t.Fatal("SystemInstruction이 nil")
	}
	if gemini.SystemInstruction.Parts[0].Text != "시스템 프롬프트" {
		t.Fatalf("시스템 프롬프트 불일치: %q", gemini.SystemInstruction.Parts[0].Text)
	}
	if len(gemini.Contents) != 2 {
		t.Fatalf("대화 2개 기대, got %d", len(gemini.Contents))
	}
	if gemini.Contents[1].Role != "model" {
		t.Fatalf("assistant→model 변환 실패: %q", gemini.Contents[1].Role)
	}
}

func TestOpenAIRespToGemini(t *testing.T) {
	resp := &OpenAIResponse{
		Choices: []OpenAIChoice{
			{Message: OpenAIMessage{Role: "assistant", Content: "응답"}, FinishReason: "stop"},
		},
		Usage: &OpenAIUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}

	gr := OpenAIRespToGemini(resp)
	if len(gr.Candidates) != 1 {
		t.Fatalf("후보 1개 기대, got %d", len(gr.Candidates))
	}
	if gr.Candidates[0].FinishReason != "STOP" {
		t.Fatalf("FinishReason 대문자 기대, got %q", gr.Candidates[0].FinishReason)
	}
	if gr.UsageMetadata == nil || gr.UsageMetadata.TotalTokenCount != 15 {
		t.Fatal("UsageMetadata 불일치")
	}
}

func TestGeminiToOllama(t *testing.T) {
	temp := 0.7
	maxTok := 1024
	req := &GeminiRequest{
		Contents: []GeminiContent{
			{Role: "user", Parts: []GeminiPart{{Text: "테스트"}}},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:     &temp,
			MaxOutputTokens: &maxTok,
		},
	}

	ollama := GeminiToOllama("qwen3.5:35b", req)
	if ollama.Model != "qwen3.5:35b" {
		t.Fatalf("모델 불일치: %q", ollama.Model)
	}
	if ollama.Options == nil {
		t.Fatal("Options가 nil")
	}
	if ollama.Options.Temperature != 0.7 {
		t.Fatalf("Temperature 불일치: %v", ollama.Options.Temperature)
	}
	if ollama.Options.NumPredict != 1024 {
		t.Fatalf("NumPredict 불일치: %v", ollama.Options.NumPredict)
	}
}

func TestOllamaRespToGemini(t *testing.T) {
	resp := &OllamaResponse{
		Message: OpenAIMessage{Role: "assistant", Content: "Ollama 응답"},
		Done:    true,
	}
	gr := OllamaRespToGemini(resp)
	if len(gr.Candidates) != 1 {
		t.Fatalf("후보 1개 기대, got %d", len(gr.Candidates))
	}
	if gr.Candidates[0].Content.Parts[0].Text != "Ollama 응답" {
		t.Fatalf("응답 내용 불일치: %q", gr.Candidates[0].Content.Parts[0].Text)
	}
	if gr.Candidates[0].FinishReason != "STOP" {
		t.Fatalf("FinishReason 불일치: %q", gr.Candidates[0].FinishReason)
	}
}

// ─── Multimodal pass-through ────────────────────────────────────────────────

func TestOpenAIToGemini_TextOnly_Unchanged(t *testing.T) {
	req := &OpenAIRequest{
		Model: "gemini-3.1-flash",
		Messages: []OpenAIMessage{
			{Role: "user", Content: "hello"},
		},
	}
	g := OpenAIToGemini(req)
	if len(g.Contents) != 1 || len(g.Contents[0].Parts) != 1 {
		t.Fatalf("want 1 content / 1 part, got %d / %d", len(g.Contents), len(g.Contents[0].Parts))
	}
	if g.Contents[0].Parts[0].Text != "hello" {
		t.Fatalf("text mismatch: %q", g.Contents[0].Parts[0].Text)
	}
	if g.Contents[0].Parts[0].InlineData != nil {
		t.Fatalf("text-only must not produce InlineData")
	}
}

func TestOpenAIToGemini_InputAudio(t *testing.T) {
	body := []byte(`{
		"model": "gemini-3.1-flash",
		"messages": [{
			"role": "user",
			"content": [
				{"type":"text","text":"이 음성 들어봐"},
				{"type":"input_audio","input_audio":{"data":"YWJjZA==","format":"wav"}}
			]
		}]
	}`)
	var req OpenAIRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("decode: %v", err)
	}
	g := OpenAIToGemini(&req)
	if len(g.Contents) != 1 {
		t.Fatalf("want 1 content, got %d", len(g.Contents))
	}
	parts := g.Contents[0].Parts
	if len(parts) != 2 {
		t.Fatalf("want 2 parts, got %d", len(parts))
	}
	if parts[0].Text != "이 음성 들어봐" {
		t.Fatalf("text part: %q", parts[0].Text)
	}
	if parts[1].InlineData == nil {
		t.Fatalf("audio part missing InlineData")
	}
	if parts[1].InlineData.MimeType != "audio/wav" || parts[1].InlineData.Data != "YWJjZA==" {
		t.Fatalf("audio inline mismatch: mime=%q data=%q", parts[1].InlineData.MimeType, parts[1].InlineData.Data)
	}
}

func TestOpenAIToGemini_ImageDataURI(t *testing.T) {
	body := []byte(`{
		"model": "gemini-3.1-pro",
		"messages": [{
			"role": "user",
			"content": [
				{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0KGgo="}}
			]
		}]
	}`)
	var req OpenAIRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("decode: %v", err)
	}
	g := OpenAIToGemini(&req)
	if len(g.Contents) != 1 || len(g.Contents[0].Parts) != 1 {
		t.Fatalf("want 1/1, got %d/%d", len(g.Contents), len(g.Contents[0].Parts))
	}
	p := g.Contents[0].Parts[0]
	if p.InlineData == nil || p.InlineData.MimeType != "image/png" || p.InlineData.Data != "iVBORw0KGgo=" {
		t.Fatalf("image inline mismatch: %+v", p.InlineData)
	}
}

func TestOpenAIToGemini_VideoDataURI(t *testing.T) {
	body := []byte(`{
		"model": "gemini-3.1-pro",
		"messages": [{
			"role": "user",
			"content": [
				{"type":"image_url","image_url":{"url":"data:video/mp4;base64,AAAA"}}
			]
		}]
	}`)
	var req OpenAIRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("decode: %v", err)
	}
	g := OpenAIToGemini(&req)
	p := g.Contents[0].Parts[0]
	if p.InlineData == nil || p.InlineData.MimeType != "video/mp4" {
		t.Fatalf("video inline via image_url data URI failed: %+v", p.InlineData)
	}
}

func TestOpenAIToGemini_InputVideoExplicit(t *testing.T) {
	body := []byte(`{
		"model": "gemini-3.1-pro",
		"messages": [{
			"role": "user",
			"content": [
				{"type":"input_video","input_video":{"data":"BBBB","format":"webm"}}
			]
		}]
	}`)
	var req OpenAIRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("decode: %v", err)
	}
	g := OpenAIToGemini(&req)
	p := g.Contents[0].Parts[0]
	if p.InlineData == nil || p.InlineData.MimeType != "video/webm" || p.InlineData.Data != "BBBB" {
		t.Fatalf("input_video failed: %+v", p.InlineData)
	}
}

func TestAudioFormatToMime(t *testing.T) {
	cases := map[string]string{
		"wav":  "audio/wav",
		"mp3":  "audio/mpeg",
		"ogg":  "audio/ogg",
		"flac": "audio/flac",
		"webm": "audio/webm",
		"m4a":  "audio/mp4",
		"":     "audio/wav",
		"opus": "audio/opus", // unknown → audio/<as-is>
	}
	for in, want := range cases {
		if got := audioFormatToMime(in); got != want {
			t.Errorf("audioFormatToMime(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseDataURI(t *testing.T) {
	mime, data, ok := parseDataURI("data:image/png;base64,iVBORw==")
	if !ok || mime != "image/png" || data != "iVBORw==" {
		t.Fatalf("png: ok=%v mime=%q data=%q", ok, mime, data)
	}
	if _, _, ok := parseDataURI("https://example.com/x.png"); ok {
		t.Fatalf("non-data URI must return ok=false")
	}
	if _, _, ok := parseDataURI("data:text/plain,hello"); ok {
		t.Fatalf("non-base64 data URI must return ok=false")
	}
}
