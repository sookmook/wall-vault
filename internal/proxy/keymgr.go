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

// ─── Cooldown Durations ───────────────────────────────────────────────────────

var cooldownDurations = map[int]time.Duration{
	429: 30 * time.Minute, // rate limit — retry later
	402: 1 * time.Hour,    // payment required — retry in an hour (was 24h)
	401: 24 * time.Hour,   // invalid key — retire for a day
	403: 24 * time.Hour,   // forbidden — retire for a day
	// 400: bad request — request format error, not a key error, no cooldown
	// 404: model not found — key not at fault, no cooldown
	400: 0,
	404: 0,
}

func cooldownFor(errCode int) time.Duration {
	if d, ok := cooldownDurations[errCode]; ok {
		return d
	}
	return 10 * time.Minute
}

// ─── Local Key ────────────────────────────────────────────────────────────────

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

// ─── KeyManager ───────────────────────────────────────────────────────────────

type KeyManager struct {
	mu       sync.Mutex
	keys     map[string][]*localKey
	idx      map[string]int // round-robin index per service
	lastUsed map[string]string // service → last successful key ID
	vaultURL string
	token    string
	clientID string
}

func NewKeyManager(vaultURL, token, clientID string) *KeyManager {
	return &KeyManager{
		keys:     make(map[string][]*localKey),
		idx:      make(map[string]int),
		lastUsed: make(map[string]string),
		vaultURL: vaultURL,
		token:    token,
		clientID: clientID,
	}
}

// LastUsedID: return the ID of the last successfully used key for a service
func (km *KeyManager) LastUsedID(service string) string {
	km.mu.Lock()
	defer km.mu.Unlock()
	return km.lastUsed[service]
}

// AddKey: directly add a key
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

// Get: return an available key (round-robin)
// cycles from the index after the last used one to find an available key
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

// CooldownSnapshot: return cooldownUntil per key ID for keys currently on cooldown
func (km *KeyManager) CooldownSnapshot() map[string]string {
	km.mu.Lock()
	defer km.mu.Unlock()
	snap := make(map[string]string)
	now := time.Now()
	for _, keys := range km.keys {
		for _, k := range keys {
			if k.cooldownUntil.After(now) {
				snap[k.id] = k.cooldownUntil.UTC().Format(time.RFC3339)
			}
		}
	}
	return snap
}

// UsageSnapshot: return current todayUsage per key ID (for heartbeat reporting)
func (km *KeyManager) UsageSnapshot() map[string]int {
	km.mu.Lock()
	defer km.mu.Unlock()
	snap := make(map[string]int)
	for _, keys := range km.keys {
		for _, k := range keys {
			if k.todayUsage > 0 {
				snap[k.id] = k.todayUsage
			}
		}
	}
	return snap
}

// RecordSuccess: record usage and track last-used key per service
func (km *KeyManager) RecordSuccess(k *localKey, tokens int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	k.todayUsage += tokens
	km.lastUsed[k.service] = k.id
}

// RecordError: set cooldown (0 duration = no cooldown, request-side error).
// Rate-limit errors (429, 402) also increment todayUsage by 1 so the dashboard
// shows that the key was attempted even when no successful tokens were returned.
func (km *KeyManager) RecordError(k *localKey, errCode int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	if errCode == http.StatusTooManyRequests || errCode == http.StatusPaymentRequired {
		k.todayUsage += 1
	}
	d := cooldownFor(errCode)
	if d == 0 {
		log.Printf("[key] 쿨다운 없음: service=%s, 오류=%d (request error)", k.service, errCode)
		return
	}
	k.cooldownUntil = time.Now().Add(d)
	log.Printf("[key] 쿨다운 설정: service=%s, 오류=%d, %.0f분", k.service, errCode, d.Minutes())
}

// ─── Load from Environment Variables ─────────────────────────────────────────
// format: WV_KEY_GOOGLE=AIza...         (single key)
//         WV_KEY_GOOGLE=AIza...,AIzb... (comma-separated multiple keys)
//         WV_KEY_GOOGLE=AIza...:500,AIzb...:1500 (key:daily_limit)

func (km *KeyManager) LoadFromEnv() {
	serviceEnvMap := map[string]string{
		"google":      "WV_KEY_GOOGLE",
		"openrouter":  "WV_KEY_OPENROUTER",
		"ollama":      "", // Ollama requires no key
	}
	// legacy compatibility
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
			// try legacy env var
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

// SyncFromVault: sync keys from vault (/api/keys endpoint)
// vault returns decrypted keys (proxy does not need the master password)
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
		ID            string    `json:"id"`
		Service       string    `json:"service"`
		PlainKey      string    `json:"plain_key"`
		DailyLimit    int       `json:"daily_limit"`
		TodayUsage    int       `json:"today_usage"`
		CooldownUntil time.Time `json:"cooldown_until"`
	}
	if err := json.Unmarshal(body, &keys); err != nil {
		return fmt.Errorf("금고 키 파싱 오류: %w", err)
	}

	km.mu.Lock()
	// preserve locally accumulated usage before replacing vault-sourced keys
	localUsage := make(map[string]int)
	for _, svcKeys := range km.keys {
		for _, k := range svcKeys {
			if !strings.HasPrefix(k.id, "env-") && k.todayUsage > 0 {
				localUsage[k.id] = k.todayUsage
			}
		}
	}
	// replace only vault-sourced keys (env var keys are kept)
	for svc := range km.keys {
		var kept []*localKey
		for _, k := range km.keys[svc] {
			if strings.HasPrefix(k.id, "env-") {
				kept = append(kept, k)
			}
		}
		km.keys[svc] = kept
	}
	now := time.Now()
	for _, k := range keys {
		if k.PlainKey == "" {
			continue
		}
		// use the higher of vault usage vs. locally accumulated usage
		usage := k.TodayUsage
		if local := localUsage[k.ID]; local > usage {
			usage = local
		}
		lk := &localKey{
			id:         k.ID,
			service:    k.Service,
			plaintext:  k.PlainKey,
			dailyLimit: k.DailyLimit,
			todayUsage: usage,
		}
		// restore cooldown only if still in the future
		if k.CooldownUntil.After(now) {
			lk.cooldownUntil = k.CooldownUntil
		}
		km.keys[k.Service] = append(km.keys[k.Service], lk)
	}
	km.mu.Unlock()

	log.Printf("[sync] 금고에서 %d개 키 동기화 완료", len(keys))
	return nil
}
