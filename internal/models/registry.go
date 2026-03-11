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

// Refresh: 모든 서비스에서 모델 재조회
func (r *Registry) Refresh(services []string, ollamaURL string, openRouterKey string) error {
	var all []Model

	for _, svc := range services {
		switch svc {
		case "google":
			ms := fetchGoogle()
			all = append(all, ms...)
		case "openrouter":
			ms := fetchOpenRouter(openRouterKey)
			all = append(all, ms...)
		case "ollama":
			ms := fetchOllama(ollamaURL)
			all = append(all, ms...)
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
		{ID: "gemini-2.5-pro",          Name: "Gemini 2.5 Pro",          Service: "google", Context: 1048576},
		{ID: "gemini-2.5-flash",        Name: "Gemini 2.5 Flash",        Service: "google", Context: 1048576},
		{ID: "gemini-2.0-flash",        Name: "Gemini 2.0 Flash",        Service: "google", Context: 1048576},
		{ID: "gemini-2.0-flash-lite",   Name: "Gemini 2.0 Flash Lite",   Service: "google", Context: 1048576},
		{ID: "gemini-1.5-pro",          Name: "Gemini 1.5 Pro",          Service: "google", Context: 2097152},
		{ID: "gemini-1.5-flash",        Name: "Gemini 1.5 Flash",        Service: "google", Context: 1048576},
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

// FetchOllamaPublic: 설정 없이도 Ollama 목록 조회 (초보자용 setup wizard)
func FetchOllamaPublic(ollamaURL string) ([]Model, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	base := strings.TrimRight(ollamaURL, "/")
	return tryFetchOllama(client, base)
}
