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
	429: 5 * time.Minute,  // rate limit — short retry (prevents total lockout)
	402: 30 * time.Minute, // payment required — moderate retry
	401: 6 * time.Hour,    // invalid key — retire
	403: 6 * time.Hour,    // forbidden — retire
	582: 3 * time.Minute,  // upstream overload — very short retry
	400: 0,                // bad request — not a key error
	404: 0,                // model not found — not a key error
}

func cooldownFor(errCode int) time.Duration {
	if d, ok := cooldownDurations[errCode]; ok {
		return d
	}
	return 5 * time.Minute
}

// ─── Local Key ────────────────────────────────────────────────────────────────

type localKey struct {
	id             string
	service        string
	plaintext      string
	todayUsage     int  // successful tokens (or 1 per successful request when token count unavailable)
	todayAttempts  int  // total requests sent to the API (success + rate-limited)
	dailyLimit     int
	cooldownUntil  time.Time
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
	mu           sync.Mutex
	keys         map[string][]*localKey
	idx          map[string]int            // round-robin index per service
	lastUsed     map[string]string         // service → last successful key ID
	vaultURL     string
	token        string
	clientID     string
	lastSyncDate string // "YYYY-MM-DD" — detects midnight rollover to discard stale local counters
}

func NewKeyManager(vaultURL, token, clientID string) *KeyManager {
	return &KeyManager{
		keys:         make(map[string][]*localKey),
		idx:          make(map[string]int),
		lastUsed:     make(map[string]string),
		vaultURL:     vaultURL,
		token:        token,
		clientID:     clientID,
		lastSyncDate: time.Now().UTC().Format("2006-01-02"),
	}
}

// ResetDailyCounters: zero all local today_usage/today_attempts counters immediately.
// Called when vault broadcasts a usage_reset SSE event (midnight rollover).
func (km *KeyManager) ResetDailyCounters() {
	km.mu.Lock()
	defer km.mu.Unlock()
	today := time.Now().UTC().Format("2006-01-02")
	for _, keys := range km.keys {
		for _, k := range keys {
			k.todayUsage = 0
			k.todayAttempts = 0
		}
	}
	km.lastSyncDate = today
	log.Printf("[sync] daily counters reset (date: %s)", today)
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
			km.idx[service] = (start + i) % n
			return k, nil
		}
	}
	// all keys on cooldown — force-retry the key whose cooldown expires soonest
	var earliest *localKey
	for _, k := range keys {
		if earliest == nil || k.cooldownUntil.Before(earliest.cooldownUntil) {
			earliest = k
		}
	}
	if earliest != nil {
		log.Printf("[key] 전체 쿨다운 — 가장 이른 키 강제 재시도: %s (%.0f초 남음)",
			earliest.service, time.Until(earliest.cooldownUntil).Seconds())
		earliest.cooldownUntil = time.Time{} // clear cooldown
		return earliest, nil
	}
	return nil, fmt.Errorf("서비스 '%s' 사용 가능한 키 없음 (등록된 키 %d개)", service, n)
}

// CanServe returns true when the service has at least one key that is not
// on cooldown and not exhausted. Used by dispatch to fast-skip fully-cooled
// cloud services instead of waiting for Get() to force-retry and collect a
// fresh 429/402.
func (km *KeyManager) CanServe(service string) bool {
	km.mu.Lock()
	defer km.mu.Unlock()
	for _, k := range km.keys[service] {
		if k.isAvailable() {
			return true
		}
	}
	return false
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

// UsageSnapshot: return todayUsage per key ID (successful tokens/requests only)
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

// AttemptsSnapshot: return todayAttempts per key ID (all requests including rate-limited)
func (km *KeyManager) AttemptsSnapshot() map[string]int {
	km.mu.Lock()
	defer km.mu.Unlock()
	snap := make(map[string]int)
	for _, keys := range km.keys {
		for _, k := range keys {
			if k.todayAttempts > 0 {
				snap[k.id] = k.todayAttempts
			}
		}
	}
	return snap
}

// RecordSuccess: record successful token usage and count the attempt
func (km *KeyManager) RecordSuccess(k *localKey, tokens int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	k.todayUsage += tokens
	k.todayAttempts += 1
	km.lastUsed[k.service] = k.id
}

// RecordError: set cooldown (0 duration = no cooldown, request-side error).
// Rate-limit and gateway errors (429, 402, 582) count as an attempt so the
// dashboard shows activity even when no successful tokens were returned.
func (km *KeyManager) RecordError(k *localKey, errCode int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	switch errCode {
	case http.StatusTooManyRequests, http.StatusPaymentRequired, 582:
		k.todayAttempts += 1
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
		TodayAttempts int       `json:"today_attempts"`
		UsageDate     string    `json:"usage_date"`
		CooldownUntil time.Time `json:"cooldown_until"`
	}
	if err := json.Unmarshal(body, &keys); err != nil {
		return fmt.Errorf("금고 키 파싱 오류: %w", err)
	}

	km.mu.Lock()
	// preserve locally accumulated counters before replacing vault-sourced keys
	type localCounters struct{ usage, attempts int }
	localCtrs := make(map[string]localCounters)
	for _, svcKeys := range km.keys {
		for _, k := range svcKeys {
			if !strings.HasPrefix(k.id, "env-") {
				localCtrs[k.id] = localCounters{k.todayUsage, k.todayAttempts}
			}
		}
	}
	// detect midnight rollover: if the date changed since last sync, discard stale local counters
	// so that max(vault=0, local=yesterday) does not keep yesterday's values.
	// Uses UTC so proxies in different time zones (or with clock drift vs vault)
	// roll over at the same wall-clock moment and produce the same UsageDate.
	today := time.Now().UTC().Format("2006-01-02")
	if km.lastSyncDate != today {
		km.lastSyncDate = today
		localCtrs = make(map[string]localCounters) // all zeros — vault values win
		log.Printf("[sync] date changed to %s (UTC) — discarding stale local counters", today)
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
		// discard vault usage if it was written on a previous day (stale data protection)
		vaultUsage := k.TodayUsage
		vaultAttempts := k.TodayAttempts
		if k.UsageDate != "" && k.UsageDate != today {
			vaultUsage = 0
			vaultAttempts = 0
			log.Printf("[sync] discarding stale vault usage for key %s (vault date: %s, today: %s)", k.ID, k.UsageDate, today)
		}
		// use the higher of vault vs. locally accumulated for each counter
		usage := vaultUsage
		attempts := vaultAttempts
		if lc := localCtrs[k.ID]; lc.usage > usage {
			usage = lc.usage
		}
		if lc := localCtrs[k.ID]; lc.attempts > attempts {
			attempts = lc.attempts
		}
		lk := &localKey{
			id:            k.ID,
			service:       k.Service,
			plaintext:     k.PlainKey,
			dailyLimit:    k.DailyLimit,
			todayUsage:    usage,
			todayAttempts: attempts,
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
