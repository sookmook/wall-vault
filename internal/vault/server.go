package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/theme"
)

// Server: 키 금고 HTTP 서버
type Server struct {
	cfg    *config.Config
	store  *Store
	broker *Broker
}

func NewServer(cfg *config.Config) (*Server, error) {
	store, err := NewStore(cfg.Vault.DataDir, cfg.Vault.MasterPass)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		cfg:    cfg,
		store:  store,
		broker: NewBroker(),
	}
	// 일일 사용량 자정 리셋 시작
	go srv.startDailyReset()
	return srv, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 공개
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.broker.ServeHTTP)
	mux.HandleFunc("/api/clients", s.handlePublicClients)

	// 프록시 전용 (클라이언트 토큰 인증)
	mux.HandleFunc("/api/keys", s.clientAuth(s.handleProxyKeys))       // 복호화된 키 목록
	mux.HandleFunc("/api/heartbeat", s.clientAuth(s.handleHeartbeat))  // Heartbeat 수신

	// 관리자
	mux.HandleFunc("/admin/clients", s.adminAuth(s.handleAdminClients))
	mux.HandleFunc("/admin/clients/", s.adminAuth(s.handleAdminClientsID))
	mux.HandleFunc("/admin/keys", s.adminAuth(s.handleAdminKeys))
	mux.HandleFunc("/admin/keys/", s.adminAuth(s.handleAdminKeysID))
	mux.HandleFunc("/admin/keys/reset", s.adminAuth(s.handleResetUsage))
	mux.HandleFunc("/admin/heartbeat", s.adminAuth(s.handleHeartbeat)) // 관리자도 가능
	mux.HandleFunc("/admin/proxies", s.adminAuth(s.handleAdminProxies))

	// 대시보드 UI
	mux.HandleFunc("/", s.handleDashboard)

	return mux
}

// ─── 미들웨어 ────────────────────────────────────────────────────────────────

func (s *Server) adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Vault.AdminToken == "" {
			next(w, r)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token != s.cfg.Vault.AdminToken {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// clientAuth: 등록된 클라이언트 토큰으로 인증
func (s *Server) clientAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		// 관리자 토큰도 허용
		if s.cfg.Vault.AdminToken != "" && token == s.cfg.Vault.AdminToken {
			next(w, r)
			return
		}
		// 클라이언트 토큰 확인
		if s.cfg.Vault.AdminToken == "" || s.store.GetClientByToken(token) != nil {
			next(w, r)
			return
		}
		jsonError(w, "unauthorized", http.StatusUnauthorized)
	}
}

// ─── 공개 API ────────────────────────────────────────────────────────────────

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	keys := s.store.ListKeys()
	clients := s.store.ListClients()
	jsonOK(w, map[string]interface{}{
		"status":  "ok",
		"version": "v0.1.0",
		"keys":    len(keys),
		"clients": len(clients),
		"sse":     s.broker.Count(),
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

// ─── 클라이언트 CRUD ──────────────────────────────────────────────────────────

func (s *Server) handleAdminClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonOK(w, s.store.ListClients())
	case http.MethodPost:
		var body struct {
			ID              string   `json:"id"`
			Name            string   `json:"name"`
			Token           string   `json:"token"`
			DefaultService  string   `json:"default_service"`
			DefaultModel    string   `json:"default_model"`
			AllowedServices []string `json:"allowed_services"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if body.Token == "" {
			body.Token = newID() + newID()
		}
		c, err := s.store.AddClient(body.ID, body.Name, body.Token, body.DefaultService, body.DefaultModel, body.AllowedServices)
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
		var body struct {
			DefaultService  string   `json:"default_service"`
			DefaultModel    string   `json:"default_model"`
			AllowedServices []string `json:"allowed_services"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := s.store.UpdateClient(id, body.DefaultService, body.DefaultModel, body.AllowedServices); err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		// SSE 브로드캐스트
		s.broker.Broadcast(SSEEvent{
			Type: "config_change",
			Data: ConfigChangeEvent{ClientID: id, Service: body.DefaultService, Model: body.DefaultModel},
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

// ─── 키 CRUD ─────────────────────────────────────────────────────────────────

func (s *Server) handleAdminKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		keys := s.store.ListKeys()
		// 암호화된 키는 내려주지 않음
		type safe struct {
			ID            string    `json:"id"`
			Service       string    `json:"service"`
			Label         string    `json:"label"`
			TodayUsage    int       `json:"today_usage"`
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
				TodayUsage: k.TodayUsage, DailyLimit: k.DailyLimit,
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
	if err := s.store.DeleteKey(id); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
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
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminProxies(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.store.ListProxies())
}

// ─── 프록시 전용 API ──────────────────────────────────────────────────────────

// handleProxyKeys: 프록시에게 복호화된 키 목록 제공 (클라이언트 토큰 인증)
func (s *Server) handleProxyKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "GET required", http.StatusMethodNotAllowed)
		return
	}

	// 요청 클라이언트 확인
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	client := s.store.GetClientByToken(token)
	serviceFilter := r.URL.Query().Get("service")

	keys := s.store.ListKeys()
	type safeKey struct {
		ID         string `json:"id"`
		Service    string `json:"service"`
		PlainKey   string `json:"plain_key"`
		DailyLimit int    `json:"daily_limit"`
	}

	result := make([]safeKey, 0, len(keys))
	for _, k := range keys {
		// 서비스 필터
		if serviceFilter != "" && k.Service != serviceFilter {
			continue
		}
		// 클라이언트의 허용 서비스 확인
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
		// 키 복호화
		plain, err := decryptKey(k.EncryptedKey, s.cfg.Vault.MasterPass)
		if err != nil {
			continue
		}
		result = append(result, safeKey{
			ID:         k.ID,
			Service:    k.Service,
			PlainKey:   plain,
			DailyLimit: k.DailyLimit,
		})
	}
	jsonOK(w, result)
}

// ─── 사용량 초기화 ────────────────────────────────────────────────────────────

func (s *Server) handleResetUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	s.store.ResetDailyUsage()
	s.broker.Broadcast(SSEEvent{Type: "usage_reset", Data: map[string]string{"time": time.Now().Format(time.RFC3339)}})
	jsonOK(w, map[string]string{"status": "reset", "time": time.Now().Format(time.RFC3339)})
}

// ─── 일일 자정 리셋 ───────────────────────────────────────────────────────────

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

// ─── 대시보드 UI ──────────────────────────────────────────────────────────────

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	t := theme.Get(s.cfg.Theme)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, buildDashboard(s, t))
}

// ─── 유틸 ────────────────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
