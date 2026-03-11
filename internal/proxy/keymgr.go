package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ─── 쿨다운 시간 ─────────────────────────────────────────────────────────────

var cooldownDurations = map[int]time.Duration{
	429: 30 * time.Minute,
	400: 24 * time.Hour,
	401: 24 * time.Hour,
	402: 24 * time.Hour,
	403: 24 * time.Hour,
}

func cooldownFor(errCode int) time.Duration {
	if d, ok := cooldownDurations[errCode]; ok {
		return d
	}
	return 10 * time.Minute // 네트워크 오류 등
}

// ─── 로컬 키 항목 ─────────────────────────────────────────────────────────────

type localKey struct {
	id            string
	service       string
	plaintext     string
	todayUsage    int
	dailyLimit    int
	cooldownUntil time.Time
}

func (k *localKey) isAvailable() bool {
	if time.Now().Before(k.cooldownUntil) {
		return false
	}
	if k.dailyLimit > 0 && k.todayUsage >= k.dailyLimit {
		return false
	}
	return true
}

// ─── KeyManager ──────────────────────────────────────────────────────────────

// KeyManager: 로컬 키 풀 관리 (볼트에서 받아오거나 config에서 직접 로드)
type KeyManager struct {
	mu       sync.Mutex
	keys     map[string][]*localKey // service → keys
	vaultURL string
	token    string
	clientID string
}

func NewKeyManager(vaultURL, token, clientID string) *KeyManager {
	return &KeyManager{
		keys:     make(map[string][]*localKey),
		vaultURL: vaultURL,
		token:    token,
		clientID: clientID,
	}
}

// AddKey: 직접 키 추가 (standalone 모드)
func (km *KeyManager) AddKey(service, id, plaintext string, dailyLimit int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys[service] = append(km.keys[service], &localKey{
		id:         id,
		service:    service,
		plaintext:  plaintext,
		dailyLimit: dailyLimit,
	})
}

// Get: 서비스에서 사용 가능한 키 반환 (라운드 로빈)
func (km *KeyManager) Get(service string) (*localKey, error) {
	km.mu.Lock()
	defer km.mu.Unlock()
	keys := km.keys[service]
	for _, k := range keys {
		if k.isAvailable() {
			return k, nil
		}
	}
	return nil, fmt.Errorf("서비스 '%s' 사용 가능한 키 없음", service)
}

// RecordSuccess: 성공 후 사용량 기록
func (km *KeyManager) RecordSuccess(k *localKey, tokens int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	k.todayUsage += tokens
	// 볼트에 비동기 보고
	if km.vaultURL != "" {
		go km.reportToVault(k)
	}
}

// RecordError: 오류 후 쿨다운 설정
func (km *KeyManager) RecordError(k *localKey, errCode int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	d := cooldownFor(errCode)
	k.cooldownUntil = time.Now().Add(d)
}

func (km *KeyManager) reportToVault(k *localKey) {
	if km.vaultURL == "" {
		return
	}
	// 볼트에 사용량 보고 (심플하게 heartbeat로 대체)
}

// SyncFromVault: 볼트에서 키 목록 동기화 (distributed 모드)
func (km *KeyManager) SyncFromVault() error {
	if km.vaultURL == "" {
		return nil
	}
	req, err := http.NewRequest("GET", km.vaultURL+"/admin/keys", nil)
	if err != nil {
		return err
	}
	if km.token != "" {
		req.Header.Set("Authorization", "Bearer "+km.token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var keys []struct {
		ID         string `json:"id"`
		Service    string `json:"service"`
		PlainKey   string `json:"plain_key"` // 볼트가 복호화해서 제공
		DailyLimit int    `json:"daily_limit"`
	}
	if err := json.Unmarshal(body, &keys); err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	// 기존 키 초기화 후 재동기화
	km.keys = make(map[string][]*localKey)
	for _, k := range keys {
		if k.PlainKey == "" {
			continue
		}
		km.keys[k.Service] = append(km.keys[k.Service], &localKey{
			id:         k.ID,
			service:    k.Service,
			plaintext:  k.PlainKey,
			dailyLimit: k.DailyLimit,
		})
	}
	return nil
}

// LoadFromEnv: 환경변수/설정에서 키 로드 (standalone 모드)
// 형식: WV_KEY_GOOGLE="AIza...,AIza..." (쉼표 구분 복수 키)
func (km *KeyManager) LoadFromEnv() {
	services := []string{"google", "openrouter", "ollama"}
	for _, svc := range services {
		envKey := fmt.Sprintf("WV_KEY_%s", strings.ToUpper(svc))
		// TODO: os.Getenv(envKey) 처리
		_ = envKey
	}
}
