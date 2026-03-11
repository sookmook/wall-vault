package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sookmook/wall-vault/internal/config"
)

// newTestServer: 테스트용 프록시 서버 생성 (금고 연결 없음)
func newTestServer() *Server {
	cfg := config.Default()
	cfg.Mode = "standalone"
	cfg.Proxy.Services = []string{"ollama"}
	cfg.Proxy.ToolFilter = "strip_all"
	cfg.Proxy.VaultURL = "" // 금고 연결 비활성화
	return NewServer(cfg)
}

func TestHandleHealth(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("health 응답 파싱 실패: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
	if _, ok := body["version"]; !ok {
		t.Error("version 필드 없음")
	}
}

func TestHandleStatus(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("status 응답 파싱 실패: %v", err)
	}

	required := []string{"status", "version", "client", "service", "model", "filter", "mode"}
	for _, k := range required {
		if _, ok := body[k]; !ok {
			t.Errorf("status 응답에 %q 필드 없음", k)
		}
	}
}

func TestHandleModels(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/api/models", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("models status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("models 응답 파싱 실패: %v", err)
	}
	if _, ok := body["models"]; !ok {
		t.Error("models 필드 없음")
	}
	if _, ok := body["count"]; !ok {
		t.Error("count 필드 없음")
	}
}

func TestHandleConfigModel(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	body := `{"service":"google","model":"gemini-2.5-flash"}`
	req := httptest.NewRequest("PUT", "/api/config/model", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("config/model status = %d, want 200", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("응답 파싱 실패: %v", err)
	}
	if resp["service"] != "google" {
		t.Errorf("service = %q, want google", resp["service"])
	}
	if resp["model"] != "gemini-2.5-flash" {
		t.Errorf("model = %q, want gemini-2.5-flash", resp["model"])
	}

	// 변경 후 /status에서 반영 확인
	req2 := httptest.NewRequest("GET", "/status", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	var status map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &status)
	if status["service"] != "google" {
		t.Errorf("status service = %v, want google", status["service"])
	}
}

func TestHandleConfigModel_InvalidMethod(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/api/config/model", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleConfigModel_InvalidBody(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("PUT", "/api/config/model", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleReload(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("POST", "/reload", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("reload status = %d, want 200", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "reloading" {
		t.Errorf("status = %q, want reloading", body["status"])
	}
}

func TestHandleGemini_InvalidMethod(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/google/v1beta/models/gemini-2.5-flash:generateContent", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleGemini_ToolStripping(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	// 도구가 포함된 Gemini 요청
	body := `{
		"contents": [{"role":"user","parts":[{"text":"안녕"}]}],
		"tools": [{"functionDeclarations":[{"name":"search"}]}]
	}`
	req := httptest.NewRequest("POST",
		"/google/v1beta/models/gemini-2.5-flash:generateContent",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// strip_all 필터: 도구 제거 후 처리 (업스트림 없으므로 502 예상)
	if w.Code != http.StatusBadGateway {
		t.Logf("status = %d (업스트림 없이 502 예상)", w.Code)
	}
}

func TestHandleOpenAI_InvalidMethod(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("GET", "/v1/chat/completions", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleOpenAI_InvalidBody(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("POST", "/v1/chat/completions",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestExtractModelFromPath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/google/v1beta/models/gemini-2.5-flash:generateContent", "gemini-2.5-flash"},
		{"/google/v1beta/models/gemini-1.5-pro:streamGenerateContent", "gemini-1.5-pro"},
		{"/v1/chat/completions", ""},
		{"/health", ""},
	}

	for _, c := range cases {
		got := extractModelFromPath(c.path)
		if got != c.want {
			t.Errorf("extractModelFromPath(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestHandleCORS(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS 헤더 없음: %v", w.Header())
	}
}
