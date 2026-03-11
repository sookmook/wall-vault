package proxy

import (
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
