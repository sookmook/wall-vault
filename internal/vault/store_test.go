package vault

import (
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

// ─── 키 CRUD ─────────────────────────────────────────────────────────────────

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

// ─── 클라이언트 CRUD ──────────────────────────────────────────────────────────

func TestStore_AddUpdateDeleteClient(t *testing.T) {
	s := newTestStore(t)

	c, err := s.AddClient("bot1", "봇1", "token-abc", "google", "gemini-2.5-flash", nil)
	if err != nil {
		t.Fatalf("AddClient 실패: %v", err)
	}
	if c.ID != "bot1" {
		t.Fatalf("ID 불일치: %q", c.ID)
	}

	if err := s.UpdateClient("bot1", "openrouter", "qwen3.5:35b", nil); err != nil {
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
	s.AddClient("bot1", "봇1", "secret-token-xyz", "google", "", nil)

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
	if err := s.UpdateClient("nonexistent", "google", "", nil); err == nil {
		t.Fatal("없는 클라이언트 업데이트 — 오류 기대")
	}
}

// ─── 영속화 ───────────────────────────────────────────────────────────────────

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()

	// 저장
	s1, _ := NewStore(dir, "password123")
	s1.AddKey("google", "secret-key", "my-key", 500)
	s1.AddClient("bot", "봇", "tok", "google", "gemini-2.5-flash", nil)

	// 재로드
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

	// 암호화된 키 복호화 확인
	_, plain, err := s2.GetAvailableKey("google")
	if err != nil {
		t.Fatalf("재로드 후 키 조회 실패: %v", err)
	}
	if plain != "secret-key" {
		t.Fatalf("복호화 불일치: %q", plain)
	}
}

// ─── 프록시 상태 ──────────────────────────────────────────────────────────────

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
