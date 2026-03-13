// Package models: 서비스별 모델 자동 조회 및 레지스트리
package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Model: 단일 모델 정보
type Model struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Service  string `json:"service"`
	Context  int    `json:"context_length,omitempty"`
	Free     bool   `json:"free,omitempty"`
}

// Registry: 전체 모델 캐시
type Registry struct {
	models    []Model
	updatedAt time.Time
	ttl       time.Duration
}

func NewRegistry(ttl time.Duration) *Registry {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &Registry{ttl: ttl}
}

// All: 전체 모델 반환 (서비스 필터 옵션)
func (r *Registry) All(service string) []Model {
	if service == "" {
		return r.models
	}
	out := make([]Model, 0)
	for _, m := range r.models {
		if m.Service == service {
			out = append(out, m)
		}
	}
	return out
}

// ServiceURLs: 로컬 서비스별 URL 맵 (service ID → base URL)
type ServiceURLs map[string]string

// Refresh: 모든 서비스에서 모델 재조회
func (r *Registry) Refresh(services []string, localURLs ServiceURLs, openRouterKey string) error {
	if localURLs == nil {
		localURLs = ServiceURLs{}
	}
	var all []Model

	for _, svc := range services {
		switch svc {
		case "google":
			all = append(all, fetchGoogle()...)
		case "openrouter":
			all = append(all, fetchOpenRouter(openRouterKey)...)
		case "ollama":
			all = append(all, fetchOllama(localURLs["ollama"])...)
		case "openai":
			all = append(all, fetchOpenAI()...)
		case "anthropic":
			all = append(all, fetchAnthropic()...)
		case "github-copilot":
			all = append(all, fetchGitHubCopilot()...)
		case "lmstudio":
			all = append(all, fetchOpenAICompat(svc, localURLs["lmstudio"], "http://localhost:1234")...)
		case "vllm":
			all = append(all, fetchOpenAICompat(svc, localURLs["vllm"], "http://localhost:8000")...)
		default:
			// 커스텀 서비스: URL이 있으면 OpenAI-호환 목록 조회 시도
			if u := localURLs[svc]; u != "" {
				all = append(all, fetchOpenAICompat(svc, u, "")...)
			}
		}
	}

	r.models = all
	r.updatedAt = time.Now()
	return nil
}

// ─── Google ──────────────────────────────────────────────────────────────────

// Google 모델은 고정 목록 사용 (ListModels API 인증 필요)
func fetchGoogle() []Model {
	return []Model{
		{ID: "gemini-2.5-pro",                Name: "Gemini 2.5 Pro",                Service: "google", Context: 1048576},
		{ID: "gemini-2.5-flash",              Name: "Gemini 2.5 Flash",              Service: "google", Context: 1048576},
		{ID: "gemini-2.5-flash-8b",           Name: "Gemini 2.5 Flash 8B",           Service: "google", Context: 1048576},
		{ID: "gemini-2.0-flash",              Name: "Gemini 2.0 Flash",              Service: "google", Context: 1048576},
		{ID: "gemini-2.0-flash-lite",         Name: "Gemini 2.0 Flash Lite",         Service: "google", Context: 1048576},
		{ID: "gemini-1.5-pro",                Name: "Gemini 1.5 Pro",                Service: "google", Context: 2097152},
		{ID: "gemini-1.5-flash",              Name: "Gemini 1.5 Flash",              Service: "google", Context: 1048576},
		// Embedding (OpenClaw 3.11 memorySearch)
		{ID: "gemini-embedding-2-preview",    Name: "Gemini Embedding 2 Preview",    Service: "google", Context: 8192},
	}
}

// ─── OpenRouter ───────────────────────────────────────────────────────────────

type openRouterResp struct {
	Data []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ContextLength int    `json:"context_length"`
		Pricing       struct {
			Prompt string `json:"prompt"`
		} `json:"pricing"`
	} `json:"data"`
}

func fetchOpenRouter(apiKey string) []Model {
	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result openRouterResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	out := make([]Model, 0, len(result.Data))
	for _, m := range result.Data {
		free := m.Pricing.Prompt == "0"
		out = append(out, Model{
			ID:      m.ID,
			Name:    m.Name,
			Service: "openrouter",
			Context: m.ContextLength,
			Free:    free,
		})
	}
	return out
}

// ─── Ollama ───────────────────────────────────────────────────────────────────

type ollamaTagsResp struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// FetchOllama: Ollama 서버에서 모델 목록 자동 조회
// ollamaURL 이 빈 문자열이면 일반적인 주소들을 순서대로 시도
func fetchOllama(ollamaURL string) []Model {
	candidates := []string{}
	if ollamaURL != "" {
		candidates = append(candidates, strings.TrimRight(ollamaURL, "/"))
	}
	// 자동 탐지 후보
	candidates = append(candidates,
		"http://localhost:11434",
		"http://127.0.0.1:11434",
	)

	client := &http.Client{Timeout: 5 * time.Second}
	for _, base := range candidates {
		models, err := tryFetchOllama(client, base)
		if err == nil {
			return models
		}
	}
	return nil
}

func tryFetchOllama(client *http.Client, base string) ([]Model, error) {
	resp, err := client.Get(fmt.Sprintf("%s/api/tags", base))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result ollamaTagsResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	out := make([]Model, 0, len(result.Models))
	for _, m := range result.Models {
		out = append(out, Model{
			ID:      m.Name,
			Name:    m.Name,
			Service: "ollama",
		})
	}
	return out, nil
}

// NeedsRefresh: TTL이 만료됐거나 한 번도 갱신되지 않았으면 true
func (r *Registry) NeedsRefresh() bool {
	return r.updatedAt.IsZero() || time.Since(r.updatedAt) >= r.ttl
}

// Search: 모델 ID·이름에서 query(대소문자 무시) 포함 여부로 필터링
func (r *Registry) Search(query string) []Model {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return r.models
	}
	out := make([]Model, 0)
	for _, m := range r.models {
		if strings.Contains(strings.ToLower(m.ID), q) ||
			strings.Contains(strings.ToLower(m.Name), q) {
			out = append(out, m)
		}
	}
	return out
}

// FetchOllamaPublic: 설정 없이도 Ollama 목록 조회 (초보자용 setup wizard)
func FetchOllamaPublic(ollamaURL string) ([]Model, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	base := strings.TrimRight(ollamaURL, "/")
	return tryFetchOllama(client, base)
}

// ─── OpenAI ──────────────────────────────────────────────────────────────────

func fetchOpenAI() []Model {
	return []Model{
		{ID: "gpt-4o",              Name: "GPT-4o",             Service: "openai", Context: 128000},
		{ID: "gpt-4o-mini",        Name: "GPT-4o mini",        Service: "openai", Context: 128000},
		{ID: "gpt-4-turbo",        Name: "GPT-4 Turbo",        Service: "openai", Context: 128000},
		{ID: "gpt-3.5-turbo",      Name: "GPT-3.5 Turbo",      Service: "openai", Context: 16385},
		{ID: "o1",                 Name: "o1",                 Service: "openai", Context: 200000},
		{ID: "o1-mini",            Name: "o1-mini",            Service: "openai", Context: 128000},
		{ID: "o3-mini",            Name: "o3-mini",            Service: "openai", Context: 200000},
		{ID: "o4-mini",            Name: "o4-mini",            Service: "openai", Context: 200000},
	}
}

// ─── Anthropic ───────────────────────────────────────────────────────────────

func fetchAnthropic() []Model {
	return []Model{
		{ID: "claude-opus-4-6",               Name: "Claude Opus 4.6",              Service: "anthropic", Context: 200000},
		{ID: "claude-sonnet-4-6",             Name: "Claude Sonnet 4.6",            Service: "anthropic", Context: 200000},
		{ID: "claude-haiku-4-5-20251001",     Name: "Claude Haiku 4.5",            Service: "anthropic", Context: 200000},
		{ID: "claude-3-5-sonnet-20241022",    Name: "Claude 3.5 Sonnet",           Service: "anthropic", Context: 200000},
		{ID: "claude-3-5-haiku-20241022",     Name: "Claude 3.5 Haiku",            Service: "anthropic", Context: 200000},
		{ID: "claude-3-opus-20240229",        Name: "Claude 3 Opus",               Service: "anthropic", Context: 200000},
	}
}

// ─── GitHub Copilot ──────────────────────────────────────────────────────────

func fetchGitHubCopilot() []Model {
	return []Model{
		{ID: "gpt-4o",            Name: "GPT-4o",        Service: "github-copilot"},
		{ID: "gpt-4o-mini",      Name: "GPT-4o mini",   Service: "github-copilot"},
		{ID: "o1-preview",       Name: "o1-preview",    Service: "github-copilot"},
		{ID: "o1-mini",          Name: "o1-mini",       Service: "github-copilot"},
		{ID: "o3-mini",          Name: "o3-mini",       Service: "github-copilot"},
		{ID: "claude-sonnet-4-6",Name: "Claude Sonnet 4.6", Service: "github-copilot"},
	}
}

// ─── OpenAI-compatible (LM Studio / vLLM / 커스텀) ──────────────────────────

// fetchOpenAICompat: OpenAI /v1/models 호환 엔드포인트에서 모델 목록 조회
func fetchOpenAICompat(service, primaryURL, fallbackURL string) []Model {
	candidates := []string{}
	if primaryURL != "" {
		candidates = append(candidates, strings.TrimRight(primaryURL, "/"))
	}
	if fallbackURL != "" {
		candidates = append(candidates, strings.TrimRight(fallbackURL, "/"))
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, base := range candidates {
		models, err := tryFetchOpenAICompat(client, service, base)
		if err == nil {
			return models
		}
	}
	// 연결 실패 시 서비스별 폴백
	return compatFallback(service)
}

type openAIModelsResp struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func tryFetchOpenAICompat(client *http.Client, service, base string) ([]Model, error) {
	resp, err := client.Get(fmt.Sprintf("%s/v1/models", base))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result openAIModelsResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	out := make([]Model, 0, len(result.Data))
	for _, m := range result.Data {
		out = append(out, Model{ID: m.ID, Name: m.ID, Service: service})
	}
	return out, nil
}

func compatFallback(service string) []Model {
	switch service {
	case "lmstudio":
		return []Model{
			{ID: "local-model", Name: "Local Model (LM Studio)", Service: "lmstudio"},
		}
	case "vllm":
		return []Model{
			{ID: "local-model", Name: "Local Model (vLLM)", Service: "vllm"},
		}
	default:
		return nil
	}
}

// FetchLocalModels: 로컬 서버에서 모델 목록 조회 (UI 자동 감지용)
func FetchLocalModels(service, serverURL string) ([]Model, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	base := strings.TrimRight(serverURL, "/")
	switch service {
	case "ollama":
		return tryFetchOllama(client, base)
	default: // lmstudio, vllm, custom
		return tryFetchOpenAICompat(client, service, base)
	}
}
