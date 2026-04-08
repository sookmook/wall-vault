package vault

import "time"

// ─── API Key ──────────────────────────────────────────────────────────────────

type APIKey struct {
	ID             string    `json:"id"`
	Service        string    `json:"service"`         // google | openrouter | ollama
	EncryptedKey   string    `json:"encrypted_key"`
	Label          string    `json:"label"`
	TodayUsage     int       `json:"today_usage"`     // successful tokens (or 1/request when unavailable)
	TodayAttempts  int       `json:"today_attempts"`  // total requests including rate-limited
	UsageDate      string    `json:"usage_date"`      // "YYYY-MM-DD" of when today_usage was last written
	DailyLimit     int       `json:"daily_limit"`     // 0 = unlimited
	CooldownUntil  time.Time `json:"cooldown_until"`
	LastError      int       `json:"last_error"`
	CreatedAt      time.Time `json:"created_at"`
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

// ─── Client ───────────────────────────────────────────────────────────────────

type Client struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Token           string    `json:"token"`
	DefaultService  string    `json:"default_service"`
	DefaultModel    string    `json:"default_model"`
	AllowedServices []string  `json:"allowed_services"`
	// extended fields
	AgentType   string   `json:"agent_type,omitempty"`   // openclaw | claude-code | cursor | custom
	WorkDir     string   `json:"work_dir,omitempty"`     // working directory
	Description string   `json:"description,omitempty"`  // description
	IPWhitelist []string `json:"ip_whitelist,omitempty"` // allowed IP list (empty array = allow all)
	Avatar      string   `json:"avatar,omitempty"`       // data URI (data:image/...) OR relative path under ~/.openclaw/ (e.g. "workspace/avatar.png")
	Enabled     bool     `json:"enabled"`                // enabled status
	SortOrder   int      `json:"sort_order"`             // dashboard card order (lower = first)
	CreatedAt   time.Time `json:"created_at"`
}

// ClientInput: client add DTO
type ClientInput struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Token           string   `json:"token"`
	DefaultService  string   `json:"default_service"`
	DefaultModel    string   `json:"default_model"`
	AllowedServices []string `json:"allowed_services"`
	AgentType       string   `json:"agent_type"`
	WorkDir         string   `json:"work_dir"`
	Description     string   `json:"description"`
	IPWhitelist     []string `json:"ip_whitelist"`
	Avatar          string   `json:"avatar,omitempty"`
	Enabled         *bool    `json:"enabled"` // nil = default true
}

// ClientUpdateInput: client update DTO
// All fields are pointers/slices — nil/omitted = no change, value present = update
type ClientUpdateInput struct {
	NewID           *string  `json:"new_id"`
	Name            *string  `json:"name"`
	Token           *string  `json:"token"`
	DefaultService  *string  `json:"default_service"`
	DefaultModel    *string  `json:"default_model"`
	AllowedServices []string `json:"allowed_services"`
	AgentType       *string  `json:"agent_type"`
	WorkDir         *string  `json:"work_dir"`
	Description     *string  `json:"description"`
	IPWhitelist     []string `json:"ip_whitelist"`
	Avatar          *string  `json:"avatar"`
	Enabled         *bool    `json:"enabled"`
}

// ─── Service Config ───────────────────────────────────────────────────────────

// ServiceConfig: per-service runtime settings (local URL, enabled status)
type ServiceConfig struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	LocalURL     string `json:"local_url,omitempty"` // Ollama/LMStudio/vLLM only
	Enabled      bool   `json:"enabled"`
	Custom       bool   `json:"custom,omitempty"`       // user-added service
	ProxyEnabled bool   `json:"proxy_enabled,omitempty"` // 오픈클로 프록시에서 이 서비스 사용
}

// IsLocal: whether this is a local server service
func (s *ServiceConfig) IsLocal() bool {
	switch s.ID {
	case "ollama", "lmstudio", "vllm":
		return true
	}
	return s.Custom && s.LocalURL != ""
}

// ─── Proxy Status (Heartbeat) ────────────────────────────────────────────────

type ProxyStatus struct {
	ClientID   string            `json:"client_id"`
	Version    string            `json:"version"`
	Service    string            `json:"service"`
	Model      string            `json:"model"`
	SSE        bool              `json:"sse_connected"`
	Host       string            `json:"host,omitempty"`
	Avatar     string            `json:"avatar,omitempty"`        // base64 data URI sent by proxy
	StartedAt  time.Time         `json:"started_at,omitempty"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Vault      VaultInfo         `json:"vault,omitempty"`
	ActiveKeys    map[string]string  `json:"active_keys,omitempty"`    // service → key ID
	KeyUsage      map[string]int     `json:"key_usage,omitempty"`      // key ID → successful tokens today
	KeyAttempts   map[string]int     `json:"key_attempts,omitempty"`   // key ID → total requests today
	KeyCooldowns  map[string]string  `json:"key_cooldowns,omitempty"`  // key ID → cooldown RFC3339
	ActiveClients []ActiveClientItem `json:"active_clients,omitempty"` // recently-served non-proxy clients
	AgentAlive    *bool              `json:"agent_alive,omitempty"`    // local agent process alive (nanoclaw/openclaw)
}

// ActiveClientItem: a non-proxy client recently served through this proxy
type ActiveClientItem struct {
	ClientID string `json:"client_id"`
	Service  string `json:"service"`
	Model    string `json:"model"`
}

type VaultInfo struct {
	TodayUsage    int    `json:"today_usage"`
	DailyLimit    int    `json:"daily_limit"`
	CooldownUntil string `json:"cooldown_until,omitempty"`
	KeyStatus     string `json:"key_status,omitempty"` // active | cooldown | exhausted
}

// ─── SSE Event ────────────────────────────────────────────────────────────────

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ConfigChangeEvent struct {
	ClientID  string `json:"client_id"`
	Service   string `json:"service"`
	Model     string `json:"model"`
	AgentType string `json:"agent_type,omitempty"`
}

// ─── Store Snapshot ───────────────────────────────────────────────────────────

type storeData struct {
	Keys     []*APIKey        `json:"keys"`
	Clients  []*Client        `json:"clients"`
	Proxies  []*ProxyStatus   `json:"proxies"`
	Services []*ServiceConfig `json:"services,omitempty"`
	Settings *StoreSettings   `json:"settings,omitempty"`
}

// StoreSettings: UI settings persisted in vault.json (theme, language)
type StoreSettings struct {
	Theme string `json:"theme,omitempty"`
	Lang  string `json:"lang,omitempty"`
}
