package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	return 10 * time.Minute
}

// ─── 로컬 키 ─────────────────────────────────────────────────────────────────

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

type KeyManager struct {
	mu       sync.Mutex
	keys     map[string][]*localKey
	idx      map[string]int // 서비스별 라운드 로빈 인덱스
	vaultURL string
	token    string
	clientID string
}

func NewKeyManager(vaultURL, token, clientID string) *KeyManager {
	return &KeyManager{
		keys:     make(map[string][]*localKey),
		idx:      make(map[string]int),
		vaultURL: vaultURL,
		token:    token,
		clientID: clientID,
	}
}

// AddKey: 직접 키 추가
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

// Get: 사용 가능한 키 반환 (라운드 로빈)
// 마지막으로 사용한 인덱스 다음부터 순환하여 가용 키를 반환
func (km *KeyManager) Get(service string) (*localKey, error) {
	km.mu.Lock()
	defer km.mu.Unlock()
	keys := km.keys[service]
	n := len(keys)
	if n == 0 {
		return nil, fmt.Errorf("서비스 '%s' 등록된 키 없음", service)
	}
	start := km.idx[service] % n
	for i := 0; i < n; i++ {
		k := keys[(start+i)%n]
		if k.isAvailable() {
			km.idx[service] = (start + i + 1) % n
			return k, nil
		}
	}
	return nil, fmt.Errorf("서비스 '%s' 사용 가능한 키 없음 (등록된 키 %d개, 모두 쿨다운/소진)", service, n)
}

// RecordSuccess: 사용량 기록
func (km *KeyManager) RecordSuccess(k *localKey, tokens int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	k.todayUsage += tokens
}

// RecordError: 쿨다운 설정
func (km *KeyManager) RecordError(k *localKey, errCode int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	d := cooldownFor(errCode)
	k.cooldownUntil = time.Now().Add(d)
	log.Printf("[key] 쿨다운 설정: service=%s, 오류=%d, %.0f분", k.service, errCode, d.Minutes())
}

// ─── 환경변수에서 로드 ────────────────────────────────────────────────────────
// 형식: WV_KEY_GOOGLE=AIza...         (단일 키)
//       WV_KEY_GOOGLE=AIza...,AIzb... (쉼표 구분 복수 키)
//       WV_KEY_GOOGLE=AIza...:500,AIzb...:1500 (키:일일한도)

func (km *KeyManager) LoadFromEnv() {
	serviceEnvMap := map[string]string{
		"google":      "WV_KEY_GOOGLE",
		"openrouter":  "WV_KEY_OPENROUTER",
		"ollama":      "", // Ollama는 키 불필요
	}
	// 레거시 호환
	legacyEnvMap := map[string]string{
		"google":     "GOOGLE_API_KEY",
		"openrouter": "OPENROUTER_API_KEY",
	}

	loaded := 0
	for svc, envName := range serviceEnvMap {
		if envName == "" {
			continue
		}
		val := os.Getenv(envName)
		if val == "" {
			// 레거시 환경변수 시도
			if legacy, ok := legacyEnvMap[svc]; ok {
				val = os.Getenv(legacy)
			}
		}
		if val == "" {
			continue
		}

		for _, entry := range strings.Split(val, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			var key string
			var limit int
			if parts := strings.SplitN(entry, ":", 2); len(parts) == 2 {
				key = parts[0]
				fmt.Sscanf(parts[1], "%d", &limit)
			} else {
				key = entry
			}
			if key != "" {
				km.AddKey(svc, fmt.Sprintf("env-%s-%d", svc, loaded), key, limit)
				loaded++
				log.Printf("[key] 환경변수에서 로드: service=%s", svc)
			}
		}
	}
}

// SyncFromVault: 금고에서 키 동기화 (/api/keys 엔드포인트)
// 금고가 복호화된 키를 반환 (프록시는 마스터 비밀번호 불필요)
func (km *KeyManager) SyncFromVault() error {
	if km.vaultURL == "" {
		return nil
	}

	url := km.vaultURL + "/api/keys"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if km.token != "" {
		req.Header.Set("Authorization", "Bearer "+km.token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("금고 키 조회 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("금고 인증 실패 — vault_token 확인 필요")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("금고 키 조회: HTTP %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var keys []struct {
		ID         string `json:"id"`
		Service    string `json:"service"`
		PlainKey   string `json:"plain_key"`
		DailyLimit int    `json:"daily_limit"`
	}
	if err := json.Unmarshal(body, &keys); err != nil {
		return fmt.Errorf("금고 키 파싱 오류: %w", err)
	}

	km.mu.Lock()
	// 금고에서 받은 키만 교체 (환경변수 키는 유지)
	for svc := range km.keys {
		var kept []*localKey
		for _, k := range km.keys[svc] {
			if strings.HasPrefix(k.id, "env-") {
				kept = append(kept, k)
			}
		}
		km.keys[svc] = kept
	}
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
	km.mu.Unlock()

	log.Printf("[sync] 금고에서 %d개 키 동기화 완료", len(keys))
	return nil
}
