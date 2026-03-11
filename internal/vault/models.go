package vault

import "time"

// ─── API 키 ──────────────────────────────────────────────────────────────────

type APIKey struct {
	ID            string    `json:"id"`
	Service       string    `json:"service"`        // google | openrouter | ollama
	EncryptedKey  string    `json:"encrypted_key"`
	Label         string    `json:"label"`
	TodayUsage    int       `json:"today_usage"`
	DailyLimit    int       `json:"daily_limit"`    // 0 = 무제한
	CooldownUntil time.Time `json:"cooldown_until"`
	LastError     int       `json:"last_error"`
	CreatedAt     time.Time `json:"created_at"`
}

func (k *APIKey) IsOnCooldown() bool {
	return time.Now().Before(k.CooldownUntil)
}

func (k *APIKey) IsExhausted() bool {
	return k.DailyLimit > 0 && k.TodayUsage >= k.DailyLimit
}

func (k *APIKey) UsagePct() int {
	if k.DailyLimit <= 0 {
		return 0
	}
	pct := k.TodayUsage * 100 / k.DailyLimit
	if pct > 100 {
		return 100
	}
	return pct
}

func (k *APIKey) IsAvailable() bool {
	return !k.IsOnCooldown() && !k.IsExhausted()
}

// ─── 클라이언트 ───────────────────────────────────────────────────────────────

type Client struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Token           string    `json:"token"`
	DefaultService  string    `json:"default_service"`
	DefaultModel    string    `json:"default_model"`
	AllowedServices []string  `json:"allowed_services"`
	CreatedAt       time.Time `json:"created_at"`
}

// ─── 프록시 상태 (Heartbeat) ─────────────────────────────────────────────────

type ProxyStatus struct {
	ClientID  string    `json:"client_id"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
	Model     string    `json:"model"`
	SSE       bool      `json:"sse_connected"`
	Host      string    `json:"host,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	Vault     VaultInfo `json:"vault,omitempty"`
}

type VaultInfo struct {
	TodayUsage    int    `json:"today_usage"`
	DailyLimit    int    `json:"daily_limit"`
	CooldownUntil string `json:"cooldown_until,omitempty"`
	KeyStatus     string `json:"key_status,omitempty"` // active | cooldown | exhausted
}

// ─── SSE 이벤트 ───────────────────────────────────────────────────────────────

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ConfigChangeEvent struct {
	ClientID string `json:"client_id"`
	Service  string `json:"service"`
	Model    string `json:"model"`
}

// ─── 저장소 스냅샷 ────────────────────────────────────────────────────────────

type storeData struct {
	Keys    []*APIKey      `json:"keys"`
	Clients []*Client      `json:"clients"`
	Proxies []*ProxyStatus `json:"proxies"`
}
