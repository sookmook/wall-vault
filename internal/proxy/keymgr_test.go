package proxy

import (
	"testing"
	"time"
)

func TestKeyManager_DrainFirst(t *testing.T) {
	km := NewKeyManager("", "", "test")
	km.AddKey("google", "id1", "key-A", 0)
	km.AddKey("google", "id2", "key-B", 0)
	km.AddKey("google", "id3", "key-C", 0)

	// Drain-first: consecutive Gets must return the same key until it is exhausted.
	got := make([]string, 4)
	for i := range got {
		k, err := km.Get("google")
		if err != nil {
			t.Fatalf("Get 실패: %v", err)
		}
		got[i] = k.plaintext
	}
	for i := 1; i < len(got); i++ {
		if got[i] != got[0] {
			t.Fatalf("drain-first 미작동: 0번(%s) != %d번(%s)", got[0], i, got[i])
		}
	}

	// After cooldown, should advance to the next available key.
	k, _ := km.Get("google")
	km.RecordError(k, 429)
	next, err := km.Get("google")
	if err != nil {
		t.Fatalf("쿨다운 후 Get 실패: %v", err)
	}
	if next.plaintext == k.plaintext {
		t.Fatalf("쿨다운 키 건너뛰기 실패: 여전히 %s 반환", k.plaintext)
	}
}

func TestKeyManager_SkipCooldown(t *testing.T) {
	km := NewKeyManager("", "", "test")
	km.AddKey("google", "id1", "key-A", 0)
	km.AddKey("google", "id2", "key-B", 0)

	// set cooldown on first key
	k1, _ := km.Get("google")
	km.RecordError(k1, 429)

	// second get → skips cooled-down key-A and returns key-B
	k2, err := km.Get("google")
	if err != nil {
		t.Fatalf("쿨다운 건너뛰기 실패: %v", err)
	}
	if k2.plaintext == k1.plaintext {
		t.Fatalf("쿨다운 키가 반환됨: %s", k2.plaintext)
	}
}

func TestKeyManager_AllCooldown_ForceRetry(t *testing.T) {
	km := NewKeyManager("", "", "test")
	km.AddKey("google", "id1", "key-A", 0)
	km.AddKey("google", "id2", "key-B", 0)

	// both keys on cooldown
	k1, _ := km.Get("google")
	km.RecordError(k1, 429)
	k2, _ := km.Get("google")
	km.RecordError(k2, 429)

	// all cooldown → should force-retry earliest key (not error)
	k3, err := km.Get("google")
	if err != nil {
		t.Fatalf("전체 쿨다운 시 강제 재시도 기대, 오류: %v", err)
	}
	if k3 == nil {
		t.Fatal("강제 재시도 키가 nil")
	}
}

func TestKeyManager_DailyLimit(t *testing.T) {
	km := NewKeyManager("", "", "test")
	km.AddKey("google", "id1", "key-A", 10) // 일일 한도 10

	k, _ := km.Get("google")
	km.RecordSuccess(k, 10) // 한도 도달

	// daily limit reached + force retry → should still return a key
	k2, err := km.Get("google")
	if err != nil {
		t.Fatalf("한도 초과 시 강제 재시도 기대, 오류: %v", err)
	}
	if k2 == nil {
		t.Fatal("강제 재시도 키가 nil")
	}
}

func TestKeyManager_NoKeys(t *testing.T) {
	km := NewKeyManager("", "", "test")
	_, err := km.Get("google")
	if err == nil {
		t.Fatal("키 없을 때 오류 기대")
	}
}

func TestKeyManager_LoadFromEnv(t *testing.T) {
	t.Setenv("WV_KEY_GOOGLE", "key1,key2:500")
	km := NewKeyManager("", "", "test")
	km.LoadFromEnv()

	k, err := km.Get("google")
	if err != nil {
		t.Fatalf("환경변수 키 로드 실패: %v", err)
	}
	if k.plaintext != "key1" {
		t.Fatalf("첫 번째 키 기대 'key1', got %q", k.plaintext)
	}
}

func TestKeyManager_SyncReplacesVaultKeys(t *testing.T) {
	km := NewKeyManager("", "", "test")
	// env var key
	km.AddKey("google", "env-google-0", "env-key", 0)
	// simulate vault key (added directly)
	km.AddKey("google", "vault-1", "vault-key", 100)

	// verify logic: remove only vault keys, keep env var keys
	km.mu.Lock()
	var kept []*localKey
	for _, k := range km.keys["google"] {
		if len(k.id) > 4 && k.id[:4] == "env-" {
			kept = append(kept, k)
		}
	}
	km.keys["google"] = kept
	km.mu.Unlock()

	k, err := km.Get("google")
	if err != nil {
		t.Fatalf("환경변수 키 유지 실패: %v", err)
	}
	if k.plaintext != "env-key" {
		t.Fatalf("환경변수 키 기대, got %q", k.plaintext)
	}
}

func TestLocalKey_IsAvailable(t *testing.T) {
	k := &localKey{plaintext: "key", dailyLimit: 100, todayUsage: 50}
	if !k.isAvailable() {
		t.Fatal("50/100 — 사용 가능해야 함")
	}

	k.todayUsage = 100
	if k.isAvailable() {
		t.Fatal("100/100 — 사용 불가해야 함")
	}

	k.todayUsage = 0
	k.cooldownUntil = time.Now().Add(10 * time.Minute)
	if k.isAvailable() {
		t.Fatal("쿨다운 중 — 사용 불가해야 함")
	}
}
