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
	"github.com/sookmook/wall-vault/internal/middleware"
	"github.com/sookmook/wall-vault/internal/models"
)

// Server: 프록시 HTTP 서버
type Server struct {
	cfg      *config.Config
	mu       sync.RWMutex
	service  string
	model    string
	keyMgr   *KeyManager
	filter   *ToolFilter
	sse      *SSEClient
	registry *models.Registry
	ollamaMu sync.Mutex // 단일 Ollama 동시 요청 보호
}

func NewServer(cfg *config.Config) *Server {
	// 기본 서비스 결정
	defaultSvc := "ollama"
	if len(cfg.Proxy.Services) > 0 {
		defaultSvc = cfg.Proxy.Services[0]
	}

	s := &Server{
		cfg:      cfg,
		service:  defaultSvc,
		model:    "",
		registry: models.NewRegistry(10 * time.Minute),
	}

	s.keyMgr = NewKeyManager(cfg.Proxy.VaultURL, cfg.Proxy.VaultToken, cfg.Proxy.ClientID)
	s.filter = NewToolFilter(FilterMode(cfg.Proxy.ToolFilter), cfg.Proxy.AllowedTools)

	// 환경변수에서 키 로드 (standalone 모드)
	s.keyMgr.LoadFromEnv()

	// distributed 모드: 금고에서 키 동기화
	if cfg.Proxy.VaultURL != "" {
		s.sse = NewSSEClient(cfg.Proxy.VaultURL, cfg.Proxy.ClientID, func(svc, mdl string) {
			s.mu.Lock()
			if svc != "" {
				s.service = svc
			}
			if mdl != "" {
				s.model = mdl
			}
			s.mu.Unlock()
		})
		s.sse.Start()

		// 금고에서 클라이언트 설정 및 키 초기 로드
		go func() {
			time.Sleep(2 * time.Second)
			s.syncFromVault()
		}()

		// 주기적 키 동기화 (5분마다)
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				s.syncFromVault()
			}
		}()

		// Heartbeat 시작
		s.startHeartbeat()
	}

	// 모델 레지스트리 초기화 (비동기)
	go func() {
		ollamaURL := s.ollamaURL()
		s.registry.Refresh(cfg.Proxy.Services, ollamaURL, "")
	}()

	return s
}

// syncFromVault: 금고에서 클라이언트 설정·키 동기화
func (s *Server) syncFromVault() {
	// 클라이언트 설정 조회
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
			if c.DefaultService != "" {
				s.service = c.DefaultService
			}
			if c.DefaultModel != "" {
				s.model = c.DefaultModel
			}
			s.mu.Unlock()
			log.Printf("[sync] 설정 로드: %s/%s", c.DefaultService, c.DefaultModel)
			break
		}
	}

	// 키 동기화
	if err := s.keyMgr.SyncFromVault(); err != nil {
		log.Printf("[sync] 키 동기화 실패: %v", err)
	}
}

// ollamaURL: 설정·환경변수에서 Ollama URL 반환
func (s *Server) ollamaURL() string {
	if v := os.Getenv("OLLAMA_URL"); v != "" {
		return v
	}
	if v := os.Getenv("WV_OLLAMA_URL"); v != "" {
		return v
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

	// Gemini API
	mux.HandleFunc("/google/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "streamGenerateContent") {
			s.handleGeminiStream(w, r)
		} else {
			s.handleGemini(w, r)
		}
	})

	// OpenAI 호환
	mux.HandleFunc("/v1/chat/completions", s.handleOpenAI)

	return middleware.Chain(mux,
		middleware.Recovery,
		middleware.CORS,
		middleware.Logger,
	)
}

// ─── 헬스 / 상태 ─────────────────────────────────────────────────────────────

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{
		"status":  "ok",
		"version": "v0.1.0",
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
		"version":  "v0.1.0",
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
	if body.Service != "" {
		s.service = body.Service
	}
	if body.Model != "" {
		s.model = body.Model
	}
	s.mu.Unlock()
	log.Printf("[config] 모델 변경: %s/%s", body.Service, body.Model)
	jsonOK(w, map[string]string{"status": "ok", "service": s.service, "model": s.model})
}

func (s *Server) handleThinkMode(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	go s.syncFromVault()
	jsonOK(w, map[string]string{"status": "reloading"})
}

// ─── Gemini API 핸들러 (비스트리밍) ──────────────────────────────────────────

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
		log.Printf("[Security] 요청에서 %d개 도구 차단 (client=%s)", stripped, s.cfg.Proxy.ClientID)
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

// ─── OpenAI API 핸들러 ────────────────────────────────────────────────────────

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

	geminiReq := OpenAIToGemini(&oaiReq)
	geminiResp, err := s.dispatch(svc, mdl, geminiReq)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	oaiResp := &OpenAIResponse{}
	for _, c := range geminiResp.Candidates {
		oaiResp.Choices = append(oaiResp.Choices, OpenAIChoice{
			Message:      OpenAIMessage{Role: "assistant", Content: extractText(c.Content.Parts)},
			FinishReason: strings.ToLower(c.FinishReason),
			Index:        c.Index,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oaiResp)
}

// ─── 요청 분배 ────────────────────────────────────────────────────────────────

func (s *Server) dispatch(service, model string, req *GeminiRequest) (*GeminiResponse, error) {
	var lastErr error
	// 지정 서비스 먼저 시도
	tryOrder := []string{service}
	for _, svc := range s.cfg.Proxy.Services {
		if svc != service {
			tryOrder = append(tryOrder, svc)
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
		default:
			continue
		}
		if err == nil {
			return resp, nil
		}
		log.Printf("[proxy] %s 실패 → 폴백: %v", svc, err)
		lastErr = err
	}
	return nil, fmt.Errorf("모든 서비스 실패: %v", lastErr)
}

// ─── Google Gemini ────────────────────────────────────────────────────────────

func (s *Server) callGoogle(model string, req *GeminiRequest) (*GeminiResponse, error) {
	key, plainKey, err := s.getKey("google")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, plainKey)
	data, _ := json.Marshal(req)
	resp, err := s.doRequest("POST", url, data, nil)
	if err != nil {
		s.keyMgr.RecordError(key, 0)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.keyMgr.RecordError(key, resp.StatusCode)
		return nil, fmt.Errorf("Google API 오류: HTTP %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("Google 응답 파싱 오류: %w", err)
	}
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Google: %s", geminiResp.Error.Message)
	}
	tokens := 0
	if geminiResp.UsageMetadata != nil {
		tokens = geminiResp.UsageMetadata.TotalTokenCount
	}
	s.keyMgr.RecordSuccess(key, tokens)
	return &geminiResp, nil
}

// ─── OpenRouter ───────────────────────────────────────────────────────────────

func (s *Server) callOpenRouter(model string, req *GeminiRequest) (*GeminiResponse, error) {
	key, plainKey, err := s.getKey("openrouter")
	if err != nil {
		return nil, err
	}

	oaiReq := GeminiToOpenAI(model, req)
	data, _ := json.Marshal(oaiReq)
	headers := map[string]string{
		"Authorization": "Bearer " + plainKey,
		"HTTP-Referer":  "https://github.com/sookmook/wall-vault",
		"X-Title":       "wall-vault",
	}
	resp, err := s.doRequest("POST", "https://openrouter.ai/api/v1/chat/completions", data, headers)
	if err != nil {
		s.keyMgr.RecordError(key, 0)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.keyMgr.RecordError(key, resp.StatusCode)
		return nil, fmt.Errorf("OpenRouter 오류: HTTP %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var oaiResp OpenAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("OpenRouter 응답 파싱 오류: %w", err)
	}
	if oaiResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter: %s", oaiResp.Error.Message)
	}
	tokens := 0
	if oaiResp.Usage != nil {
		tokens = oaiResp.Usage.TotalTokens
	}
	s.keyMgr.RecordSuccess(key, tokens)
	return OpenAIRespToGemini(&oaiResp), nil
}

// ─── Ollama (뮤텍스로 단일 동시 요청) ────────────────────────────────────────

func (s *Server) callOllama(model string, req *GeminiRequest) (*GeminiResponse, error) {
	if model == "" {
		model = "qwen3.5:35b"
	}
	ollamaURL := s.ollamaURL()

	s.ollamaMu.Lock()
	defer s.ollamaMu.Unlock()

	ollamaReq := GeminiToOllama(model, req)
	data, _ := json.Marshal(ollamaReq)

	resp, err := s.doRequest("POST", ollamaURL+"/api/chat", data, nil)
	if err != nil {
		return nil, fmt.Errorf("Ollama 연결 실패 (%s): %w", ollamaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama 오류: HTTP %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("Ollama 응답 파싱 오류: %w", err)
	}
	return OllamaRespToGemini(&ollamaResp), nil
}

// ─── 공통 HTTP 요청 ──────────────────────────────────────────────────────────

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

// ─── 유틸 ────────────────────────────────────────────────────────────────────

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
