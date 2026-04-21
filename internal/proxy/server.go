package proxy

import (
	"bytes"
	"context"
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

// request body size limits
const (
	maxAIBodySize     = 50 << 20 // 50 MB for AI request endpoints (large prompts/context)
	maxConfigBodySize = 1 << 20  // 1 MB for config endpoints
)

// tokenCacheEntry: cached result of a token→model lookup from the vault
type tokenCacheEntry struct {
	clientID  string // vault client ID for this token
	service   string
	model     string
	expiresAt time.Time
}

// clientAct: last-seen activity record for a non-proxy client served by this proxy
type clientAct struct {
	service  string
	model    string
	lastSeen time.Time
	applied  bool // true if config was applied via /agent/apply — never expires from heartbeat
}

// Server: proxy HTTP server
type Server struct {
	cfg             *config.Config
	mu              sync.RWMutex
	service         string            // user-configured preferred service (from vault dashboard)
	model           string            // user-configured preferred model (from vault dashboard)
	claudeCodeClientID string            // vault client ID for the local claude-code agent (from syncFromVault)
	ownAgentType       string            // this proxy's own agent_type (from syncFromVault)
	allowedServices []string          // proxy-enabled services from vault (empty = no restriction)
	serviceURLs      map[string]string // service ID → local URL from vault config
	serviceDefaults  map[string]string // service ID → default_model from vault config
	serviceReasoning map[string]bool   // service ID → reasoning_mode toggle from vault config
	keyMgr          *KeyManager
	filter          *ToolFilter
	sse             *SSEClient
	registry        *models.Registry
	hooksMgr        *hooks.Manager
	// ollamaSem caps concurrent Ollama requests. A buffered channel doubles as
	// a context-aware semaphore: callOllama's acquire selects on ctx.Done(),
	// so a caller whose HTTP request is cancelled won't keep holding a slot
	// behind a slow upstream — which the previous plain Mutex couldn't offer.
	ollamaSem       chan struct{}

	// stopCh is closed by Stop() to signal background goroutines (periodic
	// vault sync, token-cache eviction, initial-load delay) to exit instead
	// of leaking past server shutdown. systemd `Restart=always` or a
	// launchctl unload then runs with a clean slate.
	stopCh          chan struct{}
	tokenCacheMu    sync.RWMutex
	tokenCache      map[string]*tokenCacheEntry // Bearer token → client model config
	clientActMu     sync.Mutex
	clientActs      map[string]*clientAct // clientID → last-seen activity (for heartbeat reporting)
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
		clientActs: make(map[string]*clientAct),
		// Ollama stays serialized (size 1): large local models are typically
		// memory-bound, and running two inferences concurrently tends to be
		// slower than two sequential ones. Bumped via a constant when your
		// local setup can genuinely parallelize.
		ollamaSem: make(chan struct{}, 1),
		stopCh:    make(chan struct{}),
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
		s.sse = NewSSEClient(cfg.Proxy.VaultURL, cfg.Proxy.ClientID, cfg.Proxy.VaultToken, func(svc, mdl string) {
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
		}, func(clientID, agentType, svc, mdl string) {
			// Foreign client model changed in vault — update local agent config if applicable.
			switch agentType {
			case "cline":
				if mdl != "" {
					go updateClineModel(mdl)
				}
			case "claude-code":
				if mdl != "" {
					go updateClaudeCodeModel(mdl)
				}
			case "econoworld":
				// When mdl is empty (vault soft-cleared a stale override on a service
				// switch, or user explicitly picked "(서비스 기본 사용)"), fall back to
				// the new service's default_model so ai_config.json stays consistent
				// with what /status surfaces. Without this, ai_config.json keeps the
				// previous override and drifts from the proxy's actual routing.
				effective := mdl
				if effective == "" && svc != "" {
					s.mu.RLock()
					effective = s.serviceDefaults[svc]
					s.mu.RUnlock()
				}
				if effective != "" {
					go updateEconoWorldModel(effective)
				}
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
		}, func() {
			// SSE reconnect: flush token cache and re-sync model/config from vault
			// to pick up any changes that occurred while the connection was down.
			s.tokenCacheMu.Lock()
			s.tokenCache = make(map[string]*tokenCacheEntry)
			s.tokenCacheMu.Unlock()
			log.Printf("[SSE] reconnect — re-syncing from vault")
			go s.syncFromVault()
		})
		s.sse.Start()

		// initial load of client config and keys from vault (2s delay,
		// cancellable via stopCh so Stop() during startup doesn't leak this
		// goroutine).
		go func() {
			select {
			case <-time.After(2 * time.Second):
				s.syncFromVault()
			case <-s.stopCh:
				return
			}
		}()

		// periodic key sync (every 5 minutes)
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-s.stopCh:
					return
				case <-ticker.C:
					s.syncFromVault()
				}
			}
		}()

		// periodic token cache eviction — trim expired entries every 30s so the
		// map stays bounded even when many short-lived third-party tokens are
		// seen (without this, eviction only ran when the cache crossed a fixed
		// threshold).
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-s.stopCh:
					return
				case <-ticker.C:
					s.evictExpiredTokens()
				}
			}
		}()

		// start heartbeat
		s.startHeartbeat()
	}

	// initialize model registry (async)
	go func() {
		ollamaURL := s.ollamaURL()
		s.registry.Refresh(cfg.Proxy.Services, models.ServiceURLs{"ollama": ollamaURL}, nil)
	}()

	return s
}

// Stop signals all background goroutines to exit. Safe to call multiple times
// because closing an already-closed channel panics — we guard with a sync.Once
// pattern via a defer/recover hack would obscure intent, so callers are
// expected to call Stop at most once (typical for `os.Exit` / systemd stop).
func (s *Server) Stop() {
	select {
	case <-s.stopCh:
		return // already closed
	default:
		close(s.stopCh)
	}
}

// refreshClientAct: update lastSeen for a tracked client after response completion.
// This ensures long-running streaming requests keep the client alive on the dashboard.
func (s *Server) refreshClientAct(clientID string) {
	if clientID == "" {
		return
	}
	s.clientActMu.Lock()
	if act, ok := s.clientActs[clientID]; ok {
		act.lastSeen = time.Now()
	}
	s.clientActMu.Unlock()
}

// lookupTokenConfig: resolve a Bearer token to {service, model} via vault's /api/token/config.
// Results are cached for 5 seconds to avoid per-request vault calls.
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
		// Refresh clientAct lastSeen on cache hits so heartbeat keeps reporting this client
		if e.clientID != "" {
			s.clientActMu.Lock()
			if act, ok := s.clientActs[e.clientID]; ok {
				act.lastSeen = time.Now()
			}
			s.clientActMu.Unlock()
		}
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
		ID             string `json:"id"`
		DefaultService string `json:"default_service"`
		DefaultModel   string `json:"default_model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	entry := &tokenCacheEntry{
		clientID:  result.ID,
		service:   result.DefaultService,
		model:     result.DefaultModel,
		expiresAt: time.Now().Add(5 * time.Second),
	}
	// Record client activity so the heartbeat can report it to vault.
	// Preserve the applied flag if the entry was previously set via /agent/apply,
	// so that the agent keeps its longer grace period on the dashboard.
	if result.ID != "" {
		s.clientActMu.Lock()
		wasApplied := false
		if existing, ok := s.clientActs[result.ID]; ok {
			wasApplied = existing.applied
		}
		s.clientActs[result.ID] = &clientAct{
			service:  result.DefaultService,
			model:    result.DefaultModel,
			lastSeen: time.Now(),
			applied:  wasApplied,
		}
		s.clientActMu.Unlock()
	}
	s.tokenCacheMu.Lock()
	// keep the map bounded: if the periodic eviction can't keep up (e.g. a
	// short burst of unique tokens), prune expired entries inline before
	// inserting the new one.
	if len(s.tokenCache) > tokenCacheMaxSize {
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

// tokenCacheMaxSize bounds the in-memory token→config cache. The periodic
// eviction goroutine keeps it roughly empty under normal load; this cap is an
// inline safety valve for bursty traffic.
const tokenCacheMaxSize = 200

// evictExpiredTokens removes tokenCache entries whose TTL has passed.
// Called both periodically (ticker) and opportunistically from lookupTokenConfig.
func (s *Server) evictExpiredTokens() {
	s.tokenCacheMu.Lock()
	defer s.tokenCacheMu.Unlock()
	now := time.Now()
	for k, v := range s.tokenCache {
		if now.After(v.expiresAt) {
			delete(s.tokenCache, k)
		}
	}
}

// syncFromVault: sync client config and keys from vault
func (s *Server) syncFromVault() {
	// fetch client config (authenticated to receive agent_type in response)
	url := fmt.Sprintf("%s/api/clients", s.cfg.Proxy.VaultURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[sync] 금고 연결 실패: %v", err)
		return
	}
	if s.cfg.Proxy.VaultToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		log.Printf("[sync] 금고 연결 실패: %v", err)
		return
	}
	defer resp.Body.Close()

	var clients []struct {
		ID             string `json:"id"`
		DefaultService string `json:"default_service"`
		DefaultModel   string `json:"default_model"`
		AgentType      string `json:"agent_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return
	}
	ccID := ""
	for _, c := range clients {
		if c.AgentType == "claude-code" {
			ccID = c.ID
		}
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
		}
	}
	// find own agent_type
	ownType := ""
	for _, c := range clients {
		if c.ID == s.cfg.Proxy.ClientID {
			ownType = c.AgentType
			break
		}
	}
	s.mu.Lock()
	s.claudeCodeClientID = ccID
	s.ownAgentType = ownType
	s.mu.Unlock()

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
		ID            string `json:"id"`
		LocalURL      string `json:"local_url"`
		DefaultModel  string `json:"default_model"`
		ReasoningMode bool   `json:"reasoning_mode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&svcs); err != nil {
		return err
	}
	ids := make([]string, 0, len(svcs))
	urls := make(map[string]string, len(svcs))
	defaults := make(map[string]string, len(svcs))
	reasoning := make(map[string]bool, len(svcs))
	for _, sv := range svcs {
		ids = append(ids, sv.ID)
		if sv.LocalURL != "" {
			urls[sv.ID] = sv.LocalURL
		}
		if sv.DefaultModel != "" {
			defaults[sv.ID] = sv.DefaultModel
		}
		if sv.ReasoningMode {
			reasoning[sv.ID] = true
		}
	}
	s.mu.Lock()
	s.allowedServices = ids
	s.serviceURLs = urls
	s.serviceDefaults = defaults
	s.serviceReasoning = reasoning
	s.mu.Unlock()
	log.Printf("[sync] 프록시 서비스 목록: %v (urls: %v, defaults: %v, reasoning: %v)", ids, urls, defaults, reasoning)
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

	// Generous per-IP limit (100 req/s, burst 20) — wall-vault's proxy carries
	// AI traffic for at most a handful of agents, so legitimate callers never
	// approach this, while a misbehaving loop or scanner is bounded.
	rl := middleware.NewRateLimiter(100, 20)
	return middleware.Chain(mux,
		middleware.Recovery,
		middleware.SecurityHeaders,
		rl.Middleware,
		middleware.CORS,
		middleware.Logger,
	)
}

// ─── Health / Status ──────────────────────────────────────────────────────────

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// sseConnected: in distributed mode the proxy is only "ready" once it has
	// successfully subscribed to the vault SSE stream; otherwise key changes /
	// model changes are missed silently. We expose the bit so K8s readiness
	// probes and the doctor command can treat "port open but vault unreachable"
	// as degraded rather than healthy.
	sseConnected := true
	if s.cfg.Proxy.VaultURL != "" {
		sseConnected = s.sse != nil && s.sse.IsConnected()
	}
	readiness := sseConnected
	jsonOK(w, map[string]interface{}{
		"status":        "ok",
		"readiness":     readiness,
		"version":       Version,
		"client":        s.cfg.Proxy.ClientID,
		"sse_connected": sseConnected,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	svc := s.service
	mdl := s.model
	svcs := s.allowedServices
	defaults := s.serviceDefaults
	s.mu.RUnlock()

	clientID := s.cfg.Proxy.ClientID

	// Token-aware: if the caller presents a Bearer token for a known client,
	// return THAT client's config instead of the proxy's own. This lets
	// observability consumers (e.g. an analyzer whose token maps to a
	// different client than the proxy itself) see the routing that will be
	// applied to their own requests. Unauthenticated callers keep the
	// previous behaviour — proxy's own client config.
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		tok := strings.TrimPrefix(auth, "Bearer ")
		if tok != "" && tok != s.cfg.Proxy.VaultToken {
			if entry := s.lookupTokenConfig(tok); entry != nil {
				clientID = entry.clientID
				svc = entry.service
				mdl = entry.model
			}
		}
	}

	// "active model" semantics: if the resolved client has no explicit
	// model_override, surface the vault-synced default_model for the
	// selected service. Consumers need the model dispatch will apply,
	// not an empty string.
	if mdl == "" && defaults != nil {
		mdl = defaults[svc]
	}

	// show vault-synced services if available, else fall back to config
	if len(svcs) == 0 {
		svcs = s.cfg.Proxy.Services
	}

	sseConn := s.sse != nil && s.sse.IsConnected()

	jsonOK(w, map[string]interface{}{
		"status":   "ok",
		"version":  Version,
		"client":   clientID,
		"service":  svc,
		"model":    mdl,
		"sse":      sseConn,
		"filter":   s.cfg.Proxy.ToolFilter,
		"services": svcs,
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
	r.Body = http.MaxBytesReader(w, r.Body, maxConfigBodySize)
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
// Includes client_id query param so the vault can identify this proxy unambiguously,
// even when multiple proxies share the same token.
func (s *Server) pushConfigToVault(service, model string) {
	if s.cfg.Proxy.VaultURL == "" || s.cfg.Proxy.VaultToken == "" {
		return
	}
	payload, _ := json.Marshal(map[string]string{"service": service, "model": model})
	url := s.cfg.Proxy.VaultURL + "/api/config?client_id=" + s.cfg.Proxy.ClientID
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
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
	r.Body = http.MaxBytesReader(w, r.Body, maxAIBodySize)

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
		if strings.HasPrefix(urlModel, "gemini-") || strings.HasPrefix(urlModel, "gemma-") {
			svc = "google"
			mdl = urlModel
		}
	}

	w.Header().Set("Content-Type", "application/json")

	result, err := s.dispatch(r.Context(), svc, mdl, &req)
	if err != nil {
		log.Printf("[proxy] 오류: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(GeminiResponse{
			Error: &GeminiError{Code: 502, Message: err.Error()},
		})
		return
	}
	json.NewEncoder(w).Encode(result.Response)
}

// ─── OpenAI API Handler ───────────────────────────────────────────────────────

func (s *Server) handleOpenAI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAIBodySize)

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
	var resolvedClientID string
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken := strings.TrimPrefix(authHeader, "Bearer ")
		if entry := s.lookupTokenConfig(reqToken); entry != nil {
			resolvedClientID = entry.clientID
			if entry.service != "" {
				svc = entry.service
			}
			if entry.model != "" {
				mdl = entry.model
			}
		}
	}
	// Refresh lastSeen after response completes so long-running requests
	// (streaming AI responses) keep the client visible on the dashboard.
	defer s.refreshClientAct(resolvedClientID)

	// OpenClaw sends models as "provider/model-id" (e.g. "wall-vault/gemini-2.5-flash",
	// "anthropic/claude-opus-4-6"). Parse and route accordingly.
	svc, mdl = parseProviderModel(svc, mdl)

	geminiReq := OpenAIToGemini(&oaiReq)
	dispatchRes, err := s.dispatch(r.Context(), svc, mdl, geminiReq)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	geminiResp := dispatchRes.Response
	// Reflect the actual service/model used — may differ from the requested
	// pair when dispatch fell back to another provider (see DispatchResult).
	if dispatchRes.UsedModel != "" {
		mdl = dispatchRes.UsedModel
	}
	if dispatchRes.UsedService != "" {
		svc = dispatchRes.UsedService
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
			text:         stripControlTokens(extractTextAndMediaNotes(c.Content.Parts)),
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

	oaiResp := &OpenAIResponse{
		ID:     "chatcmpl-proxy",
		Object: "chat.completion",
		Model:  mdl,
	}
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
	r.Body = http.MaxBytesReader(w, r.Body, maxAIBodySize)

	var req AnthropicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if stripped := s.filter.FilterAnthropic(&req); stripped > 0 {
		log.Printf("[Security] blocked %d tools from anthropic request (client=%s)", stripped, s.cfg.Proxy.ClientID)
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
	// Anthropic API uses x-api-key header instead of Authorization: Bearer,
	// so check both to support Claude Code and other Anthropic-format clients.
	var resolvedClientID string
	reqToken := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else if xKey := r.Header.Get("x-api-key"); xKey != "" {
		reqToken = xKey
	}
	if reqToken != "" {
		if entry := s.lookupTokenConfig(reqToken); entry != nil {
			resolvedClientID = entry.clientID
			if entry.service != "" {
				svc = entry.service
			}
			if entry.model != "" {
				mdl = entry.model
			}
		}
	}
	defer s.refreshClientAct(resolvedClientID)

	// Parse provider/model form (e.g. "anthropic/claude-opus-4-6")
	svc, mdl = parseProviderModel(svc, mdl)

	// Native Anthropic passthrough: only when the resolved service IS anthropic
	// (or the model is a Claude model). Non-Claude models (e.g. google/gemini-*)
	// must go through dispatch() so they are routed to the correct backend.
	// Previously this tried passthrough for ALL models when anthropic was allowed,
	// which silently forced non-Claude models to claude-haiku and returned wrong results.
	usePassthrough := false
	if svc == "anthropic" || strings.HasPrefix(mdl, "claude-") {
		if len(allowedServices) == 0 {
			usePassthrough = true
		} else {
			for _, sv := range allowedServices {
				if sv == "anthropic" {
					usePassthrough = true
					break
				}
			}
		}
	}
	if usePassthrough {
		if body, _, err := s.callAnthropicPassthrough(r.Context(), &req, mdl); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(body) //nolint:errcheck
			return
		} else {
			log.Printf("[anthropic] passthrough failed → fallback to dispatch: %v", err)
		}
	}

	// Fallback: convert to GeminiRequest and dispatch via Google/OpenRouter.
	// AnthropicToGemini is lossy (drops tool_calls); attach an OpenAI-equivalent
	// via RawOAI so that OpenRouter-based fallback preserves tools.
	geminiReq := AnthropicToGemini(&req)
	geminiReq.RawOAI = anthropicToOpenAIReq(&req, mdl)
	dispatchRes, err := s.dispatch(r.Context(), svc, mdl, geminiReq)
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
	geminiResp := dispatchRes.Response
	// Use the actual serving model in the Anthropic response so a fallback
	// from claude-* to (e.g.) google/gemini-flash is accurately reflected.
	if dispatchRes.UsedModel != "" {
		mdl = dispatchRes.UsedModel
	}

	resp := GeminiRespToAnthropic(mdl, geminiResp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── OpenAI-compatible model list (/v1/models) ────────────────────────────────

func (s *Server) handleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	type oaiModel struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	}
	var data []oaiModel

	// If request comes from a client-specific token (e.g. Cline's own token),
	// return that client's configured model — not the proxy's own model.
	curModel := ""
	curService := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken := strings.TrimPrefix(authHeader, "Bearer ")
		if entry := s.lookupTokenConfig(reqToken); entry != nil && entry.model != "" {
			curModel = entry.model
			curService = entry.service
		}
	}
	// Fall back to proxy's own model
	if curModel == "" {
		s.mu.RLock()
		curModel = s.model
		curService = s.service
		s.mu.RUnlock()
	}

	if curModel != "" {
		data = []oaiModel{{ID: curModel, Object: "model", OwnedBy: curService}}
		// When the vault-configured model is a non-Claude model (e.g. google/gemini-*),
		// Claude Code's settings.json still has a Claude model (because updateClaudeCodeModel
		// skips non-Claude models). Include common Claude aliases so Claude Code's model
		// validation passes regardless of what alias is in settings.json.
		if !strings.HasPrefix(curModel, "claude-") {
			for _, alias := range []string{"opus", "sonnet", "haiku"} {
				if alias != curModel {
					data = append(data, oaiModel{ID: alias, Object: "model", OwnedBy: "proxy"})
				}
			}
		}
	} else {
		// Standalone mode (no vault): return full registry list.
		for _, m := range s.registry.All("") {
			data = append(data, oaiModel{ID: m.ID, Object: "model", OwnedBy: m.Service})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   data,
	})
}

// ─── Request Dispatch ─────────────────────────────────────────────────────────

// DispatchResult carries the dispatch response along with which service/model
// actually produced it. Callers use UsedService/UsedModel to render an accurate
// `model` field in their OpenAI/Anthropic response bodies — otherwise a fallback
// from (e.g.) anthropic/claude-sonnet-4-6 → google/gemini-flash would still be
// labeled as the original Claude model, making traces misleading.
type DispatchResult struct {
	Response    *GeminiResponse
	UsedService string
	UsedModel   string
}

func (s *Server) dispatch(ctx context.Context, service, model string, req *GeminiRequest) (*DispatchResult, error) {
	s.mu.RLock()
	allowedServices := s.allowedServices
	serviceDefaults := s.serviceDefaults
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
		// Fast-skip cloud services whose keys are all on cooldown or exhausted
		// — prevents the dispatch chain from spending seconds on forced retries
		// that will re-hit 429/402 and extend the cooldown. Local services
		// (ollama/lmstudio/vllm) carry no keys and are always tried.
		needsKey := svc != "ollama" && svc != "lmstudio" && svc != "vllm" && svc != "llamacpp"
		if needsKey && !s.keyMgr.CanServe(svc) {
			log.Printf("[proxy] %s skip: 전체 쿨다운 또는 등록된 키 없음", svc)
			lastErr = fmt.Errorf("서비스 '%s' 전체 쿨다운", svc)
			continue
		}

		// Primary respects the caller's requested model; fallback swaps in the
		// target service's default_model when available so Anthropic doesn't
		// receive "gemini-2.5-flash" etc. If the target has no default_model,
		// the original is tried (better to log a 404 than silently skip).
		targetModel := model
		if svc != service {
			if dm := serviceDefaults[svc]; dm != "" {
				targetModel = dm
			}
		}

		var resp *GeminiResponse
		var err error
		switch svc {
		case "google":
			resp, err = s.callGoogle(ctx, targetModel, req)
		case "openrouter":
			resp, err = s.callOpenRouter(ctx, targetModel, req)
		case "ollama":
			resp, err = s.callOllama(ctx, targetModel, req)
		case "openai":
			resp, err = s.callOpenAI(ctx, targetModel, req)
		case "anthropic":
			resp, err = s.callAnthropic(ctx, targetModel, req)
		case "lmstudio", "vllm", "llamacpp":
			resp, err = s.callLocalService(ctx, svc, targetModel, req)
		default:
			continue
		}
		if err == nil {
			if svc != service {
				log.Printf("[proxy] fallback: %s/%s → %s/%s", service, model, svc, targetModel)
				s.hooksMgr.Fire(hooks.EventModelChanged, map[string]string{
					"service": svc,
					"model":   targetModel,
				})
			}
			return &DispatchResult{Response: resp, UsedService: svc, UsedModel: targetModel}, nil
		}
		log.Printf("[proxy] %s (%s) failed → fallback: %v", svc, targetModel, err)
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

func (s *Server) callGoogle(ctx context.Context, model string, req *GeminiRequest) (*GeminiResponse, error) {
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
		resp, err := s.doRequest(ctx, "POST", url, data, nil)
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

func (s *Server) callOpenRouter(ctx context.Context, model string, req *GeminiRequest) (*GeminiResponse, error) {
	resp, err := s.callOpenRouterModel(ctx, model, req)
	if err == nil {
		return resp, nil
	}
	// If paid model failed with payment errors, retry with free-tier variant
	if !strings.HasSuffix(model, ":free") {
		freeModel := model + ":free"
		log.Printf("[proxy] openrouter paid failed, retrying free tier: %s", freeModel)
		if resp, err2 := s.callOpenRouterModel(ctx, freeModel, req); err2 == nil {
			return resp, nil
		}
	}
	return nil, err
}

func (s *Server) callOpenRouterModel(ctx context.Context, model string, req *GeminiRequest) (*GeminiResponse, error) {
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
		resp, err := s.doRequest(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", data, headers)
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

func (s *Server) callOpenAI(ctx context.Context, model string, req *GeminiRequest) (*GeminiResponse, error) {
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
		resp, err := s.doRequest(ctx, "POST", "https://api.openai.com/v1/chat/completions", data, headers)
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

func (s *Server) callOllama(ctx context.Context, model string, req *GeminiRequest) (*GeminiResponse, error) {
	// v0.2: Ollama name-mismatch heuristic removed.
	// dispatch_v2.go's dispatchWith() now uses each service's default_model,
	// so a Gemini/Claude model id can never reach the Ollama endpoint
	// structurally. See ResolveModel + dispatchWith.
	ollamaURL := s.ollamaURL()

	// Wait for a free slot (bounded by ollamaSem capacity), but bail out if
	// the caller's context was cancelled in the meantime.
	select {
	case s.ollamaSem <- struct{}{}:
		defer func() { <-s.ollamaSem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Use Ollama's OpenAI-compatible /v1/chat/completions endpoint.
	// The native /api/chat expects tool_calls.function.arguments as a JSON object,
	// while OpenAI format (and our internal representation) uses a JSON string —
	// sending the OpenAI format to /api/chat causes HTTP 400.
	// /v1/chat/completions accepts the standard OpenAI format including arguments-as-string.
	oaiReq := GeminiToOpenAI(model, req)
	oaiReq.Stream = false
	s.mu.RLock()
	oaiReq.Reasoning = s.serviceReasoning["ollama"]
	s.mu.RUnlock()
	data, _ := json.Marshal(oaiReq)

	// Ollama is a local service: inference can take several minutes for large models.
	// Use a dedicated client with no hard timeout so generation is never cut off.
	// Connection errors (server not running) still surface immediately.
	httpReq, err := http.NewRequestWithContext(ctx, "POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Ollama 요청 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	ollamaClient := &http.Client{Timeout: 10 * time.Minute} // generous timeout for local inference
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

// callLocalService: generic OpenAI-compatible local server (LM Studio, vLLM, etc.)
// Generic OpenAI-compatible local server handler.
func (s *Server) callLocalService(ctx context.Context, serviceID, model string, req *GeminiRequest) (*GeminiResponse, error) {
	s.mu.RLock()
	baseURL := s.serviceURLs[serviceID]
	s.mu.RUnlock()
	if baseURL == "" {
		defaults := map[string]string{"lmstudio": "http://localhost:1234", "vllm": "http://localhost:8000"}
		baseURL = defaults[serviceID]
	}
	if baseURL == "" {
		return nil, fmt.Errorf("%s: URL 미설정", serviceID)
	}

	// Strip provider prefix from model (e.g. "google/gemma-4-26b-a4b" → "gemma-4-26b-a4b")
	if i := strings.Index(model, "/"); i >= 0 {
		model = model[i+1:]
	}

	oaiReq := GeminiToOpenAI(model, req)
	oaiReq.Stream = false
	s.mu.RLock()
	oaiReq.Reasoning = s.serviceReasoning[serviceID]
	s.mu.RUnlock()
	data, _ := json.Marshal(oaiReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s 요청 생성 실패: %w", serviceID, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s 연결 실패 (%s): %w", serviceID, baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s 오류: HTTP %d: %s", serviceID, resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	var oaiResp OpenAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("%s 응답 파싱 오류: %w", serviceID, err)
	}
	if oaiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", serviceID, oaiResp.Error.Message)
	}
	return OpenAIRespToGemini(&oaiResp), nil
}

// ─── Common HTTP Request ──────────────────────────────────────────────────────

func (s *Server) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
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
		case strings.HasPrefix(bare, "gemini-"), strings.HasPrefix(bare, "gemma-"):
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
