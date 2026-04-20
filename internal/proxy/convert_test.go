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

// TestOpenAIMessage_MarshalJSON_PreservesMultiPart guards the outbound
// multimodal path: UnmarshalJSON parks multi-part content in RawContent, and
// MarshalJSON must re-emit that array so upstream OpenAI-compat servers
// (Ollama, lmstudio, vllm, llamacpp, OpenAI direct, OpenRouter) see the
// image_url / input_audio parts the client sent. Regression guard for the
// bug where multimodal content silently turned into text-only on the wire.
func TestOpenAIMessage_MarshalJSON_PreservesMultiPart(t *testing.T) {
	inbound := []byte(`{
		"role":"user",
		"content":[
			{"type":"text","text":"what's in this image?"},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw=="}}
		]
	}`)
	var m OpenAIMessage
	if err := json.Unmarshal(inbound, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(m.RawContent) == 0 {
		t.Fatal("RawContent should be populated for multi-part array content")
	}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Parse back and assert content is an array with the original parts.
	var decoded struct {
		Role    string            `json:"role"`
		Content []json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("decode roundtrip: %v (out=%s)", err, out)
	}
	if decoded.Role != "user" {
		t.Errorf("role = %q, want user", decoded.Role)
	}
	if len(decoded.Content) != 2 {
		t.Fatalf("content parts = %d, want 2 (out=%s)", len(decoded.Content), out)
	}
	// First part: text
	var p1 struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(decoded.Content[0], &p1); err != nil || p1.Type != "text" || p1.Text == "" {
		t.Errorf("part[0] type=%q text=%q err=%v", p1.Type, p1.Text, err)
	}
	// Second part: image_url with data URI preserved
	var p2 struct {
		Type     string `json:"type"`
		ImageURL struct {
			URL string `json:"url"`
		} `json:"image_url"`
	}
	if err := json.Unmarshal(decoded.Content[1], &p2); err != nil {
		t.Fatalf("part[1] decode: %v", err)
	}
	if p2.Type != "image_url" {
		t.Errorf("part[1] type = %q, want image_url", p2.Type)
	}
	if p2.ImageURL.URL != "data:image/png;base64,iVBORw==" {
		t.Errorf("part[1] url = %q, want data URI preserved", p2.ImageURL.URL)
	}
}

// TestOpenAIMessage_MarshalJSON_AssistantEmptyGuard ensures the earlier
// guard for assistant Role with empty Content (added for Claude Code's
// .trim() crash) still works when RawContent is absent.
func TestOpenAIMessage_MarshalJSON_AssistantEmptyGuard(t *testing.T) {
	m := OpenAIMessage{Role: "assistant", Content: ""}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := decoded["content"]; !ok {
		t.Fatalf("assistant empty-content guard missing content key: %s", out)
	}
	if decoded["content"] != "" {
		t.Errorf("assistant empty-content: got %v, want \"\"", decoded["content"])
	}
}

// TestAnthropicToGemini_ImageBlockBase64 verifies that an Anthropic user
// message with an `image` content block (base64 source) survives the
// Anthropic → OpenAI → Gemini pipeline and lands in the final GeminiRequest
// as InlineData. Previously image/document blocks were silently dropped.
func TestAnthropicToGemini_ImageBlockBase64(t *testing.T) {
	req := &AnthropicRequest{
		Model: "claude-haiku-4-5-20251001",
		Messages: []AnthropicMessage{{
			Role: "user",
			Content: json.RawMessage(`[
				{"type":"text","text":"describe it"},
				{"type":"image","source":{"type":"base64","media_type":"image/png","data":"iVBORw=="}}
			]`),
		}},
	}
	gem := AnthropicToGemini(req)
	var foundImage bool
	var foundText bool
	for _, c := range gem.Contents {
		for _, p := range c.Parts {
			if p.Text == "describe it" {
				foundText = true
			}
			if p.InlineData != nil && p.InlineData.MimeType == "image/png" && p.InlineData.Data == "iVBORw==" {
				foundImage = true
			}
		}
	}
	if !foundText {
		t.Error("text block not preserved through conversion")
	}
	if !foundImage {
		t.Error("image block (base64) not converted to InlineData")
	}
}

// TestAnthropicToGemini_ImageBlockURL verifies the url-source form of
// Anthropic image blocks. The URL is threaded through as image_url so the
// proxy's fetch-and-inline path can resolve it (best-effort).
func TestAnthropicToGemini_ImageBlockURL(t *testing.T) {
	url := anthropicImageSourceToURL(json.RawMessage(`{"type":"url","url":"https://example.com/a.png"}`))
	if url != "https://example.com/a.png" {
		t.Fatalf("url form: got %q", url)
	}
	if got := anthropicImageSourceToURL(json.RawMessage(`{"type":"base64","media_type":"image/jpeg","data":"/9j/4AAQ"}`)); got != "data:image/jpeg;base64,/9j/4AAQ" {
		t.Fatalf("base64 form: got %q", got)
	}
	if got := anthropicImageSourceToURL(json.RawMessage(`{"type":"base64","media_type":"image/png","data":""}`)); got != "" {
		t.Fatalf("empty data must yield empty URL, got %q", got)
	}
}

// TestAnthropicDocumentSourceToPart verifies document blocks map to
// input_file (for base64 PDFs) or image_url (for URL form).
func TestAnthropicDocumentSourceToPart(t *testing.T) {
	part := anthropicDocumentSourceToPart(json.RawMessage(`{"type":"base64","media_type":"application/pdf","data":"JVBERi0="}`))
	if part == nil || part["type"] != "input_file" || part["mime"] != "application/pdf" || part["data"] != "JVBERi0=" {
		t.Fatalf("base64 pdf: %+v", part)
	}
	part = anthropicDocumentSourceToPart(json.RawMessage(`{"type":"url","url":"https://example.com/doc.pdf"}`))
	if part == nil || part["type"] != "image_url" {
		t.Fatalf("url form: %+v", part)
	}
	if anthropicDocumentSourceToPart(json.RawMessage(`{"type":"unknown"}`)) != nil {
		t.Fatal("unknown source type must return nil")
	}
}

// TestGeminiRespToAnthropic_PreservesInlineData verifies that an image
// returned by a Gemini candidate (e.g. gemini-3.1-flash-image-preview) is
// emitted as a proper Anthropic `image` content block instead of collapsing
// to text-only.
func TestGeminiRespToAnthropic_PreservesInlineData(t *testing.T) {
	resp := &GeminiResponse{
		Candidates: []GeminiCandidate{{
			Content: GeminiContent{
				Parts: []GeminiPart{
					{Text: "here's the result"},
					{InlineData: &BlobData{MimeType: "image/png", Data: "iVBORw=="}},
				},
			},
		}},
	}
	ar := GeminiRespToAnthropic("claude-haiku-4-5-20251001", resp)
	if len(ar.Content) != 2 {
		t.Fatalf("content blocks = %d, want 2 (text + image): %+v", len(ar.Content), ar.Content)
	}
	if ar.Content[0].Type != "text" || ar.Content[0].Text != "here's the result" {
		t.Errorf("block[0] type=%q text=%q", ar.Content[0].Type, ar.Content[0].Text)
	}
	img := ar.Content[1]
	if img.Type != "image" || img.Source == nil {
		t.Fatalf("block[1] type=%q source=%v", img.Type, img.Source)
	}
	if img.Source.Type != "base64" || img.Source.MediaType != "image/png" || img.Source.Data != "iVBORw==" {
		t.Errorf("image source: %+v", img.Source)
	}
}

// TestExtractTextAndMediaNotes verifies that on the OpenAI response path
// (which can't carry binary in its string-only `content` field) InlineData
// yields an inline placeholder note rather than disappearing silently.
func TestExtractTextAndMediaNotes(t *testing.T) {
	got := extractTextAndMediaNotes([]GeminiPart{
		{Text: "header"},
		{InlineData: &BlobData{MimeType: "image/jpeg", Data: "/9j/4AAQSkZJRg=="}},
		{Text: "footer"},
	})
	if got == "" {
		t.Fatal("result must not be empty")
	}
	if !contains(got, "header") || !contains(got, "footer") {
		t.Errorf("text parts missing: %q", got)
	}
	if !contains(got, "image/jpeg") || !contains(got, "media attached") {
		t.Errorf("media note missing: %q", got)
	}
	// Plain text-only input keeps original behaviour (no noise).
	plain := extractTextAndMediaNotes([]GeminiPart{{Text: "just text"}})
	if plain != "just text" {
		t.Errorf("text-only regression: got %q", plain)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
