package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sookmook/wall-vault/internal/config"
)

// newTestVaultServer: test vault server using a temporary directory
func newTestVaultServer(t *testing.T) (*Server, func()) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Vault.DataDir = dir
	cfg.Vault.AdminToken = "test-admin"
	cfg.Vault.MasterPass = ""

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("서버 생성 실패: %v", err)
	}
	cleanup := func() { os.RemoveAll(dir) }
	return srv, cleanup
}

// ─── Public API ───────────────────────────────────────────────────────────────

func TestVaultStatus(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("status field = %v", body["status"])
	}
	if _, ok := body["keys"]; !ok {
		t.Error("keys 필드 없음")
	}
}

func TestVaultPublicClients(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/clients", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	// should be empty array
	var body []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("파싱 실패: %v", err)
	}
}

// ─── Admin Authentication ─────────────────────────────────────────────────────

func TestVaultAdminAuth_Unauthorized(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/admin/keys", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestVaultAdminAuth_WrongToken(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/admin/keys", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ─── Key CRUD ─────────────────────────────────────────────────────────────────

func TestVaultKeyAddAndList(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	// add key
	body := `{"service":"google","key":"AIzaSy-test-key","label":"테스트","daily_limit":100}`
	req := httptest.NewRequest("POST", "/admin/keys", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("키 추가 status = %d, want 200: %s", w.Code, w.Body.String())
	}

	// list keys
	req2 := httptest.NewRequest("GET", "/admin/keys", nil)
	req2.Header.Set("Authorization", "Bearer test-admin")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	var keys []map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &keys); err != nil {
		t.Fatalf("파싱 실패: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("keys 수 = %d, want 1", len(keys))
	}
	if keys[0]["service"] != "google" {
		t.Errorf("service = %v, want google", keys[0]["service"])
	}
	if keys[0]["label"] != "테스트" {
		t.Errorf("label = %v", keys[0]["label"])
	}
	// plaintext key must not be included in /admin/keys response
	if _, ok := keys[0]["plain_key"]; ok {
		t.Error("plain_key가 관리자 목록에 노출됨")
	}
}

func TestVaultKeyAdd_MissingKey(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	body := `{"service":"google"}`  // key 필드 없음
	req := httptest.NewRequest("POST", "/admin/keys", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestVaultKeyDelete(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	// add key
	body := `{"service":"openrouter","key":"sk-or-test","label":"OR 키"}`
	req := httptest.NewRequest("POST", "/admin/keys", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	// delete
	req2 := httptest.NewRequest("DELETE", "/admin/keys/"+id, nil)
	req2.Header.Set("Authorization", "Bearer test-admin")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("삭제 status = %d", w2.Code)
	}

	// verify it is gone from the list
	req3 := httptest.NewRequest("GET", "/admin/keys", nil)
	req3.Header.Set("Authorization", "Bearer test-admin")
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)

	var keys []interface{}
	json.Unmarshal(w3.Body.Bytes(), &keys)
	if len(keys) != 0 {
		t.Errorf("삭제 후 keys 수 = %d, want 0", len(keys))
	}
}

func TestVaultKeyDelete_NotFound(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/admin/keys/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer test-admin")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// ─── Client CRUD ──────────────────────────────────────────────────────────────

func TestVaultClientAddAndList(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	// add client
	body := `{"id":"bot1","name":"봇 원","token":"bot1-token","default_service":"google","default_model":"gemini-2.5-flash"}`
	req := httptest.NewRequest("POST", "/admin/clients", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("클라이언트 추가 status = %d: %s", w.Code, w.Body.String())
	}

	// admin list (includes token)
	req2 := httptest.NewRequest("GET", "/admin/clients", nil)
	req2.Header.Set("Authorization", "Bearer test-admin")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	var clients []map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &clients)
	if len(clients) != 1 {
		t.Fatalf("clients 수 = %d, want 1", len(clients))
	}
	if clients[0]["id"] != "bot1" {
		t.Errorf("id = %v", clients[0]["id"])
	}
	if clients[0]["token"] != "bot1-token" {
		t.Errorf("token = %v", clients[0]["token"])
	}
}

func TestVaultClientUpdate_SSE(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	// register client
	addBody := `{"id":"bot2","name":"봇 투","token":"bot2-token","default_service":"google","default_model":"gemini-2.0-flash"}`
	req := httptest.NewRequest("POST", "/admin/clients", strings.NewReader(addBody))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(httptest.NewRecorder(), req)

	// change model
	upBody := `{"default_service":"openrouter","default_model":"anthropic/claude-3.5-sonnet"}`
	req2 := httptest.NewRequest("PUT", "/admin/clients/bot2", strings.NewReader(upBody))
	req2.Header.Set("Authorization", "Bearer test-admin")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("업데이트 status = %d: %s", w2.Code, w2.Body.String())
	}

	// verify changes
	req3 := httptest.NewRequest("GET", "/admin/clients/bot2", nil)
	req3.Header.Set("Authorization", "Bearer test-admin")
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)

	var c map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &c)
	if c["default_service"] != "openrouter" {
		t.Errorf("default_service = %v, want openrouter", c["default_service"])
	}
}

func TestVaultClientDelete(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	addBody := `{"id":"bot3","name":"봇 셋","token":"bot3-tok"}`
	req := httptest.NewRequest("POST", "/admin/clients", strings.NewReader(addBody))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest("DELETE", "/admin/clients/bot3", nil)
	req2.Header.Set("Authorization", "Bearer test-admin")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("삭제 status = %d", w2.Code)
	}
}

// ─── Proxy-Only API ───────────────────────────────────────────────────────────

func TestVaultProxyKeys_ClientAuth(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	// register client
	addC := `{"id":"proxyclient","name":"프록시","token":"proxy-tok"}`
	req := httptest.NewRequest("POST", "/admin/clients", strings.NewReader(addC))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(httptest.NewRecorder(), req)

	// register key
	addK := `{"service":"google","key":"AIza-proxy-test","daily_limit":0}`
	req2 := httptest.NewRequest("POST", "/admin/keys", strings.NewReader(addK))
	req2.Header.Set("Authorization", "Bearer test-admin")
	req2.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(httptest.NewRecorder(), req2)

	// query /api/keys with client token
	req3 := httptest.NewRequest("GET", "/api/keys", nil)
	req3.Header.Set("Authorization", "Bearer proxy-tok")
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("api/keys status = %d: %s", w3.Code, w3.Body.String())
	}

	var keys []map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &keys)
	if len(keys) == 0 {
		t.Fatal("keys 없음")
	}
	if _, ok := keys[0]["plain_key"]; !ok {
		t.Error("plain_key 필드 없음")
	}
}

func TestVaultProxyKeys_Unauthorized(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ─── Heartbeat ────────────────────────────────────────────────────────────────

func TestVaultHeartbeat(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()
	h := srv.Handler()

	// register client
	addC := fmt.Sprintf(`{"id":"hbclient","name":"HB","token":"hb-tok"}`)
	req := httptest.NewRequest("POST", "/admin/clients", strings.NewReader(addC))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(httptest.NewRecorder(), req)

	// send Heartbeat
	hbBody := `{"client_id":"hbclient","version":"v0.1.0","service":"google","model":"gemini-2.5-flash","sse_connected":true}`
	req2 := httptest.NewRequest("POST", "/api/heartbeat", strings.NewReader(hbBody))
	req2.Header.Set("Authorization", "Bearer hb-tok")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("heartbeat status = %d: %s", w2.Code, w2.Body.String())
	}

	// verify in admin/proxies
	req3 := httptest.NewRequest("GET", "/admin/proxies", nil)
	req3.Header.Set("Authorization", "Bearer test-admin")
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)

	var proxies []map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &proxies)
	if len(proxies) == 0 {
		t.Fatal("proxies 없음")
	}
	if proxies[0]["client_id"] != "hbclient" {
		t.Errorf("client_id = %v", proxies[0]["client_id"])
	}
}

// ─── Usage Reset ──────────────────────────────────────────────────────────────

func TestVaultResetUsage(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/admin/keys/reset", nil)
	req.Header.Set("Authorization", "Bearer test-admin")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("reset status = %d", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "reset" {
		t.Errorf("status = %v, want reset", body["status"])
	}
}

// ─── Dashboard UI ─────────────────────────────────────────────────────────────

func TestVaultDashboard(t *testing.T) {
	srv, cleanup := newTestVaultServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(w.Body.String(), "wall-vault") {
		t.Error("대시보드 HTML에 wall-vault 없음")
	}
}
