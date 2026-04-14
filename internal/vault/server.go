package vault

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/middleware"
	"github.com/sookmook/wall-vault/internal/models"
	layouts "github.com/sookmook/wall-vault/internal/vault/views/layouts"
	mainview "github.com/sookmook/wall-vault/internal/vault/views/main"
	sidebar "github.com/sookmook/wall-vault/internal/vault/views/sidebar"
	slideover "github.com/sookmook/wall-vault/internal/vault/views/slideover"
)

// request body size limits
const (
	maxBodySize       = 1 << 20 // 1 MB for normal JSON endpoints
	maxAvatarBodySize = 3 << 20 // 3 MB for client CRUD (includes base64 avatar; client-side resized to <=256px)
	maxHeartbeatSize  = 5 << 20 // 5 MB for heartbeat (includes base64 avatar)
)

// secureCompare performs constant-time string comparison to prevent timing attacks.
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// authLimiter: tracks auth failure count per IP (rate limiting)
type authLimiter struct {
	mu    sync.Mutex
	fails map[string][]time.Time
}

func newAuthLimiter() *authLimiter { return &authLimiter{fails: make(map[string][]time.Time)} }

// blocked: block if 10 or more failures within 15 minutes
func (al *authLimiter) blocked(ip string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()
	cutoff := time.Now().Add(-15 * time.Minute)
	var recent []time.Time
	for _, t := range al.fails[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	al.fails[ip] = recent
	return len(recent) >= 10
}

func (al *authLimiter) record(ip string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.fails[ip] = append(al.fails[ip], time.Now())
}

// Version is set from main via ldflags injection.
var Version = "dev"

// Server: key vault HTTP server
type Server struct {
	cfg       *config.Config
	store     *Store
	broker    *Broker
	registry  *models.Registry // model cache
	cfgPath   string           // config file path to save on theme change
	startedAt time.Time        // service start time
	limiter   *authLimiter     // auth failure rate limiter
}

// SetConfigPath: specify the config file path to use for saving theme
func (s *Server) SetConfigPath(path string) {
	s.cfgPath = path
}

func NewServer(cfg *config.Config) (*Server, error) {
	store, err := NewStore(cfg.Vault.DataDir, cfg.Vault.MasterPass)
	if err != nil {
		return nil, err
	}
	// if theme/language is saved in vault.json, it takes priority over cfg
	if st := store.GetSettings(); st.Theme != "" || st.Lang != "" {
		if st.Theme != "" {
			cfg.Theme = st.Theme
		}
		if st.Lang != "" {
			cfg.Lang = st.Lang
		}
	}
	if cfg.Vault.AdminToken == "" {
		log.Println("[WARNING] ======================================================")
		log.Println("[WARNING]  No admin token configured.")
		log.Println("[WARNING]  All admin endpoints are UNPROTECTED.")
		log.Println("[WARNING]  Set admin_token in config to secure the vault.")
		log.Println("[WARNING] ======================================================")
	}
	srv := &Server{
		cfg:       cfg,
		store:     store,
		broker:    NewBroker(),
		registry:  models.NewRegistry(10 * time.Minute),
		startedAt: time.Now(),
		limiter:   newAuthLimiter(),
	}
	// send full state to every new SSE client (handles reconnect sync)
	srv.broker.OnConnect = func() { srv.broadcastAgentsSync() }
	// start midnight daily usage reset
	go srv.startDailyReset()
	// start periodic status broadcaster so the dashboard detects
	// offline transitions even when no heartbeats arrive
	go srv.startStatusTicker()
	return srv, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// public
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.sseAuth(s.broker.ServeHTTP))
	mux.HandleFunc("/api/clients", s.handlePublicClients)

	// proxy-only (client token auth)
	mux.HandleFunc("/api/keys", s.clientAuth(s.handleProxyKeys))               // decrypted key list
	mux.HandleFunc("/api/heartbeat", s.clientAuth(s.handleHeartbeat))          // heartbeat receiver
	mux.HandleFunc("/api/config", s.clientAuth(s.handleClientConfig))          // client self-config change
	mux.HandleFunc("/api/services", s.clientAuth(s.handleProxyServices))       // proxy-enabled service list
	mux.HandleFunc("/api/token/config", s.handleTokenConfig)                   // token→model lookup (for third-party clients)

	// admin
	mux.HandleFunc("/admin/theme", s.adminAuth(s.handleAdminTheme))
	mux.HandleFunc("/admin/lang", s.adminAuth(s.handleAdminLang))
	mux.HandleFunc("/admin/clients", s.adminAuth(s.handleAdminClients))
	mux.HandleFunc("/admin/clients/reorder", s.adminAuth(s.handleAdminClientsReorder))
	mux.HandleFunc("/admin/clients/", s.adminAuth(s.handleAdminClientsID))
	mux.HandleFunc("/admin/keys", s.adminAuth(s.handleAdminKeys))
	mux.HandleFunc("/admin/keys/", s.adminAuth(s.handleAdminKeysID))
	mux.HandleFunc("/admin/keys/reset", s.adminAuth(s.handleResetUsage))
	mux.HandleFunc("/admin/heartbeat", s.adminAuth(s.handleHeartbeat)) // admin also allowed
	mux.HandleFunc("/admin/proxies", s.adminAuth(s.handleAdminProxies))
	mux.HandleFunc("/admin/services", s.adminAuth(s.handleAdminServices))
	mux.HandleFunc("/admin/services/", s.adminAuth(s.handleAdminServicesID))
	mux.HandleFunc("/admin/models", s.adminAuth(s.handleAdminModels))

	// HTMX fragments
	s.RegisterHXRoutes(mux)

	// logo
	mux.HandleFunc("/logo", s.handleLogo)

	// dashboard UI
	mux.HandleFunc("/", s.handleDashboard)

	return middleware.Chain(mux,
		middleware.Recovery,
		middleware.CORS,
		middleware.Logger,
	)
}

// ─── Middleware ───────────────────────────────────────────────────────────────

func (s *Server) adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Vault.AdminToken == "" {
			next(w, r)
			return
		}
		ip := realIP(r)
		if s.limiter.blocked(ip) {
			jsonError(w, "too many failed attempts", http.StatusTooManyRequests)
			return
		}
		token := bearerToken(r)
		if !secureCompare(token, s.cfg.Vault.AdminToken) {
			s.limiter.record(ip)
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// bearerToken extracts the Bearer token from an Authorization header.
func bearerToken(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

// clientAuth: authenticate with a registered client token
func (s *Server) clientAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		// admin token also accepted (constant-time comparison)
		if s.cfg.Vault.AdminToken != "" && secureCompare(token, s.cfg.Vault.AdminToken) {
			next(w, r)
			return
		}
		// verify client token
		if s.cfg.Vault.AdminToken == "" || s.store.GetClientByToken(token) != nil {
			next(w, r)
			return
		}
		jsonError(w, "unauthorized", http.StatusUnauthorized)
	}
}

// sseAuth: protect the SSE endpoint when admin token is configured.
// If no admin token is set, the endpoint remains open (backward compatible).
// When a token IS configured, require a valid client token or admin token.
// The token can be provided via Authorization header or ?token= query param
// (the query param is needed because EventSource API cannot set headers).
func (s *Server) sseAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no admin token is configured, SSE is open (backward compatible)
		if s.cfg.Vault.AdminToken == "" {
			next(w, r)
			return
		}
		// Try Authorization header first, then ?token= query param
		token := bearerToken(r)
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		// Accept admin token
		if secureCompare(token, s.cfg.Vault.AdminToken) {
			next(w, r)
			return
		}
		// Accept registered client token
		if s.store.GetClientByToken(token) != nil {
			next(w, r)
			return
		}
		jsonError(w, "unauthorized", http.StatusUnauthorized)
	}
}

// ─── Public API ───────────────────────────────────────────────────────────────

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// When admin token is configured, only expose detailed counts to authenticated callers.
	// Unauthenticated callers get a minimal response to reduce information leakage.
	authenticated := false
	if s.cfg.Vault.AdminToken == "" {
		authenticated = true // no token configured — open access (backward compatible)
	} else {
		token := bearerToken(r)
		if secureCompare(token, s.cfg.Vault.AdminToken) || s.store.GetClientByToken(token) != nil {
			authenticated = true
		}
	}

	if !authenticated {
		jsonOK(w, map[string]interface{}{
			"status":  "ok",
			"version": Version,
		})
		return
	}

	keys := s.store.ListKeys()
	clients := s.store.ListClients()
	jsonOK(w, map[string]interface{}{
		"status":  "ok",
		"version": Version,
		"keys":    len(keys),
		"clients": len(clients),
		"sse":     s.broker.Count(),
	})
}

// handleTokenConfig: resolve a client token → {default_service, default_model}
// Used by third-party clients (Cline, Cursor, etc.) so the proxy can override
// the model they send with the dashboard-configured model for that token.
func (s *Server) handleTokenConfig(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	c := s.store.GetClientByToken(token)
	if c == nil {
		jsonError(w, "token not found", http.StatusNotFound)
		return
	}
	jsonOK(w, map[string]string{
		"id":              c.ID,
		"default_service": c.DefaultService,
		"default_model":   c.DefaultModel,
	})
}

func (s *Server) handlePublicClients(w http.ResponseWriter, r *http.Request) {
	clients := s.store.ListClients()

	// When authenticated, include agent_type (needed by proxy's syncFromVault).
	// When unauthenticated, omit agent_type to reduce information leakage.
	authenticated := false
	if s.cfg.Vault.AdminToken == "" {
		authenticated = true
	} else {
		token := bearerToken(r)
		if secureCompare(token, s.cfg.Vault.AdminToken) || s.store.GetClientByToken(token) != nil {
			authenticated = true
		}
	}

	if authenticated {
		type pub struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			DefaultService string `json:"default_service"`
			DefaultModel   string `json:"default_model"`
			AgentType      string `json:"agent_type,omitempty"`
		}
		result := make([]pub, 0, len(clients))
		for _, c := range clients {
			result = append(result, pub{c.ID, c.Name, c.DefaultService, c.DefaultModel, c.AgentType})
		}
		jsonOK(w, result)
		return
	}

	type pub struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		DefaultService string `json:"default_service"`
		DefaultModel   string `json:"default_model"`
	}
	result := make([]pub, 0, len(clients))
	for _, c := range clients {
		result = append(result, pub{c.ID, c.Name, c.DefaultService, c.DefaultModel})
	}
	jsonOK(w, result)
}

// ─── Client CRUD ──────────────────────────────────────────────────────────────

func (s *Server) handleAdminClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonOK(w, s.store.ListClients())
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBodySize)
		var inp ClientInput
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if inp.Token == "" {
			inp.Token = newID() + newID()
		}
		c, err := s.store.AddClient(inp)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, c)
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAdminClientsReorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "PUT required", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var body struct {
		Order []string `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if len(body.Order) == 0 {
		jsonError(w, "order list required", http.StatusBadRequest)
		return
	}
	if err := s.store.ReorderClients(body.Order); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminClientsID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/clients/")
	if id == "" {
		jsonError(w, "client id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		c := s.store.GetClient(id)
		if c == nil {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonOK(w, c)
	case http.MethodPut:
		r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBodySize)
		var inp ClientUpdateInput
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		// Validate model_override against the target service's AllowedModels whitelist.
		// mirrors proxy.ResolveModel whitelist check — see spec §4.2
		if inp.ModelOverride != nil && *inp.ModelOverride != "" {
			// resolve target service: prefer request's preferred_service, then legacy default_service,
			// then fall back to the existing client's PreferredService / DefaultService.
			targetService := ""
			if inp.PreferredService != nil && *inp.PreferredService != "" {
				targetService = *inp.PreferredService
			} else if inp.DefaultService != nil && *inp.DefaultService != "" {
				targetService = *inp.DefaultService
			} else if existing := s.store.GetClient(id); existing != nil {
				if existing.PreferredService != "" {
					targetService = existing.PreferredService
				} else {
					targetService = existing.DefaultService
				}
			}
			if targetService != "" {
				if sv := s.store.GetService(targetService); sv != nil && len(sv.AllowedModels) > 0 {
					allowed := false
					for _, m := range sv.AllowedModels {
						if m == *inp.ModelOverride {
							allowed = true
							break
						}
					}
					if !allowed {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnprocessableEntity)
						fmt.Fprintf(w, `{"error":"model_override %q not in allowed_models of service %q"}`,
							*inp.ModelOverride, targetService)
						return
					}
				}
			}
		}
		if err := s.store.UpdateClient(id, inp); err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		// SSE broadcast (include agent_type so proxies can decide which local agent to update)
		// Prefer v0.2 canonical fields (preferred_service / model_override), fall back to legacy.
		svc, mdl, agentType := "", "", ""
		if inp.PreferredService != nil {
			svc = *inp.PreferredService
		} else if inp.DefaultService != nil {
			svc = *inp.DefaultService
		}
		if inp.ModelOverride != nil {
			mdl = *inp.ModelOverride
		} else if inp.DefaultModel != nil {
			mdl = *inp.DefaultModel
		}
		if c := s.store.GetClient(id); c != nil {
			agentType = c.AgentType
		}
		s.broker.Broadcast(SSEEvent{
			Type: "config_change",
			Data: ConfigChangeEvent{ClientID: id, Service: svc, Model: mdl, AgentType: agentType},
		})
		jsonOK(w, map[string]string{"status": "updated"})
	case http.MethodDelete:
		if err := s.store.DeleteClient(id); err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, map[string]string{"status": "deleted"})
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ─── Key CRUD ─────────────────────────────────────────────────────────────────

func (s *Server) handleAdminKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		keys := s.store.ListKeys()
		// do not expose encrypted keys
		type safe struct {
			ID            string    `json:"id"`
			Service       string    `json:"service"`
			Label         string    `json:"label"`
			TodayUsage    int       `json:"today_usage"`
			TodayAttempts int       `json:"today_attempts"`
			DailyLimit    int       `json:"daily_limit"`
			CooldownUntil time.Time `json:"cooldown_until"`
			LastError     int       `json:"last_error"`
			CreatedAt     time.Time `json:"created_at"`
			Available     bool      `json:"available"`
			UsagePct      int       `json:"usage_pct"`
		}
		result := make([]safe, 0, len(keys))
		for _, k := range keys {
			result = append(result, safe{
				ID: k.ID, Service: k.Service, Label: k.Label,
				TodayUsage: k.TodayUsage, TodayAttempts: k.TodayAttempts, DailyLimit: k.DailyLimit,
				CooldownUntil: k.CooldownUntil, LastError: k.LastError,
				CreatedAt: k.CreatedAt, Available: k.IsAvailable(), UsagePct: k.UsagePct(),
			})
		}
		jsonOK(w, result)
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		var body struct {
			Service    string `json:"service"`
			Key        string `json:"key"`
			Label      string `json:"label"`
			DailyLimit int    `json:"daily_limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
			jsonError(w, "service, key 필수", http.StatusBadRequest)
			return
		}
		k, err := s.store.AddKey(body.Service, body.Key, body.Label, body.DailyLimit)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// immediately reflect cloud service enabled state after key add
		s.store.ReconcileCloudServices()
		s.broker.Broadcast(SSEEvent{Type: "key_added", Data: map[string]string{"service": body.Service}})
		jsonOK(w, k)
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAdminKeysID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/keys/")
	if r.Method != http.MethodDelete {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// record service info before deletion
	deletedSvc := ""
	for _, k := range s.store.ListKeys() {
		if k.ID == id {
			deletedSvc = k.Service
			break
		}
	}
	if err := s.store.DeleteKey(id); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	// immediately reflect cloud service enabled state after key deletion
	s.store.ReconcileCloudServices()
	s.broker.Broadcast(SSEEvent{Type: "key_deleted", Data: map[string]string{"service": deletedSvc}})
	jsonOK(w, map[string]string{"status": "deleted"})
}

// ─── Heartbeat ────────────────────────────────────────────────────────────────

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxHeartbeatSize)
	var ps ProxyStatus
	if err := json.NewDecoder(r.Body).Decode(&ps); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if ps.ClientID == "" {
		jsonError(w, "client_id required", http.StatusBadRequest)
		return
	}
	s.store.UpdateProxyStatus(&ps)
	// if proxy sent an avatar, persist it to the client record
	if ps.Avatar != "" {
		_ = s.store.UpdateClient(ps.ClientID, ClientUpdateInput{Avatar: &ps.Avatar})
	}
	// sync proxy key usage, attempts, and cooldowns into vault store (single lock + save)
	s.store.BatchUpdateKeyMetrics(ps.KeyUsage, ps.KeyAttempts, ps.KeyCooldowns)
	// update ProxyStatus for each recently-active sub-client served by this proxy
	for _, ac := range ps.ActiveClients {
		if ac.ClientID == "" || ac.ClientID == ps.ClientID {
			continue
		}
		s.store.UpdateProxyStatus(&ProxyStatus{
			ClientID: ac.ClientID,
			Service:  ac.Service,
			Model:    ac.Model,
			Version:  ps.Version,
			SSE:      false,
		})
	}

	// single unified broadcast: status + details for every client card
	s.broadcastAgentsSync()

	// always broadcast full key states so the dashboard reflects reality without a fetch
	{
		allKeys := s.store.ListKeys()
		now := time.Now()
		keyStates := make([]map[string]interface{}, 0, len(allKeys))
		for _, k := range allKeys {
			cdStr := ""
			if k.CooldownUntil.After(now) {
				cdStr = k.CooldownUntil.UTC().Format(time.RFC3339)
			}
			keyStates = append(keyStates, map[string]interface{}{
				"id":             k.ID,
				"service":        k.Service,
				"today_usage":    k.TodayUsage,
				"today_attempts": k.TodayAttempts,
				"daily_limit":    k.DailyLimit,
				"cooldown_until": cdStr,
			})
		}
		s.broker.Broadcast(SSEEvent{
			Type: "usage_update",
			Data: map[string]interface{}{"keys": keyStates},
		})
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminProxies(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.store.ListProxies())
}

// ─── Proxy-Only API ───────────────────────────────────────────────────────────

// handleClientConfig: client changes its own service/model config (bidirectional sync supported)
func (s *Server) handleClientConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "PUT required", http.StatusMethodNotAllowed)
		return
	}
	// identify client: explicit client_id query param takes priority (avoids
	// ambiguity when multiple proxies share the same token), then fall back
	// to token-based lookup.
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		if c := s.store.GetClientByToken(token); c != nil {
			clientID = c.ID
		} else if secureCompare(token, s.cfg.Vault.AdminToken) {
			// admin token without client_id — cannot resolve
		}
	}
	if clientID == "" {
		jsonError(w, "client not found", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var inp struct {
		Service string `json:"service"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateClient(clientID, ClientUpdateInput{
		DefaultService: &inp.Service,
		DefaultModel:   &inp.Model,
	}); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	cfgEvt := ConfigChangeEvent{ClientID: clientID, Service: inp.Service, Model: inp.Model}
	if c := s.store.GetClient(clientID); c != nil {
		cfgEvt.AgentType = c.AgentType
	}
	s.broker.Broadcast(SSEEvent{Type: "config_change", Data: cfgEvt})
	jsonOK(w, map[string]string{"status": "updated", "client_id": clientID})
}

// handleProxyServices: returns proxy-enabled services with local URLs (client token auth)
func (s *Server) handleProxyServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "GET required", http.StatusMethodNotAllowed)
		return
	}
	jsonOK(w, s.store.ListProxyEnabledServicesInfo())
}

// handleProxyKeys: provide decrypted key list to proxy (client token auth)
func (s *Server) handleProxyKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "GET required", http.StatusMethodNotAllowed)
		return
	}

	// identify requesting client
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	isAdmin := s.cfg.Vault.AdminToken != "" && secureCompare(token, s.cfg.Vault.AdminToken)
	client := s.store.GetClientByToken(token)
	// deny key access to disabled clients
	if client != nil && !client.Enabled {
		jsonError(w, "client disabled", http.StatusForbidden)
		return
	}
	// check IP whitelist (skip for admin token)
	if !isAdmin && client != nil && len(client.IPWhitelist) > 0 {
		if !ipAllowed(realIP(r), client.IPWhitelist) {
			jsonError(w, "ip not allowed", http.StatusForbidden)
			return
		}
	}
	serviceFilter := r.URL.Query().Get("service")

	keys := s.store.ListKeys()
	type safeKey struct {
		ID            string    `json:"id"`
		Service       string    `json:"service"`
		PlainKey      string    `json:"plain_key"`
		DailyLimit    int       `json:"daily_limit"`
		TodayUsage    int       `json:"today_usage"`
		TodayAttempts int       `json:"today_attempts"`
		UsageDate     string    `json:"usage_date"`
		CooldownUntil time.Time `json:"cooldown_until"`
	}

	result := make([]safeKey, 0, len(keys))
	for _, k := range keys {
		// service filter
		if serviceFilter != "" && k.Service != serviceFilter {
			continue
		}
		// check client's allowed services
		if client != nil && len(client.AllowedServices) > 0 {
			allowed := false
			for _, svc := range client.AllowedServices {
				if svc == k.Service {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		// decrypt key
		plain, err := decryptKey(k.EncryptedKey, s.cfg.Vault.MasterPass)
		if err != nil {
			continue
		}
		result = append(result, safeKey{
			ID:            k.ID,
			Service:       k.Service,
			PlainKey:      plain,
			DailyLimit:    k.DailyLimit,
			TodayUsage:    k.TodayUsage,
			TodayAttempts: k.TodayAttempts,
			UsageDate:     k.UsageDate,
			CooldownUntil: k.CooldownUntil,
		})
	}
	jsonOK(w, result)
}

// ─── Usage Reset ──────────────────────────────────────────────────────────────

func (s *Server) handleResetUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	s.store.ResetDailyUsage()
	s.broker.Broadcast(SSEEvent{Type: "usage_reset", Data: map[string]string{"time": time.Now().Format(time.RFC3339)}})
	jsonOK(w, map[string]string{"status": "reset", "time": time.Now().Format(time.RFC3339)})
}

// ─── Language Change ──────────────────────────────────────────────────────────

func (s *Server) handleAdminLang(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "PUT required", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var body struct {
		Lang string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	valid := map[string]bool{"ko": true, "en": true, "zh": true, "es": true,
		"hi": true, "ar": true, "pt": true, "fr": true, "de": true, "ja": true}
	if !valid[body.Lang] {
		jsonError(w, "unknown lang", http.StatusBadRequest)
		return
	}
	s.cfg.Lang = body.Lang
	_ = s.store.SetLang(body.Lang)
	if s.cfgPath != "" {
		_ = config.Save(s.cfg, s.cfgPath)
	}
	jsonOK(w, map[string]string{"lang": body.Lang})
}

// ─── Theme Change ─────────────────────────────────────────────────────────────

func (s *Server) handleAdminTheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "PUT required", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var body struct {
		Theme string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	valid := map[string]bool{"light": true, "dark": true, "gold": true, "cherry": true, "ocean": true, "autumn": true, "winter": true}
	if !valid[body.Theme] {
		jsonError(w, "unknown theme (light|dark|gold|cherry|ocean|autumn|winter)", http.StatusBadRequest)
		return
	}
	s.cfg.Theme = body.Theme
	_ = s.store.SetTheme(body.Theme)
	if s.cfgPath != "" {
		_ = config.Save(s.cfg, s.cfgPath)
	}
	jsonOK(w, map[string]string{"theme": body.Theme})
}

// ─── Service Management ───────────────────────────────────────────────────────

func (s *Server) handleAdminServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonOK(w, s.store.ListServices())
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		var inp ServiceConfig
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil || inp.ID == "" {
			jsonError(w, "id 필수", http.StatusBadRequest)
			return
		}
		inp.Custom = true
		if err := s.store.UpsertService(&inp); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.broker.Broadcast(SSEEvent{Type: "service_changed", Data: map[string]interface{}{
			"action": "added", "id": inp.ID,
			"proxy_services": s.store.ListProxyEnabledServices(),
		}})
		jsonOK(w, map[string]string{"status": "ok"})
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAdminServicesID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/services/")
	// ping sub-route: GET /admin/services/{id}/ping
	if strings.HasSuffix(path, "/ping") {
		id := strings.TrimSuffix(path, "/ping")
		sv := s.store.GetService(id)
		pingURL := ""
		if sv != nil && sv.LocalURL != "" {
			pingURL = sv.LocalURL
		} else {
			// fallback to default port when local_url is not configured
			switch id {
			case "ollama":
				pingURL = "http://localhost:11434"
			case "lmstudio":
				pingURL = "http://localhost:1234"
			case "vllm":
				pingURL = "http://localhost:8000"
			}
		}
		if pingURL == "" {
			jsonOK(w, map[string]interface{}{"ok": false, "reason": "no url"})
			return
		}
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(pingURL)
		if err != nil {
			jsonOK(w, map[string]interface{}{"ok": false, "reason": err.Error()})
			return
		}
		resp.Body.Close()
		jsonOK(w, map[string]interface{}{"ok": resp.StatusCode < 500})
		return
	}
	id := path
	if id == "" {
		jsonError(w, "service id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		var fields map[string]interface{}
		if err := json.Unmarshal(body, &fields); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		existing := s.store.GetService(id)
		if existing == nil {
			jsonError(w, "service not found", http.StatusNotFound)
			return
		}
		// partial update: copy existing, then apply only provided fields
		inp := *existing
		inp.ID = id
		if v, ok := fields["name"].(string); ok && v != "" {
			inp.Name = v
		}
		if v, ok := fields["local_url"].(string); ok {
			inp.LocalURL = v
		}
		if v, ok := fields["enabled"].(bool); ok {
			inp.Enabled = v
		}
		if v, ok := fields["proxy_enabled"].(bool); ok {
			inp.ProxyEnabled = v
		}
		if v, ok := fields["default_model"].(string); ok {
			inp.DefaultModel = v
		}
		if v, ok := fields["allowed_models"]; ok {
			switch val := v.(type) {
			case []interface{}:
				models := make([]string, 0, len(val))
				for _, m := range val {
					if s, ok := m.(string); ok {
						models = append(models, s)
					}
				}
				inp.AllowedModels = models
			case nil:
				inp.AllowedModels = nil
			}
		}
		if err := s.store.UpsertService(&inp); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.broker.Broadcast(SSEEvent{Type: "service_changed", Data: map[string]interface{}{
			"action": "updated", "id": id,
			"proxy_services": s.store.ListProxyEnabledServices(),
		}})
		jsonOK(w, map[string]string{"status": "updated"})
	case http.MethodDelete:
		if err := s.store.DeleteService(id); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.broker.Broadcast(SSEEvent{Type: "service_changed", Data: map[string]interface{}{
			"action": "deleted", "id": id,
			"proxy_services": s.store.ListProxyEnabledServices(),
		}})
		jsonOK(w, map[string]string{"status": "deleted"})
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminModels: query model list per service (TTL cache)
func (s *Server) handleAdminModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "GET required", http.StatusMethodNotAllowed)
		return
	}
	svcFilter := r.URL.Query().Get("service")

	if s.registry.NeedsRefresh() {
		svcs := s.store.ListServices()
		svcIDs := make([]string, 0, len(svcs))
		for _, sv := range svcs {
			if sv.Enabled {
				svcIDs = append(svcIDs, sv.ID)
			}
		}
		// look up OpenRouter key
		var orKey string
		keys := s.store.ListKeys()
		for _, k := range keys {
			if k.Service == "openrouter" && k.IsAvailable() {
				if plain, err := decryptKey(k.EncryptedKey, s.cfg.Vault.MasterPass); err == nil {
					orKey = plain
					break
				}
			}
		}
		_ = s.registry.Refresh(svcIDs, s.store.ServiceURLMap(), orKey)
	}

	result := s.registry.All(svcFilter)
	jsonOK(w, map[string]interface{}{"models": result, "count": len(result)})
}

// ─── Periodic Status Broadcast ────────────────────────────────────────────────

// broadcastAgentsSync computes the current status for every client and
// broadcasts a single "agents_sync" SSE event containing status + details.
// This is the ONE event the dashboard uses for all card state updates.
func (s *Server) broadcastAgentsSync() {
	clients := s.store.ListClients()
	proxies := s.store.ListProxies()
	proxyMap := map[string]*ProxyStatus{}
	for _, p := range proxies {
		proxyMap[p.ClientID] = p
	}
	now := time.Now()
	items := make([]map[string]interface{}, 0, len(clients))
	for _, c := range clients {
		entry := map[string]interface{}{"id": c.ID}
		if p, ok := proxyMap[c.ID]; ok {
			age := now.Sub(p.UpdatedAt)
			switch {
			case age < 90*time.Second:
				entry["st"] = "live"
			case age < 3*time.Minute:
				entry["st"] = "delay"
			default:
				entry["st"] = "offline"
			}
			// include service/model/version for live and delay cards
			if age < 3*time.Minute {
				entry["svc"] = p.Service
				entry["mdl"] = p.Model
				entry["ver"] = p.Version
			}
			if !p.StartedAt.IsZero() {
				entry["sec"] = p.StartedAt.Unix()
			}
			if p.AgentAlive != nil {
				entry["agent_alive"] = *p.AgentAlive
			}
		} else {
			entry["st"] = "noconn"
		}
		items = append(items, entry)
	}
	s.broker.Broadcast(SSEEvent{
		Type: "agents_sync",
		Data: map[string]interface{}{"clients": items},
	})
}

// startStatusTicker periodically broadcasts agents_sync so the dashboard
// detects offline transitions even when no proxy heartbeats are arriving.
func (s *Server) startStatusTicker() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.broadcastAgentsSync()
	}
}

// ─── Daily Midnight Reset ─────────────────────────────────────────────────────

func (s *Server) startDailyReset() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 30, 0, now.Location())
		time.Sleep(time.Until(next))
		s.store.ResetDailyUsage()
		s.broker.Broadcast(SSEEvent{
			Type: "usage_reset",
			Data: map[string]string{"time": time.Now().Format("2006-01-02")},
		})
	}
}

// ─── Logo ─────────────────────────────────────────────────────────────────────

func (s *Server) handleLogo(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	type candidate struct {
		path string
		ct   string
	}
	candidates := []candidate{
		{filepath.Join(home, ".wall-vault", "logo.png"), "image/png"},
		{filepath.Join(home, ".wall-vault", "logo.jpg"), "image/jpeg"},
		{filepath.Join(home, ".wall-vault", "logo.svg"), "image/svg+xml"},
		{"logo.png", "image/png"},
		{"logo.jpg", "image/jpeg"},
	}
	for _, c := range candidates {
		data, err := os.ReadFile(c.path)
		if err != nil {
			continue
		}
		w.Header().Set("Content-Type", c.ct)
		w.Header().Set("Cache-Control", "max-age=3600")
		w.Write(data) //nolint:errcheck
		return
	}
	// 외부 파일 없으면 바이너리 내장 로고 사용
	if len(embeddedLogo) > 0 {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=3600")
		w.Write(embeddedLogo) //nolint:errcheck
		return
	}
	http.NotFound(w, r)
}

// ─── Dashboard UI ─────────────────────────────────────────────────────────────

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Only handle exactly "/" — fall through elsewhere.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	themeName := s.store.GetSettings().Theme
	if themeName == "" {
		themeName = s.cfg.Theme
	}
	if themeName == "" {
		themeName = "light"
	}
	services := s.store.ListServices()
	clients := s.store.ListClients()

	inner := layouts.Shell(
		sidebar.Sidebar(toSidebarServices(services), toSidebarClients(clients)),
		renderHome(toMainServices(services), toMainClients(clients)),
		slideover.Empty(),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	layouts.Base(themeName, inner).Render(r.Context(), w) //nolint:errcheck
}

// renderHome stacks ServicesGrid + AgentsGrid as the main pane body.
func renderHome(services []*mainview.ServiceVM, clients []*mainview.ClientVM) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := mainview.ServicesGrid(services).Render(ctx, w); err != nil {
			return err
		}
		return mainview.AgentsGrid(clients).Render(ctx, w)
	})
}


// ─── Util ─────────────────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

// realIP: extract IP from RemoteAddr (strip port).
// Does NOT trust X-Forwarded-For to prevent rate-limiter bypass via spoofed headers.
func realIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip == "" {
		return r.RemoteAddr
	}
	return ip
}

// ipAllowed: compare against a list of single IPs or CIDRs
func ipAllowed(remoteIP string, whitelist []string) bool {
	for _, entry := range whitelist {
		entry = strings.TrimSpace(entry)
		if entry == remoteIP {
			return true
		}
		if _, cidr, err := net.ParseCIDR(entry); err == nil {
			if ip := net.ParseIP(remoteIP); ip != nil && cidr.Contains(ip) {
				return true
			}
		}
	}
	return false
}
