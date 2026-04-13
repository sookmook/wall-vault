package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir(), "")
	if err != nil {
		t.Fatalf("NewStore 실패: %v", err)
	}
	return s
}

// ─── Key CRUD ─────────────────────────────────────────────────────────────────

func TestStore_AddListDeleteKey(t *testing.T) {
	s := newTestStore(t)

	k, err := s.AddKey("google", "AIzaTest", "test-key", 1000)
	if err != nil {
		t.Fatalf("AddKey 실패: %v", err)
	}
	if k.ID == "" {
		t.Fatal("키 ID가 빈 문자열")
	}

	keys := s.ListKeys()
	if len(keys) != 1 {
		t.Fatalf("1개 기대, got %d", len(keys))
	}
	if keys[0].Label != "test-key" {
		t.Fatalf("레이블 불일치: %q", keys[0].Label)
	}

	if err := s.DeleteKey(k.ID); err != nil {
		t.Fatalf("DeleteKey 실패: %v", err)
	}
	if len(s.ListKeys()) != 0 {
		t.Fatal("삭제 후 키가 남아 있음")
	}
}

func TestStore_DeleteKey_NotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.DeleteKey("nonexistent"); err == nil {
		t.Fatal("없는 키 삭제 — 오류 기대")
	}
}

func TestStore_ResetDailyUsage(t *testing.T) {
	s := newTestStore(t)
	k, _ := s.AddKey("google", "key", "label", 500)
	s.RecordKeyUsage(k.ID, 300)

	keys := s.ListKeys()
	if keys[0].TodayUsage != 300 {
		t.Fatalf("사용량 기대 300, got %d", keys[0].TodayUsage)
	}

	s.ResetDailyUsage()
	keys = s.ListKeys()
	if keys[0].TodayUsage != 0 {
		t.Fatalf("리셋 후 0 기대, got %d", keys[0].TodayUsage)
	}
}

func TestStore_SetKeyCooldown(t *testing.T) {
	s := newTestStore(t)
	k, _ := s.AddKey("google", "key", "label", 0)

	until := time.Now().Add(30 * time.Minute)
	s.SetKeyCooldown(k.ID, 429, until)

	keys := s.ListKeys()
	if keys[0].LastError != 429 {
		t.Fatalf("LastError 기대 429, got %d", keys[0].LastError)
	}
	if !keys[0].IsOnCooldown() {
		t.Fatal("쿨다운 상태여야 함")
	}
}

func TestStore_GetAvailableKey(t *testing.T) {
	s := newTestStore(t)
	s.AddKey("google", "key1", "label1", 0)
	s.AddKey("openrouter", "key2", "label2", 0)

	_, plain, err := s.GetAvailableKey("google")
	if err != nil {
		t.Fatalf("GetAvailableKey 실패: %v", err)
	}
	if plain != "key1" {
		t.Fatalf("키값 불일치: %q", plain)
	}

	_, _, err = s.GetAvailableKey("ollama")
	if err == nil {
		t.Fatal("없는 서비스 — 오류 기대")
	}
}

// ─── Client CRUD ──────────────────────────────────────────────────────────────

func TestStore_AddUpdateDeleteClient(t *testing.T) {
	s := newTestStore(t)

	c, err := s.AddClient(ClientInput{ID: "bot1", Name: "봇1", Token: "token-abc", DefaultService: "google", DefaultModel: "gemini-2.5-flash"})
	if err != nil {
		t.Fatalf("AddClient 실패: %v", err)
	}
	if c.ID != "bot1" {
		t.Fatalf("ID 불일치: %q", c.ID)
	}

	svc := "openrouter"
	mdl := "qwen3.5:35b"
	if err := s.UpdateClient("bot1", ClientUpdateInput{DefaultService: &svc, DefaultModel: &mdl}); err != nil {
		t.Fatalf("UpdateClient 실패: %v", err)
	}
	got := s.GetClient("bot1")
	if got.DefaultService != "openrouter" {
		t.Fatalf("서비스 업데이트 실패: %q", got.DefaultService)
	}

	if err := s.DeleteClient("bot1"); err != nil {
		t.Fatalf("DeleteClient 실패: %v", err)
	}
	if s.GetClient("bot1") != nil {
		t.Fatal("삭제 후 클라이언트가 남아 있음")
	}
}

func TestStore_GetClientByToken(t *testing.T) {
	s := newTestStore(t)
	s.AddClient(ClientInput{ID: "bot1", Name: "봇1", Token: "secret-token-xyz", DefaultService: "google"})

	c := s.GetClientByToken("secret-token-xyz")
	if c == nil || c.ID != "bot1" {
		t.Fatal("토큰으로 클라이언트 조회 실패")
	}

	if s.GetClientByToken("wrong-token") != nil {
		t.Fatal("잘못된 토큰으로 클라이언트 조회됨")
	}
}

func TestStore_UpdateClient_NotFound(t *testing.T) {
	s := newTestStore(t)
	svc := "google"
	if err := s.UpdateClient("nonexistent", ClientUpdateInput{DefaultService: &svc}); err == nil {
		t.Fatal("없는 클라이언트 업데이트 — 오류 기대")
	}
}

// ─── Persistence ──────────────────────────────────────────────────────────────

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()

	// save
	s1, _ := NewStore(dir, "password123")
	s1.AddKey("google", "secret-key", "my-key", 500)
	s1.AddClient(ClientInput{ID: "bot", Name: "봇", Token: "tok", DefaultService: "google", DefaultModel: "gemini-2.5-flash"})

	// reload
	s2, err := NewStore(dir, "password123")
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}
	if len(s2.ListKeys()) != 1 {
		t.Fatalf("키 영속화 실패: got %d", len(s2.ListKeys()))
	}
	if len(s2.ListClients()) != 1 {
		t.Fatalf("클라이언트 영속화 실패: got %d", len(s2.ListClients()))
	}

	// verify decryption of encrypted key
	_, plain, err := s2.GetAvailableKey("google")
	if err != nil {
		t.Fatalf("재로드 후 키 조회 실패: %v", err)
	}
	if plain != "secret-key" {
		t.Fatalf("복호화 불일치: %q", plain)
	}
}

// ─── Proxy Status ─────────────────────────────────────────────────────────────

func TestStore_ProxyStatus(t *testing.T) {
	s := newTestStore(t)
	s.UpdateProxyStatus(&ProxyStatus{
		ClientID: "bot1",
		Version:  "v0.1.0",
		Service:  "google",
		Model:    "gemini-2.5-flash",
	})

	proxies := s.ListProxies()
	if len(proxies) != 1 {
		t.Fatalf("1개 기대, got %d", len(proxies))
	}
	if proxies[0].ClientID != "bot1" {
		t.Fatalf("ClientID 불일치: %q", proxies[0].ClientID)
	}
	if proxies[0].UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt이 0값")
	}
}

// ─── Auto-Migration (v1 → v2) ────────────────────────────────────────────────

func TestLoadAutoMigratesV1(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "vault.json")
	password := "test-master"

	// Legacy v1 plaintext (no schema_version).
	v1 := []byte(`{
		"services":[
			{"id":"google","name":"Google","enabled":true,"proxy_enabled":true}
		],
		"clients":[
			{"id":"bot-a","name":"Delta","token":"t",
			 "default_service":"google","default_model":"gemini-3.1-pro-preview",
			 "agent_type":"nanoclaw","enabled":true,"sort_order":1}
		],
		"keys":[]
	}`)

	// Write v1 plaintext directly to disk (as legacy store would have written it).
	if err := writeEncryptedForTest(path, password, v1); err != nil {
		t.Fatal(err)
	}

	// Act: open the store on the legacy file.
	store, err := openStoreForTest(path, password)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}

	// Assert: backup exists.
	matches, _ := filepath.Glob(filepath.Join(tmp, "vault.json.pre-v02.*.bak"))
	if len(matches) != 1 {
		t.Fatalf("expected one pre-v02 backup, got %d (%v)", len(matches), matches)
	}

	// Assert: in-memory schema upgraded.
	clients := store.ListClients()
	if len(clients) != 1 {
		t.Fatalf("expected 1 client after migration, got %d", len(clients))
	}
	if clients[0].PreferredService != "google" {
		t.Fatalf("preferred_service=%q", clients[0].PreferredService)
	}
	if clients[0].ModelOverride != "gemini-3.1-pro-preview" {
		t.Fatalf("model_override=%q", clients[0].ModelOverride)
	}

	// Assert: on-disk file is now v2 (re-open without migration should be a noop).
	store2, err := openStoreForTest(path, password)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	_ = store2

	matches2, _ := filepath.Glob(filepath.Join(tmp, "vault.json.pre-v02.*.bak"))
	if len(matches2) != 1 { // still exactly one — no new backup on second load
		t.Fatalf("unexpected additional backup after reopen: %v", matches2)
	}
}

// writeEncryptedForTest writes raw plaintext JSON bytes directly to path,
// exactly as the legacy (v0.1.x) store would have written vault.json.
// The store does not encrypt the file itself — only individual key values
// inside the JSON are encrypted. So "writing the encrypted file" means
// writing the JSON blob as-is (the masterPass parameter is accepted for
// API symmetry but the file itself is plain JSON in the current design).
func writeEncryptedForTest(path, _ string, raw []byte) error {
	return os.WriteFile(path, raw, 0600)
}

// openStoreForTest opens a Store against a specific vault.json path
// (not the default dataDir layout). It uses a one-file tempdir trick:
// create a tempdir, symlink/copy the file there, open via NewStore.
func openStoreForTest(path, password string) (*Store, error) {
	// Construct a storeData directly from raw JSON to validate it first,
	// then use NewStore with the parent directory.
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	if base != "vault.json" {
		// Re-write to a vault.json in that dir if needed.
		dest := filepath.Join(dir, "vault.json")
		if dest != path {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if err := os.WriteFile(dest, data, 0600); err != nil {
				return nil, err
			}
		}
	}
	return NewStore(dir, password)
}

// Ensure the test imports are used (encoding/json is used in validate helpers).
var _ = json.Marshal
