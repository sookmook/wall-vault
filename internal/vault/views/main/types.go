package mainview

// ServiceVM is the view model for a service in the dashboard grids.
// It mirrors the fields of vault.ServiceConfig that the templates actually use,
// breaking the import cycle: views/* → vault → views/*.
type ServiceVM struct {
	ID            string
	Name          string
	DefaultModel  string
	LocalURL      string
	Enabled       bool
	ProxyEnabled  bool
	SortOrder     int
	AllowedModels []string

	KeyCount   int
	TodayUsage int
	DailyLimit int // 0 = unlimited
}

// ClientVM is the view model for a client/agent card in the dashboard.
// It mirrors vault.Client fields used by templates.
type ClientVM struct {
	ID               string
	Name             string
	AgentType        string
	PreferredService string
	ModelOverride    string
	Enabled          bool
	Avatar           string

	Online        bool
	RemoteModel   string
	LastHeartbeat string // human-readable "3분 전" / "just now"
	Uptime        string // human-readable uptime "2d 3h" / "15h 22m" / "4m"
	// Runtime: "daemon" (default) or "on_demand". Drives dot-color logic so
	// on_demand agents (cokacdir, lambda-style sessions) are not painted as
	// offline merely because they have no heartbeat in flight.
	Runtime string

	// ServiceDefaultModel is the PreferredService's configured default_model.
	// Populated when ModelOverride is empty so the card can render the model
	// the agent will actually receive (along with a "default" badge), instead
	// of leaving the model column blank and giving the impression of "no
	// model configured".
	ServiceDefaultModel string
}

// KeyVM is the view model for an API-key card in the dashboard.
// Sensitive material (encrypted_key) is never exposed to the template; the
// dashboard only sees the short ID prefix, label, service, today usage, and
// derived state ("active" | "cooldown" | "exhausted").
type KeyVM struct {
	ID         string
	IDShort    string
	Service    string
	Label      string
	TodayUsage int
	TodayAttempts int
	DailyLimit int    // 0 = unlimited
	UsagePct   int    // 0..100, 0 when DailyLimit==0
	Status     string // "active" | "cooldown" | "exhausted"
	Cooldown   string // remaining cooldown e.g. "12m left", empty when none
}
