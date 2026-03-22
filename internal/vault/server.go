package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/middleware"
	"github.com/sookmook/wall-vault/internal/models"
	"github.com/sookmook/wall-vault/internal/theme"
)

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
	srv := &Server{
		cfg:       cfg,
		store:     store,
		broker:    NewBroker(),
		registry:  models.NewRegistry(10 * time.Minute),
		startedAt: time.Now(),
		limiter:   newAuthLimiter(),
	}
	// start midnight daily usage reset
	go srv.startDailyReset()
	return srv, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// public
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.broker.ServeHTTP)
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
	mux.HandleFunc("/admin/clients/", s.adminAuth(s.handleAdminClientsID))
	mux.HandleFunc("/admin/keys", s.adminAuth(s.handleAdminKeys))
	mux.HandleFunc("/admin/keys/", s.adminAuth(s.handleAdminKeysID))
	mux.HandleFunc("/admin/keys/reset", s.adminAuth(s.handleResetUsage))
	mux.HandleFunc("/admin/heartbeat", s.adminAuth(s.handleHeartbeat)) // admin also allowed
	mux.HandleFunc("/admin/proxies", s.adminAuth(s.handleAdminProxies))
	mux.HandleFunc("/admin/services", s.adminAuth(s.handleAdminServices))
	mux.HandleFunc("/admin/services/", s.adminAuth(s.handleAdminServicesID))
	mux.HandleFunc("/admin/models", s.adminAuth(s.handleAdminModels))

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
		if token != s.cfg.Vault.AdminToken {
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
		// admin token also accepted
		if s.cfg.Vault.AdminToken != "" && token == s.cfg.Vault.AdminToken {
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

// ─── Public API ───────────────────────────────────────────────────────────────

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
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
		var inp ClientUpdateInput
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := s.store.UpdateClient(id, inp); err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		// SSE broadcast
		svc, mdl := "", ""
		if inp.DefaultService != nil {
			svc = *inp.DefaultService
		}
		if inp.DefaultModel != nil {
			mdl = *inp.DefaultModel
		}
		s.broker.Broadcast(SSEEvent{
			Type: "config_change",
			Data: ConfigChangeEvent{ClientID: id, Service: svc, Model: mdl},
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
	// identify client: find by token, or use query param client_id if admin token
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	clientID := ""
	if c := s.store.GetClientByToken(token); c != nil {
		clientID = c.ID
	} else if token == s.cfg.Vault.AdminToken {
		clientID = r.URL.Query().Get("client_id")
	}
	if clientID == "" {
		jsonError(w, "client not found", http.StatusUnauthorized)
		return
	}
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
	s.broker.Broadcast(SSEEvent{
		Type: "config_change",
		Data: ConfigChangeEvent{ClientID: clientID, Service: inp.Service, Model: inp.Model},
	})
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
	isAdmin := s.cfg.Vault.AdminToken != "" && token == s.cfg.Vault.AdminToken
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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	t := theme.Get(s.cfg.Theme)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	fmt.Fprint(w, buildDashboard(s, t))
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

// realIP: extract real IP from X-Forwarded-For or RemoteAddr
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
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
