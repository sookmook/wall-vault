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
	// 확장 필드
	AgentType   string   `json:"agent_type,omitempty"`   // openclaw | claude-code | cursor | custom
	WorkDir     string   `json:"work_dir,omitempty"`     // 작업 디렉토리
	Description string   `json:"description,omitempty"`  // 설명
	IPWhitelist []string `json:"ip_whitelist,omitempty"` // 허용 IP 목록 (빈 배열 = 모두 허용)
	Enabled     bool     `json:"enabled"`                // 활성화 여부
	CreatedAt   time.Time `json:"created_at"`
}

// ClientInput: 클라이언트 추가 DTO
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
	Enabled         *bool    `json:"enabled"` // nil = 기본값 true
}

// ClientUpdateInput: 클라이언트 수정 DTO
type ClientUpdateInput struct {
	Name            *string  `json:"name"`
	Token           *string  `json:"token"`
	DefaultService  string   `json:"default_service"`
	DefaultModel    string   `json:"default_model"`
	AllowedServices []string `json:"allowed_services"`
	AgentType       *string  `json:"agent_type"`
	WorkDir         *string  `json:"work_dir"`
	Description     *string  `json:"description"`
	IPWhitelist     []string `json:"ip_whitelist"`
	Enabled         *bool    `json:"enabled"`
}

// ─── 서비스 설정 ──────────────────────────────────────────────────────────────

// ServiceConfig: 서비스별 런타임 설정 (로컬 URL, 활성화 여부)
type ServiceConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LocalURL string `json:"local_url,omitempty"` // Ollama/LMStudio/vLLM 전용
	Enabled  bool   `json:"enabled"`
	Custom   bool   `json:"custom,omitempty"` // 사용자가 직접 추가한 서비스
}

// IsLocal: 로컬 서버 서비스 여부
func (s *ServiceConfig) IsLocal() bool {
	switch s.ID {
	case "ollama", "lmstudio", "vllm":
		return true
	}
	return s.Custom && s.LocalURL != ""
}

// ─── 프록시 상태 (Heartbeat) ─────────────────────────────────────────────────

type ProxyStatus struct {
	ClientID  string    `json:"client_id"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
	Model     string    `json:"model"`
	SSE       bool      `json:"sse_connected"`
	Host      string    `json:"host,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
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
	Keys     []*APIKey        `json:"keys"`
	Clients  []*Client        `json:"clients"`
	Proxies  []*ProxyStatus   `json:"proxies"`
	Services []*ServiceConfig `json:"services,omitempty"`
	Settings *StoreSettings   `json:"settings,omitempty"`
}

// StoreSettings: vault.json에 영속화되는 UI 설정 (테마, 언어)
type StoreSettings struct {
	Theme string `json:"theme,omitempty"`
	Lang  string `json:"lang,omitempty"`
}
