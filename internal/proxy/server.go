package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/hooks"
	"github.com/sookmook/wall-vault/internal/middleware"
	"github.com/sookmook/wall-vault/internal/models"
)

// Version is set from main via ldflags injection: -ldflags "-X main.version=..."
// main.go calls proxy.Version = version before starting the server.
var Version = "dev"

// tokenCacheEntry: cached result of a token→model lookup from the vault
type tokenCacheEntry struct {
	service   string
	model     string
	expiresAt time.Time
}

// Server: proxy HTTP server
type Server struct {
	cfg             *config.Config
	mu              sync.RWMutex
	service         string
	model           string
	allowedServices []string          // proxy-enabled services from vault (empty = no restriction)
	serviceURLs     map[string]string // service ID → local URL from vault config
	keyMgr          *KeyManager
	filter          *ToolFilter
	sse             *SSEClient
	registry        *models.Registry
	hooksMgr        *hooks.Manager
	ollamaMu        sync.Mutex // protect single concurrent Ollama request
	tokenCacheMu    sync.RWMutex
	tokenCache      map[string]*tokenCacheEntry // Bearer token → client model config
}

func NewServer(cfg *config.Config) *Server {
	// determine default service
	defaultSvc := "ollama"
	if len(cfg.Proxy.Services) > 0 {
		defaultSvc = cfg.Proxy.Services[0]
	}

	s := &Server{
		cfg:        cfg,
		service:    defaultSvc,
		model:      "",
		registry:   models.NewRegistry(10 * time.Minute),
		tokenCache: make(map[string]*tokenCacheEntry),
	}

	s.keyMgr = NewKeyManager(cfg.Proxy.VaultURL, cfg.Proxy.VaultToken, cfg.Proxy.ClientID)
	s.filter = NewToolFilter(FilterMode(cfg.Proxy.ToolFilter), cfg.Proxy.AllowedTools)

	// initialize hooks manager
	shellCmds := map[hooks.EventType]string{
		hooks.EventModelChanged: cfg.Hooks.OnModelChange,
		hooks.EventKeyExhausted: cfg.Hooks.OnKeyExhausted,
		hooks.EventServiceDown:  cfg.Hooks.OnServiceDown,
		hooks.EventDoctorFix:    cfg.Hooks.OnDoctorFix,
	}
	s.hooksMgr = hooks.NewManager(shellCmds, cfg.Hooks.OpenClawSocket)

	// load keys from env vars (standalone mode)
	s.keyMgr.LoadFromEnv()

	// distributed mode: sync keys from vault
	if cfg.Proxy.VaultURL != "" {
		s.sse = NewSSEClient(cfg.Proxy.VaultURL, cfg.Proxy.ClientID, func(svc, mdl string) {
			s.mu.Lock()
			oldSvc, oldMdl := s.service, s.model
			if svc != "" {
				s.service = svc
			}
			if mdl != "" {
				s.model = mdl
			}
			newSvc, newMdl := s.service, s.model
			s.mu.Unlock()
			if newSvc != oldSvc || newMdl != oldMdl {
				s.hooksMgr.Fire(hooks.EventModelChanged, map[string]string{
					"service": newSvc,
					"model":   newMdl,
				})
				go updateOpenClawJSON(newSvc, newMdl)
			}
		}, func() {
			// Flush token cache so vault model changes take effect immediately
			s.tokenCacheMu.Lock()
			s.tokenCache = make(map[string]*tokenCacheEntry)
			s.tokenCacheMu.Unlock()
		}, func() {
			if err := s.keyMgr.SyncFromVault(); err != nil {
				log.Printf("[SSE] 키 동기화 실패: %v", err)
			}
		}, func() {
			// usage_reset: immediately clear stale local counters, then re-sync from vault
			s.keyMgr.ResetDailyCounters()
			if err := s.keyMgr.SyncFromVault(); err != nil {
				log.Printf("[SSE] usage_reset 후 키 동기화 실패: %v", err)
			}
		}, func(services []string) {
			s.mu.Lock()
			s.allowedServices = services
			s.mu.Unlock()
			log.Printf("[SSE] 프록시 서비스 갱신: %v", services)
			// re-sync URLs in background (SSE only carries IDs, not local_url)
			go func() {
				if err := s.syncAllowedServices(); err != nil {
					log.Printf("[SSE] 서비스 URL 재동기화 실패: %v", err)
				}
			}()
		})
		s.sse.Start()

		// initial load of client config and keys from vault
		go func() {
			time.Sleep(2 * time.Second)
			s.syncFromVault()
		}()

		// periodic key sync (every 5 minutes)
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				s.syncFromVault()
			}
		}()

		// start heartbeat
		s.startHeartbeat()
	}

	// initialize model registry (async)
	go func() {
		ollamaURL := s.ollamaURL()
		s.registry.Refresh(cfg.Proxy.Services, models.ServiceURLs{"ollama": ollamaURL}, "")
	}()

	return s
}

// lookupTokenConfig: resolve a Bearer token to {service, model} via vault's /api/token/config.
// Results are cached for 30 seconds to avoid per-request vault calls.
// Returns nil if vault URL is not configured or the token is not found.
func (s *Server) lookupTokenConfig(token string) *tokenCacheEntry {
	if s.cfg.Proxy.VaultURL == "" || token == "" {
		return nil
	}
	// skip our own proxy token (it is already applied via s.service/s.model)
	if token == s.cfg.Proxy.VaultToken {
		return nil
	}

	s.tokenCacheMu.RLock()
	if e, ok := s.tokenCache[token]; ok && time.Now().Before(e.expiresAt) {
		s.tokenCacheMu.RUnlock()
		return e
	}
	s.tokenCacheMu.RUnlock()

	// fetch from vault
	req, err := http.NewRequest("GET", s.cfg.Proxy.VaultURL+"/api/token/config", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var result struct {
		DefaultService string `json:"default_service"`
		DefaultModel   string `json:"default_model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	entry := &tokenCacheEntry{
		service:   result.DefaultService,
		model:     result.DefaultModel,
		expiresAt: time.Now().Add(30 * time.Second),
	}
	s.tokenCacheMu.Lock()
	// evict expired entries when cache grows large
	if len(s.tokenCache) > 500 {
		now := time.Now()
		for k, v := range s.tokenCache {
			if now.After(v.expiresAt) {
				delete(s.tokenCache, k)
			}
		}
	}
	s.tokenCache[token] = entry
	s.tokenCacheMu.Unlock()
	return entry
}

// syncFromVault: sync client config and keys from vault
func (s *Server) syncFromVault() {
	// fetch client config
	url := fmt.Sprintf("%s/api/clients", s.cfg.Proxy.VaultURL)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[sync] 금고 연결 실패: %v", err)
		return
	}
	defer resp.Body.Close()

	var clients []struct {
		ID             string `json:"id"`
		DefaultService string `json:"default_service"`
		DefaultModel   string `json:"default_model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return
	}
	for _, c := range clients {
		if c.ID == s.cfg.Proxy.ClientID {
			s.mu.Lock()
			oldSvc, oldMdl := s.service, s.model
			if c.DefaultService != "" {
				s.service = c.DefaultService
			}
			if c.DefaultModel != "" {
				s.model = c.DefaultModel
			}
			newSvc, newMdl := s.service, s.model
			s.mu.Unlock()
			log.Printf("[sync] 설정 로드: %s/%s", c.DefaultService, c.DefaultModel)
			if newSvc != oldSvc || newMdl != oldMdl {
				go updateOpenClawJSON(newSvc, newMdl)
			}
			break
		}
	}

	// sync keys
	if err := s.keyMgr.SyncFromVault(); err != nil {
		log.Printf("[sync] 키 동기화 실패: %v", err)
	}

	// sync proxy-enabled services
	if err := s.syncAllowedServices(); err != nil {
		log.Printf("[sync] 서비스 목록 동기화 실패: %v", err)
	}
}

// syncAllowedServices: fetch proxy-enabled service list (with local URLs) from vault
func (s *Server) syncAllowedServices() error {
	url := s.cfg.Proxy.VaultURL + "/api/services"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if s.cfg.Proxy.VaultToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("서비스 목록 조회 실패: HTTP %d", resp.StatusCode)
	}
	var svcs []struct {
		ID       string `json:"id"`
		LocalURL string `json:"local_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&svcs); err != nil {
		return err
	}
	ids := make([]string, 0, len(svcs))
	urls := make(map[string]string, len(svcs))
	for _, sv := range svcs {
		ids = append(ids, sv.ID)
		if sv.LocalURL != "" {
			urls[sv.ID] = sv.LocalURL
		}
	}
	s.mu.Lock()
	s.allowedServices = ids
	s.serviceURLs = urls
	s.mu.Unlock()
	log.Printf("[sync] 프록시 서비스 목록: %v (urls: %v)", ids, urls)
	return nil
}

// ollamaURL: return Ollama URL — env > vault service config > localhost default
func (s *Server) ollamaURL() string {
	if v := os.Getenv("OLLAMA_URL"); v != "" {
		return v
	}
	if v := os.Getenv("WV_OLLAMA_URL"); v != "" {
		return v
	}
	s.mu.RLock()
	u := s.serviceURLs["ollama"]
	s.mu.RUnlock()
	if u != "" {
		return u
	}
	return "http://localhost:11434"
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/api/models", s.handleModels)
	mux.HandleFunc("/api/config/model", s.handleConfigModel)
	mux.HandleFunc("/api/config/think-mode", s.handleThinkMode)
	mux.HandleFunc("/reload", s.handleReload)

	// Gemini API handler
	mux.HandleFunc("/google/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "streamGenerateContent") {
			s.handleGeminiStream(w, r)
		} else {
			s.handleGemini(w, r)
		}
	})

	// OpenAI compatible
	mux.HandleFunc("/v1/chat/completions", s.handleOpenAI)

	// Anthropic API (Claude Code, etc.)
	mux.HandleFunc("/v1/messages", s.handleAnthropic)

	// OpenAI-compatible model list (Cursor, VS Code, LM Studio, etc.)
	mux.HandleFunc("/v1/models", s.handleOpenAIModels)

	// Agent config writer (local proxy only — writes config files for cline/claude-code/openclaw/nanoclaw)
	mux.HandleFunc("/agent/apply", s.handleAgentApply)

	return middleware.Chain(mux,
		middleware.Recovery,
		middleware.CORS,
		middleware.Logger,
	)
}

// ─── Health / Status ──────────────────────────────────────────────────────────

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{
		"status":  "ok",
		"version": Version,
		"client":  s.cfg.Proxy.ClientID,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	svc := s.service
	mdl := s.model
	s.mu.RUnlock()

	sseConn := s.sse != nil && s.sse.IsConnected()

	jsonOK(w, map[string]interface{}{
		"status":   "ok",
		"version":  Version,
		"client":   s.cfg.Proxy.ClientID,
		"service":  svc,
		"model":    mdl,
		"sse":      sseConn,
		"filter":   s.cfg.Proxy.ToolFilter,
		"services": s.cfg.Proxy.Services,
		"mode":     s.cfg.Mode,
	})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	svc := r.URL.Query().Get("service")
	all := s.registry.All(svc)
	jsonOK(w, map[string]interface{}{
		"models": all,
		"count":  len(all),
	})
}

func (s *Server) handleConfigModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		jsonError(w, "PUT required", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Service string `json:"service"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	oldSvc, oldMdl := s.service, s.model
	if body.Service != "" {
		s.service = body.Service
	}
	if body.Model != "" {
		s.model = body.Model
	}
	newSvc, newMdl := s.service, s.model
	s.mu.Unlock()
	log.Printf("[config] model changed: %s/%s", newSvc, newMdl)
	// write-through to vault (async, best-effort)
	if newSvc != oldSvc || newMdl != oldMdl {
		go s.pushConfigToVault(newSvc, newMdl)
		s.hooksMgr.Fire(hooks.EventModelChanged, map[string]string{
			"service": newSvc,
			"model":   newMdl,
		})
		go updateOpenClawJSON(newSvc, newMdl)
	}
	jsonOK(w, map[string]string{"status": "ok", "service": newSvc, "model": newMdl})
}

// pushConfigToVault: write-through proxy model change to vault (bidirectional sync)
func (s *Server) pushConfigToVault(service, model string) {
	if s.cfg.Proxy.VaultURL == "" || s.cfg.Proxy.VaultToken == "" {
		return
	}
	payload, _ := json.Marshal(map[string]string{"service": service, "model": model})
	req, err := http.NewRequest(http.MethodPut, s.cfg.Proxy.VaultURL+"/api/config", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[config] vault 동기화 실패: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[config] vault 동기화 오류: %d", resp.StatusCode)
	}
}

func (s *Server) handleThinkMode(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	go s.syncFromVault()
	jsonOK(w, map[string]string{"status": "reloading"})
}

// ─── Gemini API Handler (non-streaming) ──────────────────────────────────────

func (s *Server) handleGemini(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "body read error", http.StatusBadRequest)
		return
	}

	var req GeminiRequest
	if err := json.Unmarshal(body, &req); err != nil {
		jsonError(w, "invalid gemini request", http.StatusBadRequest)
		return
	}

	stripped := s.filter.FilterGemini(&req)
	if stripped > 0 {
		log.Printf("[Security] blocked %d tools from request (client=%s)", stripped, s.cfg.Proxy.ClientID)
	}

	s.mu.RLock()
	svc := s.service
	mdl := s.model
	s.mu.RUnlock()

	if urlModel := extractModelFromPath(r.URL.Path); urlModel != "" {
		if strings.HasPrefix(urlModel, "gemini-") {
			svc = "google"
			mdl = urlModel
		}
	}

	w.Header().Set("Content-Type", "application/json")

	resp, err := s.dispatch(svc, mdl, &req)
	if err != nil {
		log.Printf("[proxy] 오류: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(GeminiResponse{
			Error: &GeminiError{Code: 502, Message: err.Error()},
		})
		return
	}
	json.NewEncoder(w).Encode(resp)
}

// ─── OpenAI API Handler ───────────────────────────────────────────────────────

func (s *Server) handleOpenAI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var oaiReq OpenAIRequest
	if err := json.NewDecoder(r.Body).Decode(&oaiReq); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	s.filter.FilterOpenAI(&oaiReq)

	s.mu.RLock()
	svc := s.service
	mdl := s.model
	s.mu.RUnlock()
	if mdl == "" {
		mdl = oaiReq.Model
	}

	// Token-based model override: if the request carries a different client token,
	// look up that client's dashboard-configured model and override the request model.
	// This allows third-party clients (Cline, Cursor, etc.) to be controlled from
	// the wall-vault dashboard without changing their local settings.
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken := strings.TrimPrefix(authHeader, "Bearer ")
		if entry := s.lookupTokenConfig(reqToken); entry != nil {
			if entry.service != "" {
				svc = entry.service
			}
			if entry.model != "" {
				mdl = entry.model
			}
		}
	}

	// OpenClaw sends models as "provider/model-id" (e.g. "wall-vault/gemini-2.5-flash",
	// "anthropic/claude-opus-4-6"). Parse and route accordingly.
	svc, mdl = parseProviderModel(svc, mdl)

	geminiReq := OpenAIToGemini(&oaiReq)
	geminiResp, err := s.dispatch(svc, mdl, geminiReq)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Convert Gemini functionCall parts to OAI tool_calls (for native Gemini backend).
	for i := range geminiResp.Candidates {
		if geminiResp.Candidates[i].RawToolCalls != nil {
			continue // already set by OpenAIRespToGemini (OpenRouter path)
		}
		var toolCalls []map[string]interface{}
		for _, part := range geminiResp.Candidates[i].Content.Parts {
			if part.FunctionCall == nil {
				continue
			}
			fc, ok := part.FunctionCall.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := fc["name"].(string)
			argsJSON, _ := json.Marshal(fc["args"])
			toolCalls = append(toolCalls, map[string]interface{}{
				"id":   fmt.Sprintf("call_%d", len(toolCalls)),
				"type": "function",
				"function": map[string]interface{}{
					"name":      name,
					"arguments": string(argsJSON),
				},
			})
		}
		if len(toolCalls) > 0 {
			if b, err := json.Marshal(toolCalls); err == nil {
				geminiResp.Candidates[i].RawToolCalls = b
			}
		}
	}

	type candidate struct {
		text         string
		finishReason string
		index        int
	}
	var cands []candidate
	for i, c := range geminiResp.Candidates {
		reason := strings.ToLower(c.FinishReason)
		// Map to OAI finish_reason: tool_calls when the candidate carries tool calls.
		if len(geminiResp.Candidates[i].RawToolCalls) > 0 {
			reason = "tool_calls"
		}
		cands = append(cands, candidate{
			text:         stripControlTokens(extractText(c.Content.Parts)),
			finishReason: reason,
			index:        c.Index,
		})
	}

	if oaiReq.Stream {
		// SSE streaming response (OpenAI chat.completion.chunk format)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, _ := w.(http.Flusher)

		writeSSE := func(data []byte) {
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
		}

		for _, c := range cands {
			// role delta
			roleChunk := map[string]interface{}{
				"id":      "chatcmpl-proxy",
				"object":  "chat.completion.chunk",
				"model":   mdl,
				"choices": []map[string]interface{}{{"index": c.index, "delta": map[string]string{"role": "assistant"}, "finish_reason": nil}},
			}
			if b, err := json.Marshal(roleChunk); err == nil {
				writeSSE(b)
			}
			// content or tool_calls delta
			var rawToolCalls json.RawMessage
			if c.index < len(geminiResp.Candidates) {
				rawToolCalls = geminiResp.Candidates[c.index].RawToolCalls
			}
			if len(rawToolCalls) > 0 {
				// Emit tool_calls delta so the client can execute them.
				// Each tool call must carry an index field per OAI streaming spec.
				var tcList []map[string]interface{}
				if json.Unmarshal(rawToolCalls, &tcList) == nil {
					for i, tc := range tcList {
						if _, hasIdx := tc["index"]; !hasIdx {
							tc["index"] = i
							tcList[i] = tc
						}
					}
					toolChunk := map[string]interface{}{
						"id":     "chatcmpl-proxy",
						"object": "chat.completion.chunk",
						"model":  mdl,
						"choices": []map[string]interface{}{{
							"index":         c.index,
							"delta":         map[string]interface{}{"role": "assistant", "tool_calls": tcList},
							"finish_reason": nil,
						}},
					}
					if b, err := json.Marshal(toolChunk); err == nil {
						writeSSE(b)
					}
				}
			} else {
				contentChunk := map[string]interface{}{
					"id":      "chatcmpl-proxy",
					"object":  "chat.completion.chunk",
					"model":   mdl,
					"choices": []map[string]interface{}{{"index": c.index, "delta": map[string]string{"content": c.text}, "finish_reason": nil}},
				}
				if b, err := json.Marshal(contentChunk); err == nil {
					writeSSE(b)
				}
			}
			// finish delta
			finishChunk := map[string]interface{}{
				"id":      "chatcmpl-proxy",
				"object":  "chat.completion.chunk",
				"model":   mdl,
				"choices": []map[string]interface{}{{"index": c.index, "delta": map[string]interface{}{}, "finish_reason": c.finishReason}},
			}
			if b, err := json.Marshal(finishChunk); err == nil {
				writeSSE(b)
			}
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		return
	}

	oaiResp := &OpenAIResponse{}
	for _, c := range cands {
		msg := OpenAIMessage{Role: "assistant", Content: c.text}
		// Include tool_calls if the backend returned them.
		if c.index < len(geminiResp.Candidates) {
			msg.ToolCalls = geminiResp.Candidates[c.index].RawToolCalls
		}
		oaiResp.Choices = append(oaiResp.Choices, OpenAIChoice{
			Message:      msg,
			FinishReason: c.finishReason,
			Index:        c.index,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oaiResp)
}

// ─── Anthropic API Handler (/v1/messages) ─────────────────────────────────────

func (s *Server) handleAnthropic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req AnthropicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	svc := s.service
	mdl := s.model
	allowedServices := s.allowedServices
	s.mu.RUnlock()
	if mdl == "" {
		mdl = req.Model
	}

	// Token-based model override: same logic as handleOpenAI.
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken := strings.TrimPrefix(authHeader, "Bearer ")
		if entry := s.lookupTokenConfig(reqToken); entry != nil {
			if entry.service != "" {
				svc = entry.service
			}
			if entry.model != "" {
				mdl = entry.model
			}
		}
	}

	// Parse provider/model form (e.g. "anthropic/claude-opus-4-6")
	svc, mdl = parseProviderModel(svc, mdl)

	// Native Anthropic passthrough: when Anthropic is available, forward the
	// original request body directly — no GeminiRequest round-trip conversion.
	// This preserves tool calls, tool_results, and multi-block content.
	anthropicAllowed := len(allowedServices) == 0
	for _, sv := range allowedServices {
		if sv == "anthropic" {
			anthropicAllowed = true
			break
		}
	}
	if anthropicAllowed {
		if body, _, err := s.callAnthropicPassthrough(&req, mdl); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(body) //nolint:errcheck
			return
		} else {
			log.Printf("[anthropic] passthrough failed → fallback to dispatch: %v", err)
		}
	}

	// Fallback: convert to GeminiRequest and dispatch via Google/OpenRouter
	geminiReq := AnthropicToGemini(&req)
	geminiResp, err := s.dispatch(svc, mdl, geminiReq)
	if err != nil {
		log.Printf("[anthropic] dispatch error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type":  "error",
			"error": map[string]string{"type": "api_error", "message": err.Error()},
		})
		return
	}

	resp := GeminiRespToAnthropic(mdl, geminiResp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── OpenAI-compatible model list (/v1/models) ────────────────────────────────

func (s *Server) handleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	all := s.registry.All("")
	type oaiModel struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	}
	var data []oaiModel
	for _, m := range all {
		data = append(data, oaiModel{
			ID:      m.ID,
			Object:  "model",
			OwnedBy: m.Service,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   data,
	})
}

// ─── Request Dispatch ─────────────────────────────────────────────────────────

func (s *Server) dispatch(service, model string, req *GeminiRequest) (*GeminiResponse, error) {
	s.mu.RLock()
	allowedServices := s.allowedServices
	s.mu.RUnlock()

	var lastErr error

	// Build try order from the vault UI proxy-enabled service list (allowedServices),
	// which matches the order shown in the Services section of the dashboard.
	// The currently configured service is moved to the front so it is always tried first.
	// Falls back to s.cfg.Proxy.Services when the vault list is not yet available.
	var tryOrder []string
	if len(allowedServices) > 0 {
		tryOrder = append(tryOrder, service)
		for _, svc := range allowedServices {
			if svc != service {
				tryOrder = append(tryOrder, svc)
			}
		}
	} else {
		tryOrder = append(tryOrder, service)
		for _, svc := range s.cfg.Proxy.Services {
			if svc != service {
				tryOrder = append(tryOrder, svc)
			}
		}
	}

	for _, svc := range tryOrder {
		var resp *GeminiResponse
		var err error
		switch svc {
		case "google":
			resp, err = s.callGoogle(model, req)
		case "openrouter":
			resp, err = s.callOpenRouter(model, req)
		case "ollama":
			resp, err = s.callOllama(model, req)
		case "openai":
			resp, err = s.callOpenAI(model, req)
		case "anthropic":
			resp, err = s.callAnthropic(model, req)
		default:
			continue
		}
		if err == nil {
			// When a fallback service is used, update s.service/s.model so the
			// next heartbeat, vault UI, and openclaw TUI all reflect the actual model.
			if svc != service {
				go s.onFallback(svc, s.resolveActualModel(svc, model))
			}
			return resp, nil
		}
		log.Printf("[proxy] %s failed → fallback: %v", svc, err)
		lastErr = err
	}
	if lastErr != nil {
		s.hooksMgr.Fire(hooks.EventServiceDown, map[string]string{
			"error": lastErr.Error(),
		})
	}
	return nil, fmt.Errorf("모든 서비스 실패: %v", lastErr)
}

// ─── Google Gemini ────────────────────────────────────────────────────────────

func (s *Server) callGoogle(model string, req *GeminiRequest) (*GeminiResponse, error) {
	// Strip "google/" prefix if present (passed through from parseProviderModel for fallback compatibility)
	model = strings.TrimPrefix(model, "google/")
	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		key, plainKey, err := s.getKey("google")
		if err != nil {
			s.hooksMgr.Fire(hooks.EventKeyExhausted, map[string]string{"service": "google"})
			lastErr = err
			break
		}

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, plainKey)
		data, _ := json.Marshal(req)
		resp, err := s.doRequest("POST", url, data, nil)
		if err != nil {
			s.keyMgr.RecordError(key, 0)
			lastErr = err
			break
		}

		if resp.StatusCode == http.StatusNotFound {
			// Model not found — key is fine, skip service entirely
			resp.Body.Close()
			return nil, fmt.Errorf("Google: 모델 없음 (%s)", model)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPaymentRequired {
			// Key-level error — cooldown this key and retry with another
			resp.Body.Close()
			s.keyMgr.RecordError(key, resp.StatusCode)
			lastErr = fmt.Errorf("Google API 오류: HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			s.keyMgr.RecordError(key, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("[google] HTTP %d: %s", resp.StatusCode, string(body))
			return nil, fmt.Errorf("Google API 오류: HTTP %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var geminiResp GeminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return nil, fmt.Errorf("Google 응답 파싱 오류: %w", err)
		}
		if geminiResp.Error != nil {
			return nil, fmt.Errorf("Google: %s", geminiResp.Error.Message)
		}
		tokens := 1 // minimum 1 per request for usage tracking
		if geminiResp.UsageMetadata != nil && geminiResp.UsageMetadata.TotalTokenCount > 0 {
			tokens = geminiResp.UsageMetadata.TotalTokenCount
		}
		s.keyMgr.RecordSuccess(key, tokens)
		return &geminiResp, nil
	}
	return nil, lastErr
}

// ─── OpenRouter ───────────────────────────────────────────────────────────────

func (s *Server) callOpenRouter(model string, req *GeminiRequest) (*GeminiResponse, error) {
	resp, err := s.callOpenRouterModel(model, req)
	if err == nil {
		return resp, nil
	}
	// If paid model failed with payment errors, retry with free-tier variant
	if !strings.HasSuffix(model, ":free") {
		freeModel := model + ":free"
		log.Printf("[proxy] openrouter paid failed, retrying free tier: %s", freeModel)
		if resp, err2 := s.callOpenRouterModel(freeModel, req); err2 == nil {
			return resp, nil
		}
	}
	return nil, err
}

func (s *Server) callOpenRouterModel(model string, req *GeminiRequest) (*GeminiResponse, error) {
	const maxAttempts = 3
	var lastErr error
	oaiReq := GeminiToOpenAI(model, req)
	data, _ := json.Marshal(oaiReq)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		key, plainKey, err := s.getKey("openrouter")
		if err != nil {
			s.hooksMgr.Fire(hooks.EventKeyExhausted, map[string]string{"service": "openrouter"})
			lastErr = err
			break
		}

		headers := map[string]string{
			"Authorization": "Bearer " + plainKey,
			"HTTP-Referer":  "https://github.com/sookmook/wall-vault",
			"X-Title":       "wall-vault",
		}
		resp, err := s.doRequest("POST", "https://openrouter.ai/api/v1/chat/completions", data, headers)
		if err != nil {
			s.keyMgr.RecordError(key, 0)
			lastErr = err
			break
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("OpenRouter: 모델 없음 (%s)", model)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPaymentRequired {
			resp.Body.Close()
			s.keyMgr.RecordError(key, resp.StatusCode)
			lastErr = fmt.Errorf("OpenRouter 오류: HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			s.keyMgr.RecordError(key, resp.StatusCode)
			resp.Body.Close()
			return nil, fmt.Errorf("OpenRouter 오류: HTTP %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var oaiResp OpenAIResponse
		if err := json.Unmarshal(body, &oaiResp); err != nil {
			return nil, fmt.Errorf("OpenRouter 응답 파싱 오류: %w", err)
		}
		if oaiResp.Error != nil {
			return nil, fmt.Errorf("OpenRouter: %s", oaiResp.Error.Message)
		}
		tokens := 1 // minimum 1 per request for usage tracking
		if oaiResp.Usage != nil && oaiResp.Usage.TotalTokens > 0 {
			tokens = oaiResp.Usage.TotalTokens
		}
		s.keyMgr.RecordSuccess(key, tokens)
		return OpenAIRespToGemini(&oaiResp), nil
	}
	return nil, lastErr
}

// ─── OpenAI Direct Call ───────────────────────────────────────────────────────

func (s *Server) callOpenAI(model string, req *GeminiRequest) (*GeminiResponse, error) {
	const maxAttempts = 3
	var lastErr error
	oaiReq := GeminiToOpenAI(model, req)
	data, _ := json.Marshal(oaiReq)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		key, plainKey, err := s.getKey("openai")
		if err != nil {
			s.hooksMgr.Fire(hooks.EventKeyExhausted, map[string]string{"service": "openai"})
			lastErr = err
			break
		}

		headers := map[string]string{
			"Authorization": "Bearer " + plainKey,
		}
		resp, err := s.doRequest("POST", "https://api.openai.com/v1/chat/completions", data, headers)
		if err != nil {
			s.keyMgr.RecordError(key, 0)
			lastErr = err
			break
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("OpenAI: 모델 없음 (%s)", model)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusPaymentRequired {
			resp.Body.Close()
			s.keyMgr.RecordError(key, resp.StatusCode)
			lastErr = fmt.Errorf("OpenAI 오류: HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			s.keyMgr.RecordError(key, resp.StatusCode)
			resp.Body.Close()
			return nil, fmt.Errorf("OpenAI 오류: HTTP %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var oaiResp OpenAIResponse
		if err := json.Unmarshal(body, &oaiResp); err != nil {
			return nil, fmt.Errorf("OpenAI 응답 파싱 오류: %w", err)
		}
		if oaiResp.Error != nil {
			return nil, fmt.Errorf("OpenAI: %s", oaiResp.Error.Message)
		}
		tokens := 1 // minimum 1 per request for usage tracking
		if oaiResp.Usage != nil && oaiResp.Usage.TotalTokens > 0 {
			tokens = oaiResp.Usage.TotalTokens
		}
		s.keyMgr.RecordSuccess(key, tokens)
		return OpenAIRespToGemini(&oaiResp), nil
	}
	return nil, lastErr
}

// ─── Ollama (single concurrent request via mutex) ────────────────────────────

func (s *Server) callOllama(_ string, req *GeminiRequest) (*GeminiResponse, error) {
	// always use the configured Ollama model, ignoring the upstream service's model name
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = os.Getenv("WV_OLLAMA_MODEL")
	}
	if model == "" {
		model = "qwen3.5:35b"
	}
	ollamaURL := s.ollamaURL()

	s.ollamaMu.Lock()
	defer s.ollamaMu.Unlock()

	// Use Ollama's OpenAI-compatible /v1/chat/completions endpoint.
	// The native /api/chat expects tool_calls.function.arguments as a JSON object,
	// while OpenAI format (and our internal representation) uses a JSON string —
	// sending the OpenAI format to /api/chat causes HTTP 400.
	// /v1/chat/completions accepts the standard OpenAI format including arguments-as-string.
	oaiReq := GeminiToOpenAI(model, req)
	oaiReq.Stream = false
	data, _ := json.Marshal(oaiReq)

	// Ollama is a local service: inference can take several minutes for large models.
	// Use a dedicated client with no hard timeout so generation is never cut off.
	// Connection errors (server not running) still surface immediately.
	httpReq, err := http.NewRequest("POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Ollama 요청 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	ollamaClient := &http.Client{Timeout: 0} // no timeout — inference duration is unbounded
	resp, err := ollamaClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Ollama 연결 실패 (%s): %w", ollamaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama 오류: HTTP %d: %s", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	var oaiResp OpenAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("Ollama 응답 파싱 오류: %w", err)
	}
	if oaiResp.Error != nil {
		return nil, fmt.Errorf("Ollama: %s", oaiResp.Error.Message)
	}
	return OpenAIRespToGemini(&oaiResp), nil
}

// ─── Common HTTP Request ──────────────────────────────────────────────────────

func (s *Server) doRequest(method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: s.cfg.Proxy.Timeout}
	return client.Do(req)
}

func (s *Server) getKey(service string) (*localKey, string, error) {
	k, err := s.keyMgr.Get(service)
	if err != nil {
		return nil, "", err
	}
	return k, k.plaintext, nil
}

// ─── Util ─────────────────────────────────────────────────────────────────────

// parseProviderModel handles OpenClaw's "provider/model-id" format.
//
// Routing rules:
//   - google/*, ollama/*  → native handlers
//   - openai/*            → direct OpenAI (callOpenAI)
//   - anthropic/*         → OpenRouter with "anthropic/model" path (Anthropic API format differs)
//   - openrouter/*        → OpenRouter, bare model path kept
//   - opencode*, moonshot, kimi-coding, groq, mistral, minimax, … → OpenRouter
//   - wall-vault/*        → auto-detect from model ID prefix
//   - anything else        → OpenRouter (OpenRouter-style "org/model" paths)
//
// Ollama :cloud suffix (e.g. "kimi-k2.5:cloud") is stripped and routed to OpenRouter.
func parseProviderModel(svc, mdl string) (string, string) {
	// Strip Ollama :cloud suffix — route cloud variants via OpenRouter
	if strings.HasSuffix(mdl, ":cloud") {
		bare := strings.TrimSuffix(mdl, ":cloud")
		return "openrouter", bare
	}

	if !strings.Contains(mdl, "/") {
		return svc, mdl
	}
	parts := strings.SplitN(mdl, "/", 2)
	prefix, bare := parts[0], parts[1]

	// When the caller explicitly chose OpenRouter as the service, honour that
	// choice for models with provider prefixes (e.g. google/gemini-*) instead
	// of re-routing them to the native Google/OpenAI handler.
	if svc == "openrouter" && prefix != "openrouter" && prefix != "ollama" {
		return "openrouter", mdl
	}

	switch prefix {
	// ── Native handlers ──────────────────────────────────────────────────────
	case "google":
		// Keep full "google/X" form so OpenRouter fallback receives a valid model ID.
		// callGoogle strips the "google/" prefix itself before building the API URL.
		return "google", mdl
	case "ollama":
		return "ollama", bare

	// ── OpenAI direct ────────────────────────────────────────────────────────
	case "openai":
		return "openai", bare

	// ── Anthropic via OpenRouter (API format differs) ─────────────────────────
	case "anthropic":
		return "openrouter", "anthropic/" + bare

	// ── OpenRouter pass-through (bare already has provider/model) ─────────────
	case "openrouter":
		return "openrouter", bare

	// ── OpenClaw 3.11 providers → OpenRouter ─────────────────────────────────
	case "opencode", "opencode-go", "opencode-zen":
		return "openrouter", bare
	case "moonshot", "kimi-coding":
		return "openrouter", mdl // keep full "moonshot/kimi-k2.5" for OpenRouter
	case "groq", "mistral", "cohere", "perplexity",
		"minimax", "minimax-text", "together", "together-ai",
		"huggingface", "novita", "nvidia", "venice",
		"meta-llama", "qwen", "deepseek", "01-ai":
		return "openrouter", mdl // keep full org/model path

	// ── custom/ prefix — strip and re-parse the remainder ───────────────────
	// OpenClaw model picker sends "custom/google/gemini-..." or "custom/openrouter/..."
	// when the user selects a custom entry. Strip the leading "custom/" and re-parse
	// so the actual provider is detected correctly.
	case "custom":
		return parseProviderModel(svc, bare) // bare = "google/gemini-..." etc.

	// ── wall-vault prefix — auto-detect from model ID ────────────────────────
	case "wall-vault":
		bare = strings.TrimSuffix(bare, ":cloud")
		switch {
		case strings.HasPrefix(bare, "gemini-"):
			// Keep "google/bare" so OpenRouter fallback receives a valid model ID
			return "google", "google/" + bare
		case strings.HasPrefix(bare, "claude-"):
			// Route Anthropic models through OpenRouter
			return "openrouter", "anthropic/" + bare
		case strings.HasPrefix(bare, "gpt-"),
			bare == "o1", bare == "o1-mini",
			strings.HasPrefix(bare, "o3"), strings.HasPrefix(bare, "o4"):
			return "openai", bare
		default:
			// kimi-*, glm-*, deepseek-*, qwen*, hunter-alpha, healer-alpha, etc.
			return "openrouter", bare
		}

	// ── Default: OpenRouter-style "org/model" ────────────────────────────────
	default:
		return "openrouter", mdl
	}
}

// stripControlTokens removes model-internal delimiter tokens that should not
// be exposed to users. Handles DeepSeek im_start/im_end markers and GLM-5
// special tokens (OpenClaw 3.11 security requirement).
func stripControlTokens(s string) string {
	// DeepSeek / ChatML delimiters
	for _, tok := range []string{
		"<|im_start|>", "<|im_end|>",
		"<|user|>", "<|assistant|>", "<|system|>",
		"<|endoftext|>", "<|end|>",
	} {
		s = strings.ReplaceAll(s, tok, "")
	}
	// GLM-5 special tokens
	for _, tok := range []string{"[gMASK]", "[sop]", "[EOP]", "[eop]", "[MASK]"} {
		s = strings.ReplaceAll(s, tok, "")
	}
	return strings.TrimSpace(s)
}

// onFallback: called when dispatch succeeds on a service other than the requested one.
// Updates s.service/s.model so the next heartbeat, vault UI, and openclaw TUI reflect reality.
func (s *Server) onFallback(actualSvc, actualMdl string) {
	s.mu.Lock()
	s.service = actualSvc
	s.model = actualMdl
	s.mu.Unlock()
	log.Printf("[proxy] fallback active: %s/%s", actualSvc, actualMdl)
	s.pushConfigToVault(actualSvc, actualMdl)
	go updateOpenClawJSON(actualSvc, actualMdl)
	s.hooksMgr.Fire(hooks.EventModelChanged, map[string]string{
		"service": actualSvc,
		"model":   actualMdl,
	})
}

// resolveActualModel returns the model name that will actually be used for the given service.
// Ollama ignores the upstream model and uses a local env-configured model instead.
func (s *Server) resolveActualModel(svc, model string) string {
	if svc == "ollama" {
		if m := os.Getenv("OLLAMA_MODEL"); m != "" {
			return m
		}
		if m := os.Getenv("WV_OLLAMA_MODEL"); m != "" {
			return m
		}
		return "qwen3.5:35b"
	}
	return model
}

func extractModelFromPath(path string) string {
	prefix := "/google/v1beta/models/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)
	model := strings.SplitN(rest, ":", 2)[0]
	return model
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
