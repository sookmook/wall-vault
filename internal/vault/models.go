package vault

import (
	"encoding/json"
	"strings"
	"time"
)

// StringOrList accepts either a JSON array of strings or a single
// comma-delimited string and normalises it to []string. Used for form fields
// where the dashboard submits one text input but CLI/API clients submit arrays.
type StringOrList []string

func (s *StringOrList) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*s = nil
		return nil
	}
	if trimmed[0] == '[' {
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		out := make([]string, 0, len(arr))
		for _, v := range arr {
			if v = strings.TrimSpace(v); v != "" {
				out = append(out, v)
			}
		}
		*s = out
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parts := strings.Split(str, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	*s = out
	return nil
}

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

// Client: proxy client record stored in vault.
// NOTE (v0.2): external references in store.go / server.go / proxy/*.go migrate in subsequent tasks. Keeping models self-consistent here.
// PreferredService / ModelOverride are the v0.2 canonical field names.
// DefaultService / DefaultModel are retained for backward compatibility until Stage 2 migration removes them.
// TODO (v0.2 Stage 2): remove DefaultService, DefaultModel, Description after migrating all callers.
type Client struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Token            string    `json:"token"`
	PreferredService string    `json:"preferred_service"`           // v0.2 canonical
	ModelOverride    string    `json:"model_override,omitempty"`    // v0.2 canonical
	AllowedServices  []string  `json:"allowed_services,omitempty"`
	// v0.2 extended fields
	AgentType   string   `json:"agent_type,omitempty"`
	WorkDir     string   `json:"work_dir,omitempty"`
	IPWhitelist []string `json:"ip_whitelist,omitempty"`
	Avatar      string   `json:"avatar,omitempty"`
	Enabled     bool     `json:"enabled"`
	SortOrder   int      `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	// TODO (v0.2 Stage 2): remove legacy fields below after migrating store.go / server.go / ui.go / proxy/server.go
	DefaultService  string `json:"default_service,omitempty"`  // legacy — use PreferredService
	DefaultModel    string `json:"default_model,omitempty"`    // legacy — use ModelOverride
	Description     string `json:"description,omitempty"`      // legacy — drop in Stage 2
}

// ClientInput: client add DTO
type ClientInput struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Token           string       `json:"token"`
	DefaultService  string       `json:"default_service"`
	DefaultModel    string       `json:"default_model"`
	AllowedServices StringOrList `json:"allowed_services"`
	AgentType       string       `json:"agent_type"`
	WorkDir         string       `json:"work_dir"`
	Description     string       `json:"description"`
	IPWhitelist     StringOrList `json:"ip_whitelist"`
	Avatar          string       `json:"avatar,omitempty"`
	Enabled         *bool        `json:"enabled"` // nil = default true
}

// ClientUpdateInput: client update DTO
// All fields are pointers/slices — nil/omitted = no change, value present = update.
// NOTE (v0.2): PreferredService / ModelOverride are the canonical v0.2 field names.
// DefaultService / DefaultModel are retained for backward compat until Stage 2 migration.
// TODO (v0.2 Stage 2): remove DefaultService, DefaultModel, Description after migrating all callers.
type ClientUpdateInput struct {
	NewID            *string  `json:"new_id"`
	Name             *string  `json:"name"`
	Token            *string  `json:"token"`
	PreferredService *string  `json:"preferred_service"`        // v0.2 canonical
	ModelOverride    *string  `json:"model_override"`           // v0.2 canonical
	AllowedServices  StringOrList `json:"allowed_services"`
	AgentType        *string      `json:"agent_type"`
	WorkDir          *string      `json:"work_dir"`
	IPWhitelist      StringOrList `json:"ip_whitelist"`
	Avatar           *string  `json:"avatar"`
	Enabled          *bool    `json:"enabled"`
	// TODO (v0.2 Stage 2): remove legacy fields below after migrating store.go / server.go
	DefaultService  *string  `json:"default_service"`  // legacy — use PreferredService
	DefaultModel    *string  `json:"default_model"`    // legacy — use ModelOverride
	Description     *string  `json:"description"`      // legacy — drop in Stage 2
}

// ─── Service (v0.2) ───────────────────────────────────────────────────────────

// Service: per-service config including model defaults and allowed model list.
// NOTE (v0.2): external references in store.go / server.go / proxy/*.go migrate in subsequent tasks. Keeping models self-consistent here.
type Service struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	DefaultModel  string   `json:"default_model"`
	LocalURL      string   `json:"local_url,omitempty"`
	Enabled       bool     `json:"enabled"`
	ProxyEnabled  bool     `json:"proxy_enabled"`
	SortOrder     int      `json:"sort_order"`
	AllowedModels []string `json:"allowed_models,omitempty"`
}

// IsLocal: whether this is a local server service
func (s *Service) IsLocal() bool {
	switch s.ID {
	case "ollama", "lmstudio", "vllm":
		return true
	}
	return s.LocalURL != ""
}

// ─── Service Config (legacy — migrate to Service in subsequent tasks) ─────────

// ServiceConfig: legacy alias retained for store.go / server.go / ui.go until Stage 2 migration.
// TODO (v0.2 Stage 2): replace all ServiceConfig references with Service and remove this struct.
type ServiceConfig struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	LocalURL      string   `json:"local_url,omitempty"` // Ollama/LMStudio/vLLM only
	Enabled       bool     `json:"enabled"`
	Custom        bool     `json:"custom,omitempty"`          // user-added service
	ProxyEnabled  bool     `json:"proxy_enabled,omitempty"`   // enabled for proxy dispatch
	DefaultModel  string   `json:"default_model,omitempty"`   // v0.2: most-common client model for this service
	AllowedModels []string `json:"allowed_models,omitempty"`  // v0.2: whitelist of models for this service
	SortOrder     int      `json:"sort_order,omitempty"`      // display order preserved from v1
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

// CurrentSchemaVersion is the version stamped onto every persisted vault.json.
// Task 6 migration code references this same constant when deciding whether
// an existing file needs upgrading.
const CurrentSchemaVersion = 2

type storeData struct {
	SchemaVersion int              `json:"schema_version"`
	Keys          []*APIKey        `json:"keys"`
	Clients       []*Client        `json:"clients"`
	Proxies       []*ProxyStatus   `json:"proxies"`
	Services      []*ServiceConfig `json:"services,omitempty"`
	Settings      *StoreSettings   `json:"settings,omitempty"`
}

// StoreSettings: UI settings persisted in vault.json (theme, language)
type StoreSettings struct {
	Theme string `json:"theme,omitempty"`
	Lang  string `json:"lang,omitempty"`
}
