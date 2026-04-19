// Package models: automatic model discovery and registry per service
package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Model: single model info
type Model struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Service  string `json:"service"`
	Context  int    `json:"context_length,omitempty"`
	Free     bool   `json:"free,omitempty"`
}

// Registry: full model cache
type Registry struct {
	models    []Model
	updatedAt time.Time
	ttl       time.Duration
	maxSize   int
}

// defaultRegistryMax caps the cached model list. OpenRouter alone serves
// hundreds of models and new ones land constantly; without this the cache
// can grow unbounded across refreshes, which hurts handleOpenAIModels
// response size and registry walk cost. 2000 leaves huge headroom for real
// fleets while preventing pathological blowup.
const defaultRegistryMax = 2000

func NewRegistry(ttl time.Duration) *Registry {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &Registry{ttl: ttl, maxSize: defaultRegistryMax}
}

// All: return all models (optional service filter)
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

// ServiceURLs: local service URL map (service ID → base URL)
type ServiceURLs map[string]string

// Refresh: re-fetch models from all services
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
			fetched := fetchOpenRouter(openRouterKey)
			if len(fetched) == 0 {
				// API unreachable — use curated known models as fallback
				fetched = fetchOpenRouterKnown()
			}
			all = append(all, fetched...)
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
		case "llamacpp":
			all = append(all, fetchOpenAICompat(svc, localURLs["llamacpp"], "http://localhost:8080")...)
		default:
			// custom service: try OpenAI-compatible model list if URL is present
			if u := localURLs[svc]; u != "" {
				all = append(all, fetchOpenAICompat(svc, u, "")...)
			}
		}
	}

	// Cap cache size so a buggy or unusually large upstream catalog can't
	// balloon memory indefinitely across refreshes.
	if r.maxSize > 0 && len(all) > r.maxSize {
		fmt.Fprintf(os.Stderr, "[registry] model catalog truncated: %d → %d (maxSize)\n", len(all), r.maxSize)
		all = all[:r.maxSize]
	}

	r.models = all
	r.updatedAt = time.Now()
	return nil
}

// RefreshService fetches models for a single service and upserts them into the
// cache, leaving entries for other services untouched. Called on demand from
// the dashboard edit flow so a disabled service (skipped by the TTL-gated
// full Refresh) still has model choices in its default_model dropdown.
// orKey is only read when svcID == "openrouter"; pass "" otherwise.
func (r *Registry) RefreshService(svcID, localURL, orKey string) {
	var fetched []Model
	switch svcID {
	case "google":
		fetched = fetchGoogle()
	case "openrouter":
		fetched = fetchOpenRouter(orKey)
		if len(fetched) == 0 {
			fetched = fetchOpenRouterKnown()
		}
	case "ollama":
		fetched = fetchOllama(localURL)
	case "openai":
		fetched = fetchOpenAI()
	case "anthropic":
		fetched = fetchAnthropic()
	case "github-copilot":
		fetched = fetchGitHubCopilot()
	case "lmstudio":
		fetched = fetchOpenAICompat(svcID, localURL, "http://localhost:1234")
	case "vllm":
		fetched = fetchOpenAICompat(svcID, localURL, "http://localhost:8000")
	case "llamacpp":
		fetched = fetchOpenAICompat(svcID, localURL, "http://localhost:8080")
	default:
		if localURL != "" {
			fetched = fetchOpenAICompat(svcID, localURL, "")
		}
	}
	kept := make([]Model, 0, len(r.models)+len(fetched))
	for _, m := range r.models {
		if m.Service != svcID {
			kept = append(kept, m)
		}
	}
	kept = append(kept, fetched...)
	if r.maxSize > 0 && len(kept) > r.maxSize {
		kept = kept[:r.maxSize]
	}
	r.models = kept
}

// ─── Google ──────────────────────────────────────────────────────────────────

// Google models use a fixed list (ListModels API requires authentication)
func fetchGoogle() []Model {
	return []Model{
		{ID: "gemini-3.1-pro-preview",             Name: "Gemini 3.1 Pro Preview",             Service: "google", Context: 1048576},
		{ID: "gemini-3.1-pro-preview-customtools", Name: "Gemini 3.1 Pro Preview (Custom Tools)", Service: "google", Context: 1048576},
		{ID: "gemini-3.1-flash-image-preview",     Name: "Gemini 3.1 Flash Image Preview",     Service: "google", Context: 1048576},
		{ID: "gemini-3.1-flash-lite-preview",      Name: "Gemini 3.1 Flash Lite Preview",      Service: "google", Context: 1048576},
		{ID: "gemini-3-flash-preview",             Name: "Gemini 3 Flash Preview",             Service: "google", Context: 1048576},
		{ID: "gemini-2.5-pro",                     Name: "Gemini 2.5 Pro",                     Service: "google", Context: 1048576},
		{ID: "gemini-2.5-flash",              Name: "Gemini 2.5 Flash",              Service: "google", Context: 1048576},
		{ID: "gemini-2.5-flash-8b",           Name: "Gemini 2.5 Flash 8B",           Service: "google", Context: 1048576},
		{ID: "gemini-2.0-flash",              Name: "Gemini 2.0 Flash",              Service: "google", Context: 1048576},
		{ID: "gemini-2.0-flash-lite",         Name: "Gemini 2.0 Flash Lite",         Service: "google", Context: 1048576},
		{ID: "gemini-1.5-pro",                Name: "Gemini 1.5 Pro",                Service: "google", Context: 2097152},
		{ID: "gemini-1.5-flash",              Name: "Gemini 1.5 Flash",              Service: "google", Context: 1048576},
		// Embedding (OpenClaw 3.11 memorySearch)
		{ID: "gemini-embedding-2-preview",    Name: "Gemini Embedding 2 Preview",    Service: "google", Context: 8192},
		// Gemma 4 (open models served via Gemini API)
		{ID: "gemma-4-31b-it",                Name: "Gemma 4 31B IT",                Service: "google", Context: 262144},
		{ID: "gemma-4-26b-a4b-it",            Name: "Gemma 4 26B MoE (4B active)",   Service: "google", Context: 262144},
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

// FetchOllama: auto-fetch model list from Ollama server
// if ollamaURL is empty, tries common addresses in order
func fetchOllama(ollamaURL string) []Model {
	candidates := []string{}
	if ollamaURL != "" {
		candidates = append(candidates, strings.TrimRight(ollamaURL, "/"))
	}
	// auto-detection candidates
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
	// Ollama server not responding — fallback to recommended model list
	return OllamaRecommended()
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

// NeedsRefresh: returns true if TTL has expired or never refreshed
func (r *Registry) NeedsRefresh() bool {
	return r.updatedAt.IsZero() || time.Since(r.updatedAt) >= r.ttl
}

// Search: filter models by whether model ID or name contains query (case-insensitive)
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

// FetchOllamaPublic: fetch Ollama list without configuration (for beginner setup wizard)
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
		{ID: "o3",                 Name: "o3",                 Service: "openai", Context: 200000},
		{ID: "o3-mini",            Name: "o3-mini",            Service: "openai", Context: 200000},
		{ID: "o4-mini",            Name: "o4-mini",            Service: "openai", Context: 200000},
	}
}

// ─── Notable OpenRouter models (OpenClaw 3.11) ────────────────────────────────

// fetchOpenRouterKnown returns a curated list of notable OpenRouter models
// that are shown in the UI even before a live model fetch. The full list is
// fetched dynamically from the OpenRouter API at runtime.
func fetchOpenRouterKnown() []Model {
	return []Model{
		// Free 1M-context models (OpenClaw 3.11)
		{ID: "openrouter/hunter-alpha",              Name: "Hunter Alpha (1M ctx, free)",       Service: "openrouter", Context: 1048576, Free: true},
		{ID: "openrouter/healer-alpha",              Name: "Healer Alpha (omni-modal, free)",   Service: "openrouter", Context: 1048576, Free: true},
		// Kimi / Moonshot
		{ID: "moonshot/kimi-k2.5",                  Name: "Kimi K2.5",                         Service: "openrouter", Context: 256000},
		{ID: "moonshot/kimi-k2-turbo-preview",      Name: "Kimi K2 Turbo Preview",             Service: "openrouter", Context: 256000},
		// GLM / Z.AI
		{ID: "z-ai/glm-5",                          Name: "GLM-5",                             Service: "openrouter"},
		{ID: "z-ai/glm-4.7-flash",                  Name: "GLM-4.7 Flash",                     Service: "openrouter"},
		// DeepSeek
		{ID: "deepseek/deepseek-r1",                Name: "DeepSeek R1",                       Service: "openrouter", Context: 65536},
		{ID: "deepseek/deepseek-chat",              Name: "DeepSeek V3",                       Service: "openrouter", Context: 65536},
		// Qwen / Alibaba
		{ID: "qwen/qwen-2.5-72b-instruct",          Name: "Qwen 2.5 72B",                      Service: "openrouter", Context: 131072},
		// MiniMax
		{ID: "minimax/minimax-m2.5",                Name: "MiniMax M2.5",                      Service: "openrouter"},
		// Meta Llama
		{ID: "meta-llama/llama-3.3-70b-instruct",  Name: "Llama 3.3 70B",                     Service: "openrouter", Context: 131072},
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

// ─── OpenAI-compatible (LM Studio / vLLM / custom) ───────────────────────────

// fetchOpenAICompat: fetch model list from OpenAI /v1/models compatible endpoint
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
	// fallback per service on connection failure
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
	case "llamacpp":
		return []Model{
			{ID: "local-model", Name: "Local Model (llama.cpp)", Service: "llamacpp"},
		}
	default:
		return nil
	}
}

// OllamaRecommended: local recommended Ollama model list based on OpenClaw 3.11
// (used as UI hints when Ollama server is not responding)
func OllamaRecommended() []Model {
	return []Model{
		{ID: "glm-4.7-flash",   Name: "GLM-4.7 Flash (recommended)",   Service: "ollama"},
		{ID: "qwen3.5:35b",     Name: "Qwen3.5 35B",            Service: "ollama"},
		{ID: "qwen2.5:7b",      Name: "Qwen2.5 7B",             Service: "ollama"},
		{ID: "llama3.3:70b",    Name: "Llama 3.3 70B",          Service: "ollama"},
		{ID: "llama3.3",        Name: "Llama 3.3",              Service: "ollama"},
		{ID: "deepseek-r1:7b",  Name: "DeepSeek R1 7B",         Service: "ollama"},
		{ID: "phi4",            Name: "Phi-4",                  Service: "ollama"},
	}
}

// FetchLocalModels: fetch model list from local server (for UI auto-detection)
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
