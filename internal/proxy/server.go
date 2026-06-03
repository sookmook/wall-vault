package proxy

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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
	clientID         string // vault client ID for this token
	service          string
	model            string
	fallbackServices []string // ordered fallback chain; empty = strict primary-only
	allowedServices  []string // security whitelist; empty = no restriction
	expiresAt        time.Time
}

// clientAct: last-seen activity record for a non-proxy client served by this proxy
type clientAct struct {
	service  string
	model    string
	lastSeen time.Time
	applied  bool // true if config was applied via /agent/apply — never expires from heartbeat
}

// hostAgent is a vault client that shares this proxy's physical host (matched
// by Client.Host == os.Hostname()). Cached on syncFromVault, emitted verbatim
// into each heartbeat's ActiveClients list — trusting the operator-assigned
// Host field is simpler and more reliable than per-type process probes for
// agents whose liveness can't be inferred from a single pgrep (VSCode
// extensions, Windows-side clients behind WSL, etc.).
type hostAgent struct {
	ClientID  string
	AgentType string
	Service   string
	Model     string
}

// Server: proxy HTTP server
type Server struct {
	cfg             *config.Config
	mu              sync.RWMutex
	service         string            // user-configured preferred service (from vault dashboard)
	model           string            // user-configured preferred model (from vault dashboard)
	claudeCodeClientID string            // vault client ID for the local claude-code agent (from syncFromVault)
	ownAgentType       string            // this proxy's own agent_type (from syncFromVault)
	// ownFallback is the proxy's own client config's fallback_services list,
	// applied to incoming requests that do NOT carry a token (or whose token
	// is the proxy's own VaultToken — explicitly excluded by lookupTokenConfig).
	// Local OpenClaw / Claude Code processes typically call localhost:56244
	// without an Authorization header, so without this they would otherwise
	// run strict-by-default and 502 the moment the inferred service is on
	// cooldown. Loaded from /api/clients and refreshed on every syncFromVault.
	ownFallback        []string
	// ownModelOverride mirrors the proxy's own client_id's model_override and
	// is applied to token-less calls so the operator's vault-configured model
	// wins over whatever the local OpenClaw / Claude Code happens to put in
	// the request body. Without this, an operator who switches host-B to
	// OpenRouter via vault still has to chase down OpenClaw's primary model
	// in openclaw.json. With it, vault is the single source of truth.
	ownModelOverride   string
	// hostAgents is the set of vault clients whose Host field matches this
	// proxy's os.Hostname(). Populated by syncFromVault, consumed by sendHeartbeat
	// to report every co-hosted sub-client in a single heartbeat — so a host
	// running N agents lights up N signal lights instead of one. Self-reference
	// (cfg.ClientID) is filtered out; the main heartbeat already covers self.
	hostAgents         []hostAgent
	allowedServices []string          // proxy-enabled services from vault (empty = no restriction)
	serviceURLs      map[string]string // service ID → local URL from vault config
	serviceDefaults  map[string]string // service ID → default_model from vault config
	serviceReasoning map[string]bool   // service ID → reasoning_mode toggle from vault config
	keyMgr          *KeyManager
	filter          *ToolFilter
	sse             *SSEClient
	registry        *models.Registry
	hooksMgr        *hooks.Manager
	// localSems caps concurrent requests per local inference backend. Each
	// service gets its own cap-1 buffered channel that doubles as a
	// context-aware semaphore — the acquire site selects on ctx.Done() so
	// a caller whose HTTP request is cancelled won't keep holding a slot
	// behind a slow upstream. Populated for every local backend wall-vault
	// knows about (ollama, llamacpp, lmstudio, vllm) regardless of whether
	// the current config enables them, so a later config reload that adds
	// a service does not race against a missing entry.
	localSems       map[string]chan struct{}

	// stopCh is closed by Stop() to signal background goroutines (periodic
	// vault sync, token-cache eviction, initial-load delay) to exit instead
	// of leaking past server shutdown. systemd `Restart=always` or a
	// launchctl unload then runs with a clean slate.
	stopCh          chan struct{}
	tokenCacheMu    sync.RWMutex
	tokenCache      map[string]*tokenCacheEntry // Bearer token → client model config
	clientActMu     sync.Mutex
	clientActs      map[string]*clientAct // clientID → last-seen activity (for heartbeat reporting)
	// ollamaHTTP is a long-lived client used for every callOllama. Reusing
	// the underlying transport keeps idle TCP/keepalive connections to
	// :11434 warm so we save a few ms per call (TLS isn't in play locally,
	// but the syscall + handshake cost still adds up under sustained
	// traffic). Built once in NewServer.
	ollamaHTTP *http.Client

	// dispatchHTTP / dispatchSSEHTTP are shared keep-alive pools for every
	// non-Ollama upstream call. Before they existed, callLocalService,
	// streamLocalService and doRequest each constructed a fresh
	// &http.Client{} per request whose Transport defaulted to
	// http.DefaultTransport — and DefaultTransport.MaxIdleConnsPerHost is
	// 2. When a single client (e.g. EconoWorld's 5 s heartbeat) drives
	// sustained dispatch to one upstream host (a llama.cpp / lmstudio /
	// vllm / cloud endpoint), the 2-conn limit forces new TCP+TLS
	// handshakes constantly, piles TIME_WAIT, and eventually starves
	// ephemeral ports — visible as occasional dispatch stalls. Sharing a
	// dedicated Transport with MaxIdleConnsPerHost: 20 lets keep-alive
	// actually work. SSE variant keeps Timeout: 0 so streaming responses
	// rely on the caller ctx deadline, never the client-wide timeout.
	dispatchHTTP    *http.Client
	dispatchSSEHTTP *http.Client

	// servicePools routes each local-service dispatch to whichever instance
	// hosts the requested model. Keyed by service id. The "ollama" pool uses
	// /api/tags for discovery; every other local service (lmstudio, vllm,
	// llamacpp, custom OpenAI-compat plugins) uses /v1/models. Single-URL
	// deployments behave identically to the pre-pool path. URL list is
	// rebuilt from env (WV_<ID>_URLS list + WV_<ID>_URL) ∪ plugin DefaultURL
	// ∪ vault serviceURLs[id] every vault sync. Operators can also enter a
	// comma-separated list directly into the dashboard URL field.
	servicePools map[string]*servicePool
	poolsMu      sync.RWMutex

	// pluginByID indexes cfg.Plugins by service id so the dispatch path
	// can look up auth / TLS / default_url settings in O(1). Empty map
	// when no plugin yamls are present — every existing call site keeps
	// working unchanged.
	pluginByID map[string]*config.ServicePlugin
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
		// Each local backend stays serialized (cap 1): large local models
		// are typically memory-bound, and running two inferences concurrently
		// tends to be slower than two sequential ones. Keyed by service id
		// so llamacpp / lmstudio / vllm queue independently from ollama even
		// when they share a host.
		localSems: map[string]chan struct{}{
			"ollama":   make(chan struct{}, 1),
			"llamacpp": make(chan struct{}, 1),
			"lmstudio": make(chan struct{}, 1),
			"vllm":     make(chan struct{}, 1),
		},
		stopCh: make(chan struct{}),
	}

	// Index plugin yamls by service id. The plugin schema in
	// internal/config/services.go carries auth / tls_internal_ca /
	// default_url / default_model, which the dispatch path consults below
	// when calling a service whose backend is "another wall-vault" or any
	// non-default OpenAI-compat local LLM (LM Studio, vLLM, llama.cpp,
	// LocalAI, Jan, KoboldCpp, TabbyAPI, mlx_lm.server, LiteLLM, etc.).
	// A plugin with an empty/missing auth.type and tls_internal_ca=false
	// reproduces the pre-v0.2.44 behaviour exactly.
	if len(cfg.Plugins) > 0 {
		s.pluginByID = make(map[string]*config.ServicePlugin, len(cfg.Plugins))
		for i := range cfg.Plugins {
			s.pluginByID[cfg.Plugins[i].ID] = &cfg.Plugins[i]
		}
		// A plugin yaml with enabled=true is the operator's declaration
		// that this proxy can handle that service. Append the id to
		// cfg.Proxy.Services if absent so the dispatcher actually picks
		// the service up — without this, an installed plugin would sit
		// inert because cfg.Proxy.Services (default: google/openrouter/
		// ollama) gates which services dispatch_v2 even tries.
		known := map[string]bool{}
		for _, id := range cfg.Proxy.Services {
			known[id] = true
		}
		for _, p := range cfg.Plugins {
			if p.Enabled && !known[p.ID] {
				cfg.Proxy.Services = append(cfg.Proxy.Services, p.ID)
				known[p.ID] = true
			}
			// Auto-register OAI-compat plugins so parseProviderModelDepth
			// and dispatch handle "publisher/model" routing without a Go-side
			// edit. RequestFormat default is "openai" because the historic
			// behaviour predates the field; only explicit non-openai shapes
			// (gemini / ollama / raw) are excluded.
			if p.Enabled {
				switch p.RequestFormat {
				case "", "openai":
					registerOAICompatPlugin(p.ID)
				}
			}
		}
		// Surface plugin URLs that look like a misconfiguration (a remote
		// host reached over plain HTTP — for any wall-vault-style backend
		// this should be HTTPS so the bearer token isn't sent in the
		// clear). One log line per offending plugin at boot; no fatal
		// behaviour because the operator may legitimately be running a
		// behind-the-firewall http-only test backend.
		for _, p := range cfg.Plugins {
			warnIfPluginURLLooksRemoteHTTP(p)
		}
	}

	s.keyMgr = NewKeyManager(cfg.Proxy.VaultURL, cfg.Proxy.VaultToken, cfg.Proxy.ClientID)
	s.filter = NewToolFilter(FilterMode(cfg.Proxy.ToolFilter), cfg.Proxy.AllowedTools)
	// Dedicated long-lived Ollama HTTP client. The 10-minute timeout is the
	// same generous bound the inline client used; the Transport tuning is
	// new — without it every callOllama spun up a fresh client whose
	// connection went straight to TIME_WAIT after the response, defeating
	// HTTP keep-alive.
	s.ollamaHTTP = &http.Client{
		Timeout: 10 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     120 * time.Second,
		},
	}
	// Dispatch pools for everything other than Ollama. Sized larger than
	// ollamaHTTP because llamacpp / lmstudio / vllm / cloud upstreams can
	// see more concurrent fan-out from a single bursty client. Separate
	// Transport instances so SSE long-poll connections don't crowd the
	// idle slot accounting for short-lived non-stream requests.
	s.dispatchHTTP = &http.Client{
		Timeout: cfg.Proxy.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     120 * time.Second,
		},
	}
	s.dispatchSSEHTTP = &http.Client{
		Timeout: 0, // ctx deadline owns the per-request bound
		Transport: &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     120 * time.Second,
		},
	}

	// Per-service dispatch pools. Built-in local services (ollama plus the
	// three OpenAI-compatible defaults) always get a pool; plugin-defined
	// services with an OAI-compatible request_format join the rotation too.
	// Vault sync later re-merges in serviceURLs[id]. Started below so the
	// first request post-NewServer already sees a populated model→URL map.
	// Stop() tears the goroutines down at shutdown.
	s.servicePools = map[string]*servicePool{}
	s.servicePools["ollama"] = newServicePool("ollama", s.resolveOllamaURLs(), fetchOllamaTags, s.ollamaHTTP)
	for _, id := range []string{"lmstudio", "vllm", "llamacpp"} {
		s.servicePools[id] = newServicePool(id, s.resolveLocalServiceURLs(id), fetchOpenAICompatModels, s.ollamaHTTP)
	}
	for id, plugin := range s.pluginByID {
		if _, exists := s.servicePools[id]; exists {
			continue
		}
		// Only OpenAI-compatible plugins expose /v1/models reliably. The
		// other RequestFormat values (gemini, ollama, raw) need a different
		// discovery path or have no list endpoint at all — skip them so a
		// dead /v1/models call doesn't spam the log.
		if plugin.RequestFormat != "" && plugin.RequestFormat != "openai" {
			continue
		}
		s.servicePools[id] = newServicePool(id, s.resolveLocalServiceURLs(id), fetchOpenAICompatModels, s.ollamaHTTP)
	}
	for _, p := range s.servicePools {
		p.Start(context.Background())
	}

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
			} else if svc != "" {
				// Vault soft-cleared the override (operator picked "(use
				// service default)" in dashboard). Fill from the new
				// service's default_model so downstream sync writers
				// (OpenClaw, cline, EconoWorld) propagate a coherent
				// model id matching what dispatch will actually route.
				if def := s.serviceDefaults[svc]; def != "" {
					s.model = def
				}
			}
			newSvc, newMdl := s.service, s.model
			s.mu.Unlock()
			if newSvc != oldSvc || newMdl != oldMdl {
				s.hooksMgr.Fire(hooks.EventModelChanged, map[string]string{
					"service": newSvc,
					"model":   newMdl,
				})
				go updateOpenClawJSON(newSvc, newMdl, s.defaultProxyOrigin())
			}
		}, func(clientID, agentType, svc, mdl string) {
			// Foreign client model changed in vault — update local agent config if applicable.
			// When mdl is empty (operator picked "(use service default)") fall back to
			// the new service's default_model so the local config writers propagate
			// a coherent model id matching what dispatch will actually route.
			effective := mdl
			if effective == "" && svc != "" {
				s.mu.RLock()
				effective = s.serviceDefaults[svc]
				s.mu.RUnlock()
			}
			switch agentType {
			case "cline":
				if effective != "" {
					go updateClineModel(effective)
				}
			case "claude-code":
				if effective != "" {
					go updateClaudeCodeModel(effective)
				}
			case "econoworld":
				if effective != "" {
					go updateEconoWorldModel(effective)
				}
				// Mirror the live ollama reasoning toggle into
				// EconoWorld's ai_config.json so _call_ollama_native
				// picks up dashboard changes without a manual edit.
				s.mu.RLock()
				ollamaThink := s.serviceReasoning["ollama"]
				s.mu.RUnlock()
				go updateEconoWorldOllamaThink(ollamaThink)
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

		// Initial vault sync — synchronous so the first inbound request
		// after Handler() goes live can find keys / services / cooldown
		// state in place. Without this, dispatch raced the 2s-delayed
		// goroutine and reported "전체 쿨다운 또는 등록된 키 없음" when
		// the keyMgr was simply empty.
		//
		// Retried with backoff because vault + proxy frequently share a
		// host (any local-vault deploy) and the proxy can lose the race
		// for the listen socket — the first sync hits "connection refused"
		// while vault is still binding, returns, and dispatch then sees
		// the empty keyMgr for the full 5-minute periodic gap (observed
		// on the v0.2.82 deploy). Loop until at least one key is loaded
		// or the budget is exhausted; fall through either way so a
		// permanently-down vault doesn't block startup forever — the
		// periodic sync still retries.
		for i := 0; i < 10; i++ {
			s.syncFromVault()
			if s.keyMgr.HasAnyKey() {
				break
			}
			select {
			case <-time.After(1 * time.Second):
			case <-s.stopCh:
				return s
			}
		}

		// periodic key sync (every 5 minutes)
		go func() {
			// Phase-shift the sync cadence by a client-specific offset
			// so proxies booted within the same second do not re-pull
			// vault state in lock-step. Bounded by the 5-minute period.
			const syncPeriodMs = 5 * 60 * 1000
			if offset := AgentOffset(s.cfg.Proxy.ClientID, syncPeriodMs); offset > 0 {
				select {
				case <-time.After(offset):
				case <-s.stopCh:
					return
				}
			}

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

	// Sanitize ~/.openclaw/openclaw.json once at boot. OpenClaw 2026.4.29
	// rejects model entries with empty `id`, which historically slipped
	// past pre-guard versions of applyOpenClawConfig — we observed a
	// host-A gateway crash-loop from a single such entry.
	// No-op for hosts that don't run OpenClaw (most of the fleet).
	runStartupSanitize()

	// Heal stale provider settings once at boot. The sanitize pass above
	// only removes empty-id entries; the heal pass forces baseUrls back
	// to localhost and prunes duplicate / dangling-name entries. Same
	// host-A root cause: an external host got written into providers.custom
	// and providers.anthropic.baseUrl, bypassing the proxy entirely.
	// CA bundle location: alongside the proxy TLS cert as `ca.crt`. The
	// heal pass writes it into models.providers.<id>.request.tls.ca for
	// any operator-configured client that does honour that hint.
	caBundlePath := ""
	if cfg.Proxy.TLS.CertFile != "" {
		candidate := filepath.Join(filepath.Dir(cfg.Proxy.TLS.CertFile), "ca.crt")
		if _, err := os.Stat(candidate); err == nil {
			caBundlePath = candidate
		}
	}
	// Same-host clients that can't be coerced into trusting the proxy's
	// self-signed CA ((operator host, earlier): OpenClaw's macOS daemon rewrites
	// NODE_EXTRA_CA_CERTS to /etc/ssl/cert.pem at spawn, dropping any
	// operator-provided CA hint) use a loopback-only plain-HTTP
	// companion listener instead — OpenClaw never sees TLS, so the
	// trust problem disappears. Heal pass routes any same-host
	// OpenClaw config to that port. LAN callers keep using the TLS
	// listener with ca.crt. localBaseURL is empty when TLS is off (no
	// companion needed) or when the operator disabled it.
	localBaseOrigin := ""
	if cfg.Proxy.TLS.Enabled && cfg.Proxy.PlainPort > 0 {
		// Origin only (no /v1 path) — providerHealURLs appends /v1 for
		// the OpenAI-compat custom provider and leaves the bare origin
		// for anthropic / google.
		localBaseOrigin = fmt.Sprintf("http://127.0.0.1:%d", cfg.Proxy.PlainPort)
	}
	defaultOrigin := DefaultProxyOrigin(cfg.Proxy.Port, cfg.Proxy.TLS.Enabled)
	runStartupOpenClawHeal(cfg.Proxy.VaultToken, caBundlePath, localBaseOrigin, defaultOrigin)
	// Same loopback-companion benefit applies to EconoWorld's
	// analyzer/ai_config.json on hosts that have it. Hosts without an
	// EconoWorld install no-op silently.
	runStartupEconoWorldHeal(
		localBaseOrigin,
		cfg.Proxy.EconoWorldMaxTokens,
		cfg.Proxy.EconoWorldStream,
		cfg.Proxy.EconoWorldRequestTimeout,
	)

	return s
}

// defaultProxyOrigin returns the origin every config writer should target on
// this proxy: scheme + localhost + the operator's chosen port. Used by
// updateOpenClawJSON / agent_apply / setup wizard generators so the canonical
// "where do clients reach us" answer comes from one place instead of
// duplicated `https://localhost:56244` literals scattered through the code.
func (s *Server) defaultProxyOrigin() string {
	return DefaultProxyOrigin(s.cfg.Proxy.Port, s.cfg.Proxy.TLS.Enabled)
}

// refreshAllServicePoolURLs re-resolves every service pool's URL list and
// pushes it via SetURLs. Called from syncAllowedServices after vault sync
// updates s.serviceURLs. SetURLs is a no-op when the list is unchanged, so
// machines whose vault config doesn't move see no extra polling.
func (s *Server) refreshAllServicePoolURLs() {
	s.poolsMu.RLock()
	pools := make(map[string]*servicePool, len(s.servicePools))
	for id, p := range s.servicePools {
		pools[id] = p
	}
	s.poolsMu.RUnlock()
	for id, p := range pools {
		if id == "ollama" {
			p.SetURLs(s.resolveOllamaURLs())
			continue
		}
		p.SetURLs(s.resolveLocalServiceURLs(id))
	}
}

// resolveLocalServiceURLs builds the multi-instance URL list for a local
// service's dispatch pool. Sources, in order, with the first occurrence
// winning on duplicates:
//
//  1. WV_<ID>_URLS env (comma list)
//  2. WV_<ID>_URL env (single — accepts comma list too)
//  3. plugin.DefaultURL (single — accepts comma list)
//  4. vault serviceURLs[id] (single from SSE — accepts comma list)
//  5. Built-in fallbacks for the historic three (lmstudio / vllm / llamacpp)
//
// Every textual source goes through splitOllamaURLs so an operator who types
// "http://a:1234, http://b:1234" into the dashboard URL field gets the same
// multi-instance behaviour as the dedicated env list.
//
// Returns nil when nothing produces a URL — callers must detect that and
// surface a config error rather than dispatching to "".
func (s *Server) resolveLocalServiceURLs(serviceID string) []string {
	if serviceID == "" {
		return nil
	}
	envBase := "WV_" + strings.ToUpper(strings.ReplaceAll(serviceID, "-", "_"))
	out := make([]string, 0, 4)
	out = append(out, splitOllamaURLs(os.Getenv(envBase+"_URLS"))...)
	out = append(out, splitOllamaURLs(os.Getenv(envBase+"_URL"))...)
	if plugin := s.pluginByID[serviceID]; plugin != nil && plugin.DefaultURL != "" {
		out = append(out, splitOllamaURLs(plugin.DefaultURL)...)
	}
	s.mu.RLock()
	u := s.serviceURLs[serviceID]
	s.mu.RUnlock()
	out = append(out, splitOllamaURLs(u)...)
	if len(out) == 0 {
		defaults := map[string]string{
			"lmstudio": "http://localhost:1234",
			"vllm":     "http://localhost:8000",
			"llamacpp": "http://localhost:8080",
		}
		if d, ok := defaults[serviceID]; ok {
			out = append(out, d)
		}
	}
	return out
}

// resolveLocalServiceURL returns the first URL for the service. Kept as a
// thin shim for callers that need a single representative URL. Multi-
// instance dispatch goes through resolveLocalServiceURLForModel.
func (s *Server) resolveLocalServiceURL(serviceID string) string {
	if serviceID == "" {
		return ""
	}
	urls := s.resolveLocalServiceURLs(serviceID)
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

// resolveLocalServiceURLForModel returns the URL of the instance that hosts
// the given model id, or the first configured URL on miss (so the upstream
// returns a real 404). Single-URL setups resolve every model to that URL
// — same as the pre-pool resolveLocalServiceURL.
func (s *Server) resolveLocalServiceURLForModel(serviceID, model string) string {
	if serviceID == "" {
		return ""
	}
	s.poolsMu.RLock()
	pool := s.servicePools[serviceID]
	s.poolsMu.RUnlock()
	if pool != nil {
		if u := pool.URLForModel(model); u != "" {
			return u
		}
	}
	return s.resolveLocalServiceURL(serviceID)
}

// applyQwen3NoThinkSuffix appends the "/no_think" inline directive to the
// last user-role message in oaiReq when reasoning is disabled, the model
// is a qwen3 variant, and the target backend is known to honour the
// inline marker. Centralises the rule previously duplicated in callOllama,
// callLocalService, and the streaming variants.
//
// Backends opt in either via:
//   - native built-ins: serviceID == "ollama" || serviceID == "lmstudio"
//   - plugin yaml: inline_no_think_for_qwen3: true
//
// Any other backend silently no-ops (the marker would echo back as
// literal text in the response).
func (s *Server) applyQwen3NoThinkSuffix(oaiReq *OpenAIRequest, serviceID, model string, reasoning bool) {
	if reasoning {
		return
	}
	if !strings.HasPrefix(strings.ToLower(model), "qwen3") {
		return
	}
	allowed := false
	switch serviceID {
	case "ollama", "lmstudio":
		allowed = true
	default:
		if plugin := s.pluginByID[serviceID]; plugin != nil && plugin.InlineNoThinkForQwen3 {
			allowed = true
		}
	}
	if !allowed {
		return
	}
	for i := len(oaiReq.Messages) - 1; i >= 0; i-- {
		if oaiReq.Messages[i].Role == "user" {
			oaiReq.Messages[i].Content = strings.TrimRight(oaiReq.Messages[i].Content, " \n") + " /no_think"
			break
		}
	}
}

// serviceNeedsKey reports whether dispatch should consult keyMgr for the
// given service before attempting a call. Cloud services with rotated
// upstream API keys return true; local backends (ollama / lmstudio / vllm
// / llamacpp) and plugin-defined services whose auth.type is "" / "none"
// / "bearer" return false because they either have no upstream credential
// or use this proxy's own vault token rather than rotated cloud keys.
//
// Replaces the previous hardcoded `svc != "ollama" && svc != "lmstudio" && …`
// chain so a plugin can opt out of the keyMgr gate via its yaml without
// requiring a Go-side edit.
func (s *Server) serviceNeedsKey(svc string) bool {
	switch svc {
	case "ollama", "lmstudio", "vllm", "llamacpp":
		return false
	case "google", "openrouter", "openai", "anthropic":
		return true
	}
	if plugin := s.pluginByID[svc]; plugin != nil {
		switch plugin.Auth.Type {
		case "", "none", "bearer":
			return false
		}
		return true
	}
	// Unknown service id: conservatively skip the keyMgr gate so the
	// switch in dispatchWithChain can return a clearer error.
	return false
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
	s.poolsMu.RLock()
	pools := make([]*servicePool, 0, len(s.servicePools))
	for _, p := range s.servicePools {
		pools = append(pools, p)
	}
	s.poolsMu.RUnlock()
	for _, p := range pools {
		p.Stop()
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
// tokenLookupReason categorises why a token lookup failed, so callers can
// emit specific 401/503 messages instead of the old generic "invalid token"
// — the latter conflated "client not registered", "vault unreachable", and
// "everything fine but the token literally isn't in the registry", which made
// 401 triage frustrating (see earlier reviewer feedback wasting time on IP
// whitelist when the cause was actually vault-side).
type tokenLookupReason int

const (
	tokenLookupOK              tokenLookupReason = iota // entry returned
	tokenLookupNotRegistered                            // vault explicitly returned 4xx
	tokenLookupVaultUnreachable                         // network / dial / timeout
	tokenLookupVaultError                               // vault 5xx or malformed response
	tokenLookupSkipped                                  // empty config (vault_url unset) — caller should fall back
	tokenLookupVaultStale                               // vault unreachable, returning an entry past expiry but still inside VaultStaleGrace
)

func (s *Server) lookupTokenConfig(token string) *tokenCacheEntry {
	entry, _ := s.lookupTokenConfigDetailed(token)
	return entry
}

// lookupTokenConfigDetailed wraps lookupTokenConfig with a reason code so
// auth middlewares can pick a precise error message. Cache hits are reported
// as tokenLookupOK; misses / errors carry the actual cause.
func (s *Server) lookupTokenConfigDetailed(token string) (*tokenCacheEntry, tokenLookupReason) {
	if s.cfg.Proxy.VaultURL == "" || token == "" {
		return nil, tokenLookupSkipped
	}
	// skip our own proxy token (it is already applied via s.service/s.model)
	if token == s.cfg.Proxy.VaultToken {
		return nil, tokenLookupSkipped
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
		return e, tokenLookupOK
	}
	s.tokenCacheMu.RUnlock()

	// fetch from vault
	req, err := http.NewRequest("GET", s.cfg.Proxy.VaultURL+"/api/token/config", nil)
	if err != nil {
		return nil, tokenLookupVaultError
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := internalHTTPClient(3 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		// Vault unreachable. If a previously validated entry is still inside
		// the configured grace window, return it stale so a brief vault
		// blip doesn't translate to a 503 storm for every token-bearing
		// client. Default grace is 30s; operators can set it to 0 to
		// restore strict immediate invalidation. A vault that actively
		// revoked the token (key_deleted SSE) would have already flushed
		// this entry, so the grace path only fires when vault simply
		// isn't answering.
		if grace := s.cfg.Proxy.VaultStaleGrace; grace > 0 {
			s.tokenCacheMu.RLock()
			stale, ok := s.tokenCache[token]
			s.tokenCacheMu.RUnlock()
			if ok && time.Now().Before(stale.expiresAt.Add(grace)) {
				log.Printf("[token-auth] vault unreachable, serving stale token entry (client_id=%s, expired %v ago)",
					stale.clientID, time.Since(stale.expiresAt).Round(time.Millisecond))
				return stale, tokenLookupVaultStale
			}
		}
		return nil, tokenLookupVaultUnreachable
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return nil, tokenLookupNotRegistered
	}
	if resp.StatusCode != http.StatusOK {
		return nil, tokenLookupVaultError
	}

	var result struct {
		ID               string   `json:"id"`
		DefaultService   string   `json:"default_service"`
		DefaultModel     string   `json:"default_model"`
		FallbackServices []string `json:"fallback_services,omitempty"`
		AllowedServices  []string `json:"allowed_services,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, tokenLookupVaultError
	}

	entry := &tokenCacheEntry{
		clientID:         result.ID,
		service:          result.DefaultService,
		model:            result.DefaultModel,
		fallbackServices: result.FallbackServices,
		allowedServices:  result.AllowedServices,
		expiresAt:        time.Now().Add(5 * time.Second),
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
	return entry, tokenLookupOK
}

// tokenCacheMaxSize bounds the in-memory token→config cache. The periodic
// eviction goroutine keeps it roughly empty under normal load; this cap is an
// inline safety valve for bursty traffic.
const tokenCacheMaxSize = 200

// evictExpiredTokens removes tokenCache entries whose TTL has passed.
// Called both periodically (ticker) and opportunistically from lookupTokenConfig.
// VaultStaleGrace extends the keep-around window so that an entry stays
// available for emergency stale-fallback during vault unreachable spells.
// Once now is past expiresAt + grace, the entry is unconditionally dropped.
func (s *Server) evictExpiredTokens() {
	s.tokenCacheMu.Lock()
	defer s.tokenCacheMu.Unlock()
	now := time.Now()
	grace := s.cfg.Proxy.VaultStaleGrace
	for k, v := range s.tokenCache {
		if now.After(v.expiresAt.Add(grace)) {
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
	resp, err := internalHTTPClient(10 * time.Second).Do(req)
	if err != nil {
		log.Printf("[sync] 금고 연결 실패: %v", err)
		return
	}
	defer resp.Body.Close()

	var clients []struct {
		ID               string   `json:"id"`
		PreferredService string   `json:"preferred_service"`         // v0.2 canonical
		DefaultService   string   `json:"default_service"`           // v0.1 legacy
		DefaultModel     string   `json:"default_model"`             // v0.1 legacy
		ModelOverride    string   `json:"model_override,omitempty"`  // v0.2 canonical
		AgentType        string   `json:"agent_type"`
		Host             string   `json:"host"`
		WorkDir          string   `json:"work_dir,omitempty"`
		FallbackServices []string `json:"fallback_services,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return
	}
	// Refresh the service-level defaults *before* walking clients so the
	// per-client effMdl fallback (ModelOverride → DefaultModel →
	// serviceDefaults[svc]) can resolve to the live service default in the
	// same sync that just cleared a client's legacy default_model. Without
	// this ordering, clearing a client's model in vault would leave s.model
	// blank for one full sync cycle and OpenClaw would skip its config
	// rewrite (newMdl == oldMdl when both are blank).
	if err := s.syncAllowedServices(); err != nil {
		log.Printf("[sync] 서비스 목록 동기화 실패: %v", err)
	}
	// Build the co-hosted agent set for this proxy. The operator-assigned
	// Client.Host field drives inclusion — every client whose Host matches
	// os.Hostname() is reported in every heartbeat's ActiveClients, so one
	// host running N agents keeps N signal lights green. We trust the Host
	// field rather than probing per-agent-type liveness because reliable
	// detection varies (claude-code uses pgrep; VSCode extensions have no
	// matching binary; Windows-side clients are invisible from WSL pgrep).
	//
	// cfg.Proxy.ClaudeCodeClientID is kept as a compatibility override — if
	// set, the named client is force-included even when its Host field is
	// blank. Useful when os.Hostname() is unreliable (renamed boxes, WSL
	// before admin fills in Host values).
	host, _ := os.Hostname()
	var hosted []hostAgent
	seen := map[string]bool{}
	override := strings.TrimSpace(s.cfg.Proxy.ClaudeCodeClientID)
	for _, c := range clients {
		if c.ID == "" || c.ID == s.cfg.Proxy.ClientID {
			continue // proxy's own client_id is already covered by the main heartbeat
		}
		match := false
		if override != "" && c.ID == override {
			match = true
		} else if host != "" && c.Host == host {
			match = true
		}
		if !match || seen[c.ID] {
			continue
		}
		seen[c.ID] = true
		hosted = append(hosted, hostAgent{
			ClientID:  c.ID,
			AgentType: c.AgentType,
			Service:   c.DefaultService,
			Model:     c.DefaultModel,
		})
		// Seed agent-specific work_dir caches so SSE/sync-driven config
		// writers (updateEconoWorldOllamaThink, etc.) resolve the right
		// directory before the first /agent/apply call lands. Without this,
		// a fresh proxy boot on a host where the agent never bootstrapped
		// after restart silently no-ops on every config_change because
		// resolveEconoWorldDir falls back to the hard-coded default.
		if c.AgentType == "econoworld" && c.WorkDir != "" {
			setCachedEconoWorldDir(resolveEconoWorldDir(c.WorkDir))
		}
	}
	// ccID kept for any legacy reader that still references claudeCodeClientID.
	// Prefer the first claude-code entry in the hosted set.
	ccID := ""
	for _, a := range hosted {
		if a.AgentType == "claude-code" {
			ccID = a.ClientID
			break
		}
	}
	for _, c := range clients {
		if c.ID == s.cfg.Proxy.ClientID {
			s.mu.Lock()
			oldSvc, oldMdl := s.service, s.model
			// v0.2 canonical (PreferredService / ModelOverride) wins; v0.1
			// legacy (DefaultService / DefaultModel) is the fallback so
			// pre-migration installs keep working. A v0.2-only client (which
			// leaves the legacy fields empty) used to drop through silently
			// and freeze s.service / s.model at whatever value was set last,
			// causing dispatch to keep routing to the previous service even
			// after the operator changed the dashboard.
			effSvc := c.PreferredService
			if effSvc == "" {
				effSvc = c.DefaultService
			}
			if effSvc != "" {
				s.service = effSvc
			}
			effMdl := c.ModelOverride
			if effMdl == "" {
				effMdl = c.DefaultModel
			}
			if effMdl == "" {
				effMdl = s.serviceDefaults[effSvc]
			}
			if effMdl != "" {
				s.model = effMdl
			}
			s.ownFallback = append(s.ownFallback[:0], c.FallbackServices...)
			s.ownModelOverride = c.ModelOverride
			newSvc, newMdl := s.service, s.model
			s.mu.Unlock()
			log.Printf("[sync] 설정 로드: %s/%s  override=%q  fallback=%v",
				effSvc, effMdl, c.ModelOverride, c.FallbackServices)
			if newSvc != oldSvc || newMdl != oldMdl {
				go updateOpenClawJSON(newSvc, newMdl, s.defaultProxyOrigin())
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
	prev := len(s.hostAgents)
	s.claudeCodeClientID = ccID
	s.ownAgentType = ownType
	s.hostAgents = hosted
	s.mu.Unlock()
	if prev != len(hosted) {
		ids := make([]string, 0, len(hosted))
		for _, a := range hosted {
			ids = append(ids, a.ClientID)
		}
		log.Printf("[sync] host agents (host=%q): %v", host, ids)
	}

	// sync keys
	if err := s.keyMgr.SyncFromVault(); err != nil {
		log.Printf("[sync] 키 동기화 실패: %v", err)
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
	client := internalHTTPClient(5 * time.Second)
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
	oldOllamaThink := s.serviceReasoning["ollama"]
	s.allowedServices = ids
	s.serviceURLs = urls
	s.serviceDefaults = defaults
	s.serviceReasoning = reasoning
	newOllamaThink := reasoning["ollama"]
	s.mu.Unlock()
	// Mirror an ollama reasoning_mode change into EconoWorld's ai_config.json.
	// The SSE config_change handler only fires for client-targeted edits
	// (model / preferred_service); a service-level reasoning toggle in the
	// vault dashboard reaches us through this poll instead. Without the
	// diff-and-call here, EconoWorld's ollama.think would only refresh on
	// the next /agent/apply or model swap.
	if oldOllamaThink != newOllamaThink {
		go updateEconoWorldOllamaThink(newOllamaThink)
	}
	// Re-merge every local-service URL list so a vault-side change (operator
	// added or removed an instance, edited a comma-separated list in the
	// dashboard URL field) flows into the dispatch pools without restart.
	s.refreshAllServicePoolURLs()
	log.Printf("[sync] 프록시 서비스 목록: %v (urls: %v, defaults: %v, reasoning: %v)", ids, urls, defaults, reasoning)
	return nil
}

// resolveOllamaURLs builds the multi-instance URL list for ollamaPool.
// Sources, in order, with the first entry winning on duplicates:
//
//  1. cfg.Proxy.OllamaURLs (yaml + WV_OLLAMA_URLS env list)
//  2. WV_OLLAMA_URL env (legacy single — accepts comma list too)
//  3. OLLAMA_URL env (legacy single — accepts comma list too)
//  4. vault serviceURLs["ollama"] (single string from SSE — accepts comma list)
//
// Every textual source goes through splitOllamaURLs so an operator who types
// "http://a:11434, http://b:11434" into the vault dashboard's URL field gets
// the same multi-instance behaviour as the dedicated WV_OLLAMA_URLS env. The
// vault schema still carries one string field; this is a UX-only relaxation.
//
// Empty result is acceptable — the pool falls back to localhost on dispatch.
// Called at NewServer and again after every vault sync so a vault-side URL
// change propagates without restart.
func (s *Server) resolveOllamaURLs() []string {
	out := make([]string, 0, len(s.cfg.Proxy.OllamaURLs)+3)
	out = append(out, s.cfg.Proxy.OllamaURLs...)
	out = append(out, splitOllamaURLs(os.Getenv("WV_OLLAMA_URL"))...)
	out = append(out, splitOllamaURLs(os.Getenv("OLLAMA_URL"))...)
	s.mu.RLock()
	u := s.serviceURLs["ollama"]
	s.mu.RUnlock()
	out = append(out, splitOllamaURLs(u)...)
	return out
}

// splitOllamaURLs accepts a string that may contain a single URL or a
// comma-separated list. Empty / whitespace-only entries are dropped so a
// trailing comma or stray space doesn't poison the resulting list.
func splitOllamaURLs(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ollamaURL returns the first URL in the ollama dispatch pool. Kept as a
// thin shim for callers that need a single representative URL (e.g. log
// lines), preserving the legacy resolution order for single-URL deployments.
// Multi-instance dispatch goes through ollamaURLForModel.
func (s *Server) ollamaURL() string {
	s.poolsMu.RLock()
	pool := s.servicePools["ollama"]
	s.poolsMu.RUnlock()
	if pool != nil {
		urls := pool.URLs()
		if len(urls) > 0 {
			return urls[0]
		}
	}
	if v := os.Getenv("WV_OLLAMA_URL"); v != "" {
		return v
	}
	if v := os.Getenv("OLLAMA_URL"); v != "" {
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

// ollamaURLForModel returns the Ollama URL hosting the given model id, or
// the first configured URL on miss (so the upstream returns a real 404).
// Single-URL deployments resolve every model to that one URL.
func (s *Server) ollamaURLForModel(model string) string {
	s.poolsMu.RLock()
	pool := s.servicePools["ollama"]
	s.poolsMu.RUnlock()
	if pool != nil {
		if u := pool.URLForModel(model); u != "" {
			return u
		}
	}
	return s.ollamaURL()
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
	// IPAllowlist runs ahead of every other gate so an off-LAN scanner can't
	// even touch the auth surface when the operator has scoped the proxy down
	// to a specific CIDR set. Empty list (default) is a no-op and keeps the
	// historic "trust the LAN" behaviour.
	allow := middleware.IPAllowlist(s.cfg.Proxy.AllowCIDRs)
	return middleware.Chain(mux,
		middleware.Recovery,
		allow,
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
	// Unauthenticated /health gets status + readiness + version. clientID and
	// sse_connected used to be public, which let an off-LAN scanner finger-
	// print the deploy topology (which client this proxy serves, whether
	// vault SSE is up). Those are now hidden behind auth. Version stays
	// public because deploy verification (Makefile post-deploy step) reads it
	// without a token, and the version string is already published in
	// GitHub releases — gating it would break ops without raising the bar
	// for a real attacker.
	if !s.callerHasValidToken(r) {
		jsonOK(w, map[string]interface{}{
			"status":    "ok",
			"readiness": readiness,
			"version":   Version,
		})
		return
	}
	jsonOK(w, map[string]interface{}{
		"status":        "ok",
		"readiness":     readiness,
		"version":       Version,
		"client":        s.cfg.Proxy.ClientID,
		"sse_connected": sseConnected,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Unauthenticated /status used to expose clientID + service + model +
	// tool_filter + active services list to anyone who could reach the
	// listener. Off-LAN that's a topology fingerprinting freebie. Now those
	// fields require a token; unauthenticated callers get status + version
	// only (version is already public via GitHub releases, and Makefile
	// post-deploy verification reads it without auth).
	if !s.callerHasValidToken(r) {
		jsonOK(w, map[string]interface{}{
			"status":  "ok",
			"version": Version,
		})
		return
	}

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
	// applied to their own requests.
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
	if !s.requireProxyToken(w, r) {
		return
	}
	svc := r.URL.Query().Get("service")
	all := s.registry.All(svc)
	jsonOK(w, map[string]interface{}{
		"models": all,
		"count":  len(all),
	})
}

// callerHasValidToken reports whether the request carries a Bearer token
// that the proxy would accept for an authenticated route. Unlike
// requireProxyToken it never writes a response — it's used by /health and
// /status to decide whether to surface the rich payload (deploy fingerprint)
// or the minimal liveness one. Cheap because lookupTokenConfig is cached.
func (s *Server) callerHasValidToken(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == "" {
		return false
	}
	if s.cfg.Proxy.VaultToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.Proxy.VaultToken)) == 1 {
		return true
	}
	if entry := s.lookupTokenConfig(token); entry != nil {
		return true
	}
	return false
}

// requireProxyToken authenticates a request to /v1/chat/completions,
// /v1/models, /google/*, and /api/models. Accepted tokens: the proxy's own
// VaultToken, or any token that the vault recognises as a registered client
// (resolved via lookupTokenConfig with a 5s cache). Writes 401 and returns
// false on rejection.
func (s *Server) requireProxyToken(w http.ResponseWriter, r *http.Request) bool {
	s.substituteSelfManagedSentinel(r)
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		jsonError(w, "Authorization header required", http.StatusUnauthorized)
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == "" {
		jsonError(w, "Authorization header required", http.StatusUnauthorized)
		return false
	}
	if s.cfg.Proxy.VaultToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.Proxy.VaultToken)) == 1 {
		return true
	}
	entry, reason := s.lookupTokenConfigDetailed(token)
	if entry != nil {
		return true
	}
	// Temporary diagnostic for the <internal incident> 401 storm
	// where the operator's openclaw.json apiKey looked correct but
	// token-auth still rejected. Logs lengths + 4-char prefix of the
	// incoming token vs. the proxy's own VaultToken so the divergence
	// is visible without writing the full secret to the journal.
	log.Printf("[token-auth] reject reason=%d recv=%s+%d vault=%s+%d",
		reason, safePrefix(token, 4), len(token),
		safePrefix(s.cfg.Proxy.VaultToken, 4), len(s.cfg.Proxy.VaultToken))
	tokenAuthFail(w, reason, s.cfg.Proxy.VaultToken != "")
	return false
}

// safePrefix returns up to n characters of s, never panicking on short input.
func safePrefix(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// requireAnthropicToken authenticates a request to /v1/messages. Accepts the
// same tokens as requireProxyToken plus (a) Anthropic-native `x-api-key`
// header and (b) BYO `sk-ant-*` credentials (Claude Code OAuth via NanoClaw's
// credential proxy) — those are forwarded upstream by the handler and so are
// trusted as long as they are present.
func (s *Server) requireAnthropicToken(w http.ResponseWriter, r *http.Request) bool {
	s.substituteSelfManagedSentinel(r)
	var token string
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		token = strings.TrimPrefix(auth, "Bearer ")
	} else if xKey := r.Header.Get("x-api-key"); xKey != "" {
		token = xKey
	}
	if token == "" {
		jsonError(w, "Authorization header or x-api-key required", http.StatusUnauthorized)
		return false
	}
	// BYO Anthropic OAuth credentials — pass through to upstream as-is.
	if strings.HasPrefix(token, "sk-ant-") {
		return true
	}
	if s.cfg.Proxy.VaultToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.Proxy.VaultToken)) == 1 {
		return true
	}
	entry, reason := s.lookupTokenConfigDetailed(token)
	if entry != nil {
		return true
	}
	tokenAuthFail(w, reason, s.cfg.Proxy.VaultToken != "")
	return false
}

// tokenAuthFail emits a 401 (or 503 when the cause is "vault unreachable")
// with a message that names the actual reason instead of the historic
// catch-all "invalid token". Operators triaging a 401 can immediately tell
// whether the token is unknown to vault, vault itself is down, or the
// proxy's own vault_token is what mismatched.
func tokenAuthFail(w http.ResponseWriter, reason tokenLookupReason, proxyTokenConfigured bool) {
	switch reason {
	case tokenLookupNotRegistered:
		jsonError(w, "token not registered with vault — register the client and reuse its token", http.StatusUnauthorized)
	case tokenLookupVaultUnreachable:
		jsonError(w, "vault unreachable — check proxy.vault_url and that the vault listener is up", http.StatusServiceUnavailable)
	case tokenLookupVaultError:
		jsonError(w, "vault returned an unexpected response — check vault logs", http.StatusBadGateway)
	case tokenLookupSkipped:
		// Either vault_url is unset or the caller used the proxy's own token
		// and that token mismatched (we never reached vault). Differentiate
		// so operators don't waste time inspecting vault when the issue is
		// actually proxy.vault_token.
		if proxyTokenConfigured {
			jsonError(w, "proxy.vault_token mismatch (no vault fallback configured)", http.StatusUnauthorized)
		} else {
			jsonError(w, "no auth configured (proxy.vault_token unset and proxy.vault_url empty)", http.StatusServiceUnavailable)
		}
	default:
		jsonError(w, "invalid token", http.StatusUnauthorized)
	}
}

// requireAdminToken authenticates a request to a privileged proxy endpoint
// (model override, reload, think-mode toggle). Only the proxy's own VaultToken
// passes — client tokens are explicitly rejected because these endpoints mutate
// the proxy's own routing state, which a per-client token has no authority over.
// Fail-closed when no VaultToken is configured.
func (s *Server) requireAdminToken(w http.ResponseWriter, r *http.Request) bool {
	if s.cfg.Proxy.VaultToken == "" {
		jsonError(w, "admin endpoints disabled (proxy.vault_token not configured)", http.StatusServiceUnavailable)
		return false
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		jsonError(w, "Authorization header required", http.StatusUnauthorized)
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.Proxy.VaultToken)) != 1 {
		jsonError(w, "invalid admin token", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *Server) handleConfigModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		jsonError(w, "PUT required", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAdminToken(w, r) {
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
		go updateOpenClawJSON(newSvc, newMdl, s.defaultProxyOrigin())
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
	client := internalHTTPClient(5 * time.Second)
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
	if !s.requireAdminToken(w, r) {
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdminToken(w, r) {
		return
	}
	go s.syncFromVault()
	jsonOK(w, map[string]string{"status": "reloading"})
}

// ─── Gemini API Handler (non-streaming) ──────────────────────────────────────

func (s *Server) handleGemini(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireProxyToken(w, r) {
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
			// Promote to google by default, but a local quantized variant
			// (GGUF file, Unsloth Dynamic, q*/iq*/bf16/MXFP markers) belongs
			// on llamacpp — Google's API rejects those names with HTTP 400
			// "unexpected model name format".
			if isLocalQuantizedModel(urlModel) {
				mdl = urlModel
			} else {
				svc = "google"
				mdl = urlModel
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")

	// Native Gemini path: strict-by-default. Add per-token fallback chain
	// support here too if a use case emerges (currently this path does not
	// look up Bearer token, so it always runs strict).
	result, err := s.dispatch(r.Context(), svc, mdl, &req)
	if err != nil {
		log.Printf("[proxy] 오류: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(GeminiResponse{
			Error: &GeminiError{Code: 502, Message: err.Error()},
		})
		return
	}
	if result.UsedService != "" {
		w.Header().Set("X-WV-Used-Service", result.UsedService)
	}
	if result.UsedModel != "" {
		w.Header().Set("X-WV-Used-Model", result.UsedModel)
	}
	if result.FallbackReason != "" {
		w.Header().Set("X-WV-Fallback-Reason", result.FallbackReason)
	}
	json.NewEncoder(w).Encode(result.Response)
}

// ─── OpenAI API Handler ───────────────────────────────────────────────────────

func (s *Server) handleOpenAI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireProxyToken(w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAIBodySize)

	var oaiReq OpenAIRequest
	if err := json.NewDecoder(r.Body).Decode(&oaiReq); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	s.filter.FilterOpenAI(&oaiReq)

	// Stream early-flush: when the caller asked for stream=true we commit
	// status 200 + SSE headers and emit a valid empty-delta `data:` chunk
	// *before* starting the upstream LLM call. Without this, dispatch can
	// take 60-180s on a cold/large local model (qwen3.6:27b prompt eval)
	// and the caller's first-byte timeout fires before any data is
	// written.
	//
	// v0.2.49 used a `: warming up` SSE comment frame here; raw fetch
	// accepted it but the OpenAI Node SDK that OpenClaw embeds rejected
	// the stream as "Connection error" ~14s in (<internal incident>
	// gateway.err.log: 4× retry / lane durationMs ≈ 75s). The SDK seems
	// to require the first frame to be a parsable `data:` chunk before it
	// considers the stream live. Keepalive frames are the same — empty
	// `delta:{}` with finish_reason:null, treated as no-op by every
	// well-behaved SSE consumer (delta merge produces no content) but
	// counts as a real frame for the SDK's stream-start / idle counters.
	//
	// Interval is 8s (was 15s) — still well below typical 60s+ idle
	// timeouts but short enough that even a 14s SDK quirk gets two
	// frames before tripping. The actual response chunks are produced
	// below the dispatch call. On dispatch error we surface the error
	// as an SSE chunk + [DONE] (status is already committed).
	var (
		keepaliveStop chan struct{}
		keepaliveDone chan struct{}
		stopOnce      sync.Once
	)
	stopKeepalive := func() {
		stopOnce.Do(func() {
			if keepaliveStop != nil {
				close(keepaliveStop)
			}
			if keepaliveDone != nil {
				<-keepaliveDone
			}
		})
	}
	const keepaliveFrame = `data: {"id":"chatcmpl-warmup","object":"chat.completion.chunk","model":"warmup","choices":[{"index":0,"delta":{},"finish_reason":null}]}` + "\n\n"
	if oaiReq.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			fmt.Fprint(w, keepaliveFrame)
			f.Flush()
			keepaliveStop = make(chan struct{})
			keepaliveDone = make(chan struct{})
			go func(stop <-chan struct{}, done chan<- struct{}) {
				defer close(done)
				ticker := time.NewTicker(8 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-stop:
						return
					case <-r.Context().Done():
						return
					case <-ticker.C:
						fmt.Fprint(w, keepaliveFrame)
						f.Flush()
					}
				}
			}(keepaliveStop, keepaliveDone)
			defer stopKeepalive()
		}
	}

	// Resolution order for (service, model) — most-specific source wins:
	//   1. Vault-side token override (entry.service / entry.model)
	//      → operator's enforcement on a specific client. Final.
	//   2. Request body's explicit model
	//      → caller's stated intent. Honoured when vault has no override.
	//   3. Proxy's own default (s.service / s.model)
	//      → fallback for unauthenticated callers / empty body.
	//
	// Pre-v0.2.27 the proxy's default was layered ABOVE the request body, so
	// a token-auth'd client whose model_override is empty would get the
	// proxy's own model regardless of what they typed in the request — the
	// pattern that hit earlier reviewer feedback, where an econoworld token route
	// to host-D's proxy (s.model="anthropic/claude-opus-4-7") swallowed the
	// requested "qwen3.6:27b".
	mdl := oaiReq.Model
	s.mu.RLock()
	svc := s.service
	proxyDefaultMdl := s.model
	ownFb := append([]string(nil), s.ownFallback...)
	ownOverride := s.ownModelOverride
	s.mu.RUnlock()

	var resolvedClientID string
	var fallbackChain []string
	tokenResolved := false
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken := strings.TrimPrefix(authHeader, "Bearer ")
		if entry := s.lookupTokenConfig(reqToken); entry != nil {
			tokenResolved = true
			resolvedClientID = entry.clientID
			if entry.service != "" {
				svc = entry.service
			}
			if entry.model != "" {
				mdl = entry.model
			} else if svc != "" {
				// Vault explicitly cleared the client's model override (or
				// the operator picked "(서비스 기본 사용)" in dashboard).
				// Honour that intent by forcing the service-level default —
				// the alternative ("request body's model wins") silently
				// bypasses the dashboard knob. Surfaced when an OAI-compat
				// client kept a stale path-style model id in its
				// ai_config.json even after the operator switched the LM
				// Studio service default to a different model.
				s.mu.RLock()
				if def := s.serviceDefaults[svc]; def != "" {
					mdl = def
				}
				s.mu.RUnlock()
			}
			fallbackChain = entry.fallbackServices
		}
	}
	// Local-process callers (OpenClaw, Claude Code) usually call without an
	// Authorization header. Inherit the proxy's own client config so vault
	// stays the single source of truth for routing — both fallback chain
	// and (when set) the operator's explicit model_override.
	if !tokenResolved {
		if ownOverride != "" {
			mdl = ownOverride
		}
		if fallbackChain == nil {
			fallbackChain = ownFb
		}
	}
	// No request-body model and no vault override → fall back to proxy default.
	if mdl == "" {
		mdl = proxyDefaultMdl
	}
	// Refresh lastSeen after response completes so long-running requests
	// (streaming AI responses) keep the client visible on the dashboard.
	defer s.refreshClientAct(resolvedClientID)

	// OpenClaw sends models as "provider/model-id" (e.g. "wall-vault/gemini-2.5-flash",
	// "anthropic/claude-opus-4-6"). Parse and route accordingly.
	svc, mdl = parseProviderModel(svc, mdl)

	// OAI-compat stream passthrough — gated. When enabled, callers that
	// asked for stream:true and resolved to one of the OAI-compat services
	// pipe backend SSE chunks straight through. Stream-mode callers do
	// NOT consult the fallback chain (see spec §3.3).
	if oaiReq.Stream && s.cfg.Proxy.OAIStreamForward {
		if _, ok := oaiCompatServices[svc]; ok {
			// SSE headers were already committed by the early-flush block
			// above (line ~1293-1295). streamLocalService's internal
			// idempotence guard handles the case where this gate is
			// reached without that commit. Connection: keep-alive remains
			// intentionally unset — the early-flush block does not set it
			// and adding it post-WriteHeader is dead.
			flusher, _ := w.(http.Flusher)
			// The keepalive goroutine started above for the buffered path is
			// irrelevant here — the real backend stream provides natural
			// keepalives. stopKeepalive is always defined as a closure and
			// guards keepaliveStop internally, so direct call is safe.
			stopKeepalive()
			if err := s.streamLocalService(r.Context(), w, flusher, svc, mdl, &oaiReq); err != nil {
				writeOpenAIErrorChunk(w, flusher, err)
			}
			return
		}
	}

	geminiReq := OpenAIToGemini(&oaiReq)
	dispatchRes, err := s.dispatchWithChain(r.Context(), svc, mdl, fallbackChain, geminiReq)
	if err != nil {
		// When stream early-flush already committed status 200, we cannot
		// switch to a JSON 502 response — emit the error as a final SSE
		// chunk + [DONE] so the caller's stream parser can surface it.
		if keepaliveStop != nil {
			stopKeepalive()
			if f, ok := w.(http.Flusher); ok {
				errChunk := map[string]interface{}{
					"id":      "chatcmpl-proxy",
					"object":  "chat.completion.chunk",
					"model":   mdl,
					"choices": []map[string]interface{}{{"index": 0, "delta": map[string]string{"content": "[wall-vault: " + err.Error() + "]"}, "finish_reason": "stop"}},
				}
				if b, mErr := json.Marshal(errChunk); mErr == nil {
					fmt.Fprintf(w, "data: %s\n\n", b)
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
				f.Flush()
			}
			return
		}
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	// Stop keepalive before writing real response chunks to avoid racy
	// concurrent Write/Flush against the same ResponseWriter.
	stopKeepalive()
	geminiResp := dispatchRes.Response
	// Reflect the actual service/model used — may differ from the requested
	// pair when dispatch fell back to another provider (see DispatchResult).
	if dispatchRes.UsedModel != "" {
		mdl = dispatchRes.UsedModel
	}
	if dispatchRes.UsedService != "" {
		svc = dispatchRes.UsedService
	}
	// Surface routing decisions to the caller via response headers so downstream
	// agents (<separate incident> etc.) can react to silent provider switches without
	// having to compare the response body's model field against the request.
	w.Header().Set("X-WV-Used-Service", svc)
	w.Header().Set("X-WV-Used-Model", mdl)
	if dispatchRes.FallbackReason != "" {
		w.Header().Set("X-WV-Fallback-Reason", dispatchRes.FallbackReason)
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
	if !s.requireAnthropicToken(w, r) {
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

	// Same resolution priority as handleOpenAI:
	//   1. Vault-side token override (final)
	//   2. Request body's model (caller intent)
	//   3. Proxy's default (s.model)
	mdl := req.Model
	s.mu.RLock()
	svc := s.service
	proxyDefaultMdl := s.model
	allowedServices := s.allowedServices
	ownFb := append([]string(nil), s.ownFallback...)
	ownOverride := s.ownModelOverride
	s.mu.RUnlock()

	// Token-based model override: same logic as handleOpenAI.
	// Anthropic API uses x-api-key header instead of Authorization: Bearer,
	// so check both to support Claude Code and other Anthropic-format clients.
	var resolvedClientID string
	var fallbackChain []string
	tokenResolved := false
	reqToken := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		reqToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else if xKey := r.Header.Get("x-api-key"); xKey != "" {
		reqToken = xKey
	}
	if reqToken != "" {
		if entry := s.lookupTokenConfig(reqToken); entry != nil {
			tokenResolved = true
			resolvedClientID = entry.clientID
			if entry.service != "" {
				svc = entry.service
			}
			if entry.model != "" {
				mdl = entry.model
			} else if svc != "" {
				// Vault explicitly cleared the client's model override (or
				// the operator picked "(서비스 기본 사용)" in dashboard).
				// Honour that intent by forcing the service-level default —
				// the alternative ("request body's model wins") silently
				// bypasses the dashboard knob. Surfaced when an OAI-compat
				// client kept a stale path-style model id in its
				// ai_config.json even after the operator switched the LM
				// Studio service default to a different model.
				s.mu.RLock()
				if def := s.serviceDefaults[svc]; def != "" {
					mdl = def
				}
				s.mu.RUnlock()
			}
			fallbackChain = entry.fallbackServices
		}
	}
	if !tokenResolved {
		if ownOverride != "" {
			mdl = ownOverride
		}
		if fallbackChain == nil {
			fallbackChain = ownFb
		}
	}
	if mdl == "" {
		mdl = proxyDefaultMdl
	}
	defer s.refreshClientAct(resolvedClientID)

	// BYO Anthropic credentials — if the caller supplied an `sk-ant-*` token
	// (Claude Code OAuth via NanoClaw's credential-proxy is the primary case),
	// forward it through instead of letting the passthrough clobber it with
	// a vault x-api-key. Vault-issued client tokens use a different prefix
	// so there's no overlap.
	byoAPIKey := ""
	byoBearer := ""
	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		tok := strings.TrimPrefix(authHeader, "Bearer ")
		if strings.HasPrefix(tok, "sk-ant-") {
			byoBearer = tok
		}
	}
	if xKey := r.Header.Get("x-api-key"); strings.HasPrefix(xKey, "sk-ant-") {
		byoAPIKey = xKey
	}

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
	// Streaming bridge: Claude Code (agent SDK) opens with `stream: true`
	// but wall-vault currently buffers upstream via io.ReadAll. Returning a
	// single JSON blob for a stream-opened session makes the SDK hang until
	// its 30-minute ceiling. Force upstream to non-streaming, capture the
	// full JSON, then replay it as a proper Anthropic SSE event stream so
	// the SDK sees a complete turn arrive in one fast burst.
	originalStream := req.Stream
	req.Stream = false

	if usePassthrough {
		if body, contentType, _, err := s.callAnthropicPassthrough(r.Context(), &req, mdl, byoAPIKey, byoBearer); err == nil {
			if originalStream {
				WriteAnthropicSSEFromJSON(w, body)
				return
			}
			if contentType == "" {
				contentType = "application/json"
			}
			w.Header().Set("Content-Type", contentType)
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
	dispatchRes, err := s.dispatchWithChain(r.Context(), svc, mdl, fallbackChain, geminiReq)
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
	if dispatchRes.UsedService != "" {
		w.Header().Set("X-WV-Used-Service", dispatchRes.UsedService)
	}
	w.Header().Set("X-WV-Used-Model", mdl)
	if dispatchRes.FallbackReason != "" {
		w.Header().Set("X-WV-Fallback-Reason", dispatchRes.FallbackReason)
	}

	resp := GeminiRespToAnthropic(mdl, geminiResp)
	if originalStream {
		body, _ := json.Marshal(resp)
		WriteAnthropicSSEFromJSON(w, body)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── OpenAI-compatible model list (/v1/models) ────────────────────────────────

func (s *Server) handleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	if !s.requireProxyToken(w, r) {
		return
	}
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
//
// When the primary service produced the response, FallbackReason is empty and
// UsedService == requested primary. When dispatch fell over to a service later
// in the fallback chain, FallbackReason is the human-readable error from the
// primary attempt — rendered into the X-WV-Fallback-Reason response header so
// callers can detect and react to silent provider switches.
type DispatchResult struct {
	Response       *GeminiResponse
	UsedService    string
	UsedModel      string
	FallbackReason string
}

// dispatchInternal executes the (primary, fallbackChain) pair against the
// configured backends and returns a DispatchResult capturing the actual
// service/model that produced the response. Strict-by-default: an empty
// fallbackChain means "primary or 502" — no silent substitution to a different
// model on a different provider. When fallbackChain is non-empty, services in
// it are tried in order after primary fails, with the primary error captured
// as FallbackReason on whichever attempt succeeds.
//
// The previous behaviour (auto-fallback through s.cfg.Proxy.Services with the
// destination service's default_model) was removed in v0.2.27 — it surfaced as
// <earlier incident> where a qwen3.6:27b request returned google/gemini-flash
// because Ollama was unreachable and dispatch silently swapped the model.
// Now the operator sets FallbackServices explicitly per client when fallback
// is desired; default is strict.
func (s *Server) dispatch(ctx context.Context, service, model string, req *GeminiRequest) (*DispatchResult, error) {
	return s.dispatchWithChain(ctx, service, model, nil, req)
}

func (s *Server) dispatchWithChain(ctx context.Context, primary, model string, fallbackChain []string, req *GeminiRequest) (*DispatchResult, error) {
	s.mu.RLock()
	serviceDefaults := s.serviceDefaults
	s.mu.RUnlock()

	// Try order: primary first, then fallback chain (skipping the primary if
	// it appears in the chain). Empty fallback chain → strict primary-only.
	tryOrder := []string{primary}
	seen := map[string]bool{primary: true}
	for _, svc := range fallbackChain {
		if svc == "" || seen[svc] {
			continue
		}
		tryOrder = append(tryOrder, svc)
		seen[svc] = true
	}

	var primaryErr error
	for i, svc := range tryOrder {
		isPrimary := i == 0
		// Fast-skip cloud services whose keys are all on cooldown or exhausted
		// — prevents dispatch from spending seconds on retries that will re-hit
		// 429/402. Local services and plugin-defined backends with non-key auth
		// (bearer to vault token / none) skip the keyMgr gate via
		// serviceNeedsKey.
		if s.serviceNeedsKey(svc) && !s.keyMgr.CanServe(svc) {
			log.Printf("[proxy] %s skip: 전체 쿨다운 또는 등록된 키 없음", svc)
			err := fmt.Errorf("서비스 '%s' 전체 쿨다운", svc)
			if isPrimary {
				primaryErr = err
			}
			continue
		}

		// Primary uses the caller's requested model; fallback uses the
		// destination service's default_model (different namespaces).
		targetModel := model
		if !isPrimary {
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
			// Plugin-defined service: route by the plugin's declared
			// request_format. The empty default is treated as openai-compat
			// because that's the shape >90 % of community plugins ship
			// (LM Studio / vLLM / llama.cpp / Jan / Kobold / TabbyAPI / …).
			// A drop-in plugin file is therefore enough to teach dispatch
			// about a new backend — no Go-side switch edit required.
			if plugin := s.pluginByID[svc]; plugin != nil {
				switch plugin.RequestFormat {
				case "", "openai":
					resp, err = s.callLocalService(ctx, svc, targetModel, req)
				default:
					err = fmt.Errorf("plugin %q: dispatch only supports request_format 'openai' (got %q)", svc, plugin.RequestFormat)
				}
			} else {
				err = fmt.Errorf("unknown service: %s (no plugin registered for this id)", svc)
			}
		}
		if err == nil {
			result := &DispatchResult{Response: resp, UsedService: svc, UsedModel: targetModel}
			if !isPrimary {
				log.Printf("[proxy] fallback: %s/%s → %s/%s (primary err: %v)", primary, model, svc, targetModel, primaryErr)
				if primaryErr != nil {
					result.FallbackReason = primaryErr.Error()
				} else {
					result.FallbackReason = "primary failed"
				}
				s.hooksMgr.Fire(hooks.EventModelChanged, map[string]string{
					"service": svc,
					"model":   targetModel,
				})
			}
			if dispatchTraceEnabled() {
				reason := "primary"
				if !isPrimary {
					reason = "fallback"
				}
				log.Printf("[dispatch] requested=%s/%s resolved=%s/%s reason=%s",
					primary, model, svc, targetModel, reason)
			}
			return result, nil
		}
		log.Printf("[proxy] %s (%s) failed: %v", svc, targetModel, err)
		if isPrimary {
			primaryErr = err
		}
	}
	// Build aggregate error preserving the primary failure as the headline so
	// callers can render "primary X failed: <reason>" without surfacing a
	// possibly-misleading last-fallback error.
	lastErr := primaryErr
	if lastErr == nil {
		lastErr = fmt.Errorf("dispatch: 모든 서비스 실패")
	}
	if lastErr != nil {
		s.hooksMgr.Fire(hooks.EventServiceDown, map[string]string{
			"error": lastErr.Error(),
		})
	}
	return nil, fmt.Errorf("모든 서비스 실패: %v", lastErr)
}

// promptTokenCapDefault is the dispatch-time auto-truncate threshold for
// OAI-compat backend calls. Default is 0 (disabled): the operator made
// the call (v0.2.81) that the agent runtime, not the proxy, owns
// conversation pruning. Set WV_PROMPT_TOKEN_CAP=<positive int> to opt
// back in per-host — useful as a temporary safety net while a
// misbehaving agent gets a /compact patch.
//
// Why the default flipped: v0.2.80 trimmed by message granularity, which
// can split assistant tool_call / tool result pairs and trigger a
// different 400 from the backend ("orphaned tool message"). Until the
// trimmer is taught to drop pair-by-pair, the safer default is "do not
// touch the prompt and let the agent compact its own history".
const promptTokenCapDefault = 0

func promptTokenCap() int {
	if v := os.Getenv("WV_PROMPT_TOKEN_CAP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return promptTokenCapDefault
}

// estimateOAIRequestTokens approximates the prompt size an OAI-compat
// backend will see. Counts message content chars plus the marshalled
// tools schema (claude-code agents send 16+ tool defs that dominate
// prompt size, so omitting them produces wildly low estimates). Divides
// by 3 — empirical compromise between English (~4 chars/token) and
// Korean (~2 chars/token) for the wall-vault fleet's mixed traffic.
func estimateOAIRequestTokens(req *OpenAIRequest) int {
	if req == nil {
		return 0
	}
	chars := 0
	for _, m := range req.Messages {
		chars += len(m.Content)
	}
	if len(req.Tools) > 0 {
		if data, err := json.Marshal(req.Tools); err == nil {
			chars += len(data)
		}
	}
	return chars / 3
}

// truncateOldestOAIMessages drops the oldest non-system, non-final
// messages from req until the estimated token count fits under cap.
// Preserves the system prompt at index 0 (when present) and the final
// user message (current turn). Returns the number of messages dropped
// so callers can log when truncation kicked in.
//
// cap ≤ 0 disables the truncation (returns 0).
func truncateOldestOAIMessages(req *OpenAIRequest, cap int) int {
	if req == nil || cap <= 0 || len(req.Messages) == 0 {
		return 0
	}
	dropped := 0
	for estimateOAIRequestTokens(req) > cap {
		dropIdx := -1
		for i, m := range req.Messages {
			if i == len(req.Messages)-1 {
				break // never drop the final (current) message
			}
			if m.Role == "system" {
				continue
			}
			dropIdx = i
			break
		}
		if dropIdx < 0 {
			return dropped // nothing droppable left
		}
		req.Messages = append(req.Messages[:dropIdx], req.Messages[dropIdx+1:]...)
		dropped++
	}
	return dropped
}

// translateBackendError converts a non-2xx response from a local OAI-compat
// backend (llamacpp / ollama / lmstudio / vLLM) into an error whose message
// is meaningful to the end-user agent. Today we only special-case
// llama.cpp's `exceed_context_size_error` because that's the failure mode
// the fleet hits in practice — agents accumulate conversation history and
// blow past the per-server `-c` budget, the raw HTTP 400 surfaces as an
// opaque blob in the bot's UI, and the operator has no clue why their
// chat went silent. For other 4xx / 5xx the original HTTP status + body
// is preserved so debugging info isn't lost.
//
// statusCode is the HTTP status from the upstream backend. body is the
// raw response body (already read by the caller — the dispatch loop and
// streamLocalService use a small LimitReader so this is bounded).
func translateBackendError(serviceID string, statusCode int, body []byte) error {
	bodyStr := string(body)
	if statusCode == http.StatusBadRequest && strings.Contains(bodyStr, "exceed_context_size_error") {
		return fmt.Errorf(
			"%s: 대화 컨텍스트가 너무 깁니다 — 백엔드의 처리 한계를 초과했습니다. "+
				"대화를 압축하거나(/compact) 새 conversation 으로 시작하세요. "+
				"백엔드 원본: %s",
			serviceID, bodyStr,
		)
	}
	return fmt.Errorf("%s 오류: HTTP %d: %s", serviceID, statusCode, bodyStr)
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
	ollamaURL := s.ollamaURLForModel(model)
	// "<model>@local<N>" pins the request to a specific instance; the alias
	// is consumed by URLForModel above. Strip it before forwarding so the
	// upstream Ollama receives a plain model id.
	model = stripInstanceAlias(model)

	// Distribute fleet arrival times: each proxy delays entry by a
	// deterministic phase offset (sha256(client_id) → [0, 500ms)) plus a
	// small uniform jitter ([0, 200ms)). Prevents simultaneous fan-in when
	// multiple proxies reach a shared local backend in lock-step. Restored
	// in v0.2.27 after the v0.2.23 removal — see timing.go for context.
	if d := AgentOffset(s.cfg.Proxy.ClientID, localAgentOffsetMs) +
		FallbackJitter(localFallbackJitterMs); d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Wait for a free slot on the per-service semaphore, but bail out if
	// the caller's context was cancelled in the meantime.
	sem := s.localSems["ollama"]
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
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
	oaiReq.Messages = s.maybeInjectModelIdentity(oaiReq.Messages, model)
	s.mu.RLock()
	reasoning := s.serviceReasoning["ollama"]
	s.mu.RUnlock()
	oaiReq.Reasoning = reasoning
	// Pin think so thinking-capable models do not silently consume num_predict
	// on hidden reasoning. vault reasoning_mode=true → think=true, default → false.
	oaiReq.Think = &reasoning
	// Ollama's OpenAI-compat /v1/chat/completions endpoint silently ignores
	// the top-level `think` field — confirmed (operator host, earlier) where the same
	// request that returns "Hello!" via native /api/chat with think:false
	// returns an empty content + multi-minute hidden reasoning via /v1.
	// applyQwen3NoThinkSuffix encapsulates the qwen3-family inline `/no_think`
	// rule — opt-in per backend (ollama/lmstudio + plugin yaml flag) so other
	// families don't echo the literal text into the response.
	s.applyQwen3NoThinkSuffix(oaiReq, "ollama", model, reasoning)
	// Tell Ollama how long to keep the model resident after this response.
	// Recent Ollama (>=0.3.x) honours top-level keep_alive on the OpenAI
	// compat endpoint; older versions silently ignore it. Empty config means
	// "stay on Ollama's own default" — we don't second-guess the operator.
	if ka := s.cfg.Proxy.OllamaKeepAlive; ka != "" {
		oaiReq.KeepAlive = &ka
	}
	// Forward NumCtx through Ollama's options bag when configured. NumPredict
	// stays in the standard MaxTokens path. Ollama's OpenAI compat layer
	// passes through unknown top-level fields to the runtime where supported;
	// older versions ignore — same no-regression rationale as keep_alive.
	if n := s.cfg.Proxy.OllamaNumCtx; n > 0 {
		nc := n
		oaiReq.Options = &OllamaOptions{NumCtx: &nc}
	}
	if dropped := truncateOldestOAIMessages(oaiReq, promptTokenCap()); dropped > 0 {
		log.Printf("[proxy] ollama prompt cap: dropped %d oldest messages (cap=%d tokens)", dropped, promptTokenCap())
	}
	data, _ := json.Marshal(oaiReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Ollama 요청 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Reuse the long-lived Ollama client — see Server.ollamaHTTP for why.
	resp, err := s.ollamaHTTP.Do(httpReq)
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

// callLocalService: generic OpenAI-compatible local server. Used for every
// non-cloud backend that speaks /v1/chat/completions — LM Studio, vLLM,
// llama.cpp, text-generation-webui, LocalAI, Jan, KoboldCpp, TabbyAPI,
// mlx_lm.server, LiteLLM proxy, *and another wall-vault instance running
// in hub mode*. The transport-level differences (TLS trust, bearer auth,
// default URL) are read from the per-service plugin yaml so the same code
// path serves every backend without if/else hardcoding.
func (s *Server) callLocalService(ctx context.Context, serviceID, model string, req *GeminiRequest) (*GeminiResponse, error) {
	plugin := s.pluginByID[serviceID]

	// URL resolution: env var > plugin.DefaultURL > vault SSE > built-in
	// default. See resolveLocalServiceURLs for rationale; model-aware so a
	// multi-instance pool routes per-request.
	baseURL := s.resolveLocalServiceURLForModel(serviceID, model)
	if baseURL == "" {
		return nil, fmt.Errorf("%s: URL 미설정", serviceID)
	}
	// Drop the "@local<N>" instance pin before forwarding to the
	// OpenAI-compatible backend.
	model = stripInstanceAlias(model)

	// Fleet time distribution — same AgentOffset + FallbackJitter as callOllama.
	// Restored in v0.2.27. See timing.go.
	if d := AgentOffset(s.cfg.Proxy.ClientID, localAgentOffsetMs) +
		FallbackJitter(localFallbackJitterMs); d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Per-service cap-1 semaphore. Local inference is memory-bound, so
	// two concurrent requests on the same backend usually run slower
	// than two sequential ones. If the caller cancels while queued,
	// bail out cleanly.
	if sem, ok := s.localSems[serviceID]; ok {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Strip provider prefix from model (e.g. "google/gemma-4-26b-a4b" → "gemma-4-26b-a4b").
	// Skipped for hub-topology plugins (preserve_model_id: true) — the receiving
	// wall-vault needs the publisher prefix for correct service routing, and a
	// bare "gemma-…" arriving at a hub triggers inferServiceFromBareModel's
	// google promotion path, which then 502s out as "Google: 모델 없음 (…)".
	if plugin == nil || !plugin.PreserveModelID {
		if i := strings.Index(model, "/"); i >= 0 {
			model = model[i+1:]
		}
	}

	oaiReq := GeminiToOpenAI(model, req)
	oaiReq.Stream = false
	oaiReq.Messages = s.maybeInjectModelIdentity(oaiReq.Messages, model)
	s.mu.RLock()
	reasoning := s.serviceReasoning[serviceID]
	s.mu.RUnlock()
	oaiReq.Reasoning = reasoning
	oaiReq.Think = &reasoning
	// Qwen3 family also honours an inline `/no_think` token. applyQwen3NoThinkSuffix
	// covers both the historic native LM Studio opt-in and any plugin yaml
	// that sets inline_no_think_for_qwen3 — so vllm/llamacpp/jan/koboldcpp/…
	// can opt in via yaml without a Go-side switch edit, and backends whose
	// templates don't strip the marker stay safe by default.
	s.applyQwen3NoThinkSuffix(oaiReq, serviceID, model, reasoning)
	if dropped := truncateOldestOAIMessages(oaiReq, promptTokenCap()); dropped > 0 {
		log.Printf("[proxy] %s prompt cap: dropped %d oldest messages (cap=%d tokens)", serviceID, dropped, promptTokenCap())
	}
	data, _ := json.Marshal(oaiReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s 요청 생성 실패: %w", serviceID, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Plugin-driven auth: a plugin yaml with auth.type=bearer means this
	// backend expects an Authorization: Bearer <token> header. The token
	// source is the proxy's vault_token — same secret already used to
	// authenticate this proxy to its vault, which a remote wall-vault
	// running in hub mode treats as the caller's identity.
	if plugin != nil && plugin.Auth.Type == "bearer" && s.cfg.Proxy.VaultToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	}
	// Plugin-driven TLS trust: a plugin yaml with tls_internal_ca=true
	// uses internalHTTPClient so the wall-vault internal CA is trusted.
	// This lets a hub backend present a self-signed cert without the
	// caller having to install the CA into the OS trust store. The
	// internal-CA path keeps its own short-lived client (different root
	// CAs can't share a pool); the default path uses the shared
	// dispatchHTTP pool so keep-alive actually works under sustained
	// fan-out to a single upstream host.
	var client *http.Client
	if plugin != nil && plugin.TLSInternalCA {
		client = internalHTTPClient(10 * time.Minute)
	} else {
		client = s.dispatchHTTP
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s 연결 실패 (%s): %w", serviceID, baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, translateBackendError(serviceID, resp.StatusCode, body)
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
	return s.dispatchHTTP.Do(req)
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
	return parseProviderModelDepth(svc, mdl, 0)
}

// parseProviderModelDepth bounds the "custom/" recursion. A pathological input
// like "custom/custom/.../foo" would otherwise unwind through Go's stack as
// many times as the prefix repeats — small in normal traffic but cheap to
// guard against given the parser sits on the request hot path.
const maxParseProviderDepth = 8

// oaiCompatServices: backends that speak OpenAI /v1/chat/completions and
// publish models in path-style ids (publisher/model). Their publisher segment
// frequently collides with native service names (google/, qwen/, anthropic/),
// so parseProviderModelDepth uses this single set for both directions:
//   - caller chose one as svc → honour it regardless of the body model's prefix
//   - caller wrote one as the model prefix → route to that backend
//
// Default entries cover the most common community OAI-compat backends out
// of the box. Operator-installed plugin yamls with request_format=="openai"
// (or unset, since "" defaults to openai-compat in dispatch) are merged in
// at server boot via registerOAICompatPlugin so a drop-in plugin file is
// enough — no Go edit required for new backends.
var oaiCompatServices = map[string]struct{}{
	"lmstudio":      {},
	"vllm":          {},
	"llamacpp":      {},
	"tgwui":         {},
	"localai":       {},
	"jan":           {},
	"koboldcpp":     {},
	"tabbyapi":      {},
	"mlx-server":    {},
	"litellm-proxy": {},
}

// oaiCompatServicesMu guards oaiCompatServices on the boot-time merge from
// plugin yamls. Reads are not synchronised (the map is treated as
// immutable after Server.New finishes), but the merge writes happen from
// init goroutines so a mutex prevents Go's race detector from flagging
// the construction phase. Production traffic only reads.
var oaiCompatServicesMu sync.Mutex

// registerOAICompatPlugin adds a plugin id to oaiCompatServices so
// parseProviderModelDepth and the dispatch path treat it as
// OpenAI-compatible without requiring a Go-side edit. Idempotent.
func registerOAICompatPlugin(id string) {
	if id == "" {
		return
	}
	oaiCompatServicesMu.Lock()
	oaiCompatServices[id] = struct{}{}
	oaiCompatServicesMu.Unlock()
}

func parseProviderModelDepth(svc, mdl string, depth int) (string, string) {
	if depth >= maxParseProviderDepth {
		return svc, mdl
	}
	// Strip Ollama :cloud suffix — route cloud variants via OpenRouter
	if strings.HasSuffix(mdl, ":cloud") {
		bare := strings.TrimSuffix(mdl, ":cloud")
		return "openrouter", bare
	}

	if !strings.Contains(mdl, "/") {
		// No provider prefix in the model name. Try to infer the service from
		// the bare name itself — a request like {"model": "gemini-2.5-flash"}
		// addressed to a client whose preferred_service is ollama would
		// otherwise be force-routed to ollama and 404 with "model not found",
		// because ollama does not host google models. Surfaced by <separate incident>.
		if inferred := inferServiceFromBareModel(mdl); inferred != "" && inferred != svc {
			// Suppress the "colon ⇒ ollama" hijack when the caller has
			// already chosen a non-ollama service: an EconoWorld client
			// whose ai_config.json carries model="qwen3.6:27b" but whose
			// vault-side default_service was deliberately moved to lmstudio
			// must still land on lmstudio ((operator host, earlier) incident — the
			// hijack force-routed the call into a dead ollama and the
			// operator's explicit lmstudio choice was silently ignored).
			// The <separate incident> cure (gemini-* misrouted to ollama) survives
			// because cloud-service inferences (anthropic / google / openai)
			// fall through to the unconditional return below.
			if inferred == "ollama" && svc != "" && svc != "ollama" {
				return svc, mdl
			}
			return inferred, mdl
		}
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
	// OAI-compat backends (LM Studio / vLLM / llama.cpp / tgwui / LocalAI /
	// Jan / koboldcpp / tabbyapi / mlx-server / litellm-proxy) publish models
	// in path-style ids (publisher/model). Their publisher segment frequently
	// collides with another service's name (google/, qwen/, anthropic/, …),
	// so when the caller explicitly chose one of these the parser must honour
	// that choice and not re-route based on the prefix segment. Was previously
	// duplicated as case-by-case if blocks per backend; consolidating into a
	// single set keeps it in lockstep with the dispatch case below.
	if _, oai := oaiCompatServices[svc]; oai && prefix != svc && prefix != "ollama" {
		return svc, mdl
	}
	// Prefix-as-OAI-compat: caller wrote "lmstudio/qwen3.6-27b" (or any other
	// OAI-compat backend's id as the prefix) — route to that backend. The
	// full mdl (e.g. "lmstudio/qwen/qwen3.6-27b") flows on so callLocalService's
	// own one-segment prefix stripper drops only the leading "lmstudio/",
	// preserving "qwen/qwen3.6-27b" for LM Studio / vLLM / llama.cpp.
	if _, oai := oaiCompatServices[prefix]; oai {
		return prefix, mdl
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
		// When the caller's preferred service is ollama, treat
		// "anthropic/<model>" as a local Ollama Modelfile alias on this host —
		// the operator wired vault model_override="anthropic/<model>" so a
		// fleet-wide call mirrors Claude naming while resolving against a
		// locally built tag (e.g. FROM qwen3.6:27b). Without this, the prefix
		// would force the call to OpenRouter regardless of preferred_service,
		// bypassing the operator-pinned local route.
		if svc == "ollama" {
			return "ollama", mdl
		}
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
		return parseProviderModelDepth(svc, bare, depth+1) // bare = "google/gemini-..." etc.

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

// inferServiceFromBareModel picks a service for a model name that has no
// explicit "provider/" prefix. Returns "" when the name can't be matched —
// the caller then keeps whatever preferred_service was already in play.
//
// Mapping rules:
//   - any name containing ":"  →  ollama        (tag-style: gemma4:26b,
//                                                  qwen3.6:27b, llama3:8b…)
//   - claude-*                  →  anthropic
//   - gemini-*  /  gemma-*      →  google         (no colon — Google catalogue)
//   - gpt-*  /  o1*  /  o3*  /  o4*  →  openai
//   - everything else           →  ""             (caller's choice stands)
//
// Introduced in v0.2.28 after <separate incident>: a request for "gemini-2.5-flash"
// arrived at a client whose preferred_service was ollama and was forwarded
// to ollama unchanged, producing a 404 from ollama and a noisy cascade of
// downstream errors. The cure isn't to override preferred_service for every
// call — only when the bare model name unambiguously belongs to a different
// service can wall-vault correct the route. Ambiguous or unknown names
// (e.g. "qwen3.5-32b" without a colon) leave the caller's choice intact.
func inferServiceFromBareModel(mdl string) string {
	if mdl == "" {
		return ""
	}
	// Tag-style names (foo:bar) are ollama's local-model convention.
	if strings.Contains(mdl, ":") {
		return "ollama"
	}
	// Local quantized variants of any cloud model (e.g. unsloth GGUF Gemma,
	// Llama, Qwen) live on llamacpp — checked before the cloud-prefix table
	// so a name like "gemma-4-26B-A4B-it-UD-Q4_K_M.gguf" doesn't get force
	// routed to Google's API and 400 with "unexpected model name format".
	if isLocalQuantizedModel(mdl) {
		return "llamacpp"
	}
	switch {
	case strings.HasPrefix(mdl, "claude-"):
		return "anthropic"
	case strings.HasPrefix(mdl, "gemini-"),
		strings.HasPrefix(mdl, "gemma-"):
		return "google"
	case strings.HasPrefix(mdl, "gpt-"),
		mdl == "o1", mdl == "o1-mini",
		strings.HasPrefix(mdl, "o3"),
		strings.HasPrefix(mdl, "o4"):
		return "openai"
	}
	return ""
}

// isLocalQuantizedModel returns true when the model name carries markers
// indicating a locally-served quantized weight file (GGUF, Unsloth Dynamic,
// k-quant suffixes, IQ-quant suffixes, bf16, MXFP). Such names belong on
// llamacpp / lmstudio / vllm, not on a cloud provider's API.
func isLocalQuantizedModel(name string) bool {
	n := strings.ToLower(name)
	if strings.Contains(n, ".gguf") || strings.Contains(n, "-ud-") {
		return true
	}
	for _, marker := range []string{
		"-q2_", "-q3_", "-q4_", "-q5_", "-q6_", "-q8_",
		"-iq1_", "-iq2_", "-iq3_", "-iq4_",
		"-bf16", "-mxfp4", "-fp8",
	} {
		if strings.Contains(n, marker) {
			return true
		}
	}
	return false
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
// warnIfPluginURLLooksRemoteHTTP emits a single startup log line when a
// service plugin's default_url (or generate endpoint) points at a non-local
// host over plain HTTP. That combination almost always means the operator
// pasted a remote backend URL but forgot the "s" in "https://" — sending
// any bearer token in the clear, and (for self-signed wall-vault hubs)
// guaranteeing a TLS handshake error on the first request anyway.
//
// Local hosts (localhost, 127.0.0.1, ::1, *.local) are exempt: an http://
// LM Studio / ollama / llama.cpp on the same machine is the normal case
// and warning about it would be noise.
func warnIfPluginURLLooksRemoteHTTP(p config.ServicePlugin) {
	candidates := []string{p.DefaultURL, p.Endpoints.Generate}
	for _, raw := range candidates {
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u == nil {
			continue
		}
		if u.Scheme != "http" {
			continue
		}
		host := u.Hostname()
		if host == "" || host == "localhost" || host == "127.0.0.1" || host == "::1" {
			continue
		}
		if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".localhost") {
			continue
		}
		log.Printf("[plugin] warn: %s url=%s reaches remote host over plain HTTP — "+
			"use https:// (set tls_internal_ca: true if the backend is a "+
			"self-signed wall-vault hub)", p.ID, raw)
		return
	}
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
