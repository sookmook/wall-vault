package slideover

// ServiceVM is the view model for the service edit form.
type ServiceVM struct {
	ID            string
	Name          string
	DefaultModel  string
	LocalURL      string
	Enabled       bool
	ProxyEnabled  bool
	ReasoningMode bool     // local services: inject reasoning flag on forwarded requests
	SortOrder     int
	AllowedModels []string
	IsLocal       bool     // ollama / lmstudio / vllm / llamacpp — show LocalURL input + reasoning toggle
	Models        []string // live-queried model options for the default_model datalist
	// CatalogUnused lists provider-registry models that are NOT already in
	// DefaultModel or AllowedModels — rendered as a checkbox picker so admins
	// can add registry models into AllowedModels without typing.
	CatalogUnused []string
}

// ServiceModelGroup groups a service's candidate models into the default and
// admin-curated allowed list so the model_override <select> can render them
// as two separate <optgroup>s.
type ServiceModelGroup struct {
	Default string   `json:"default,omitempty"` // service default_model
	Allowed []string `json:"allowed,omitempty"` // admin-curated allowed_models minus the default
}

// ClientVM is the view model for the client edit form.
type ClientVM struct {
	ID               string
	Name             string
	AgentType        string
	PreferredService string
	ModelOverride    string
	Enabled          bool
	WorkDir          string
	Host             string // machine hostname for claude-code auto-match
	FallbackServices string // comma-joined ordered service ids; empty = strict primary-only
	IPWhitelist      string // comma-joined for single-line form input
	Avatar           string // current avatar data URI (preview only)
	// ServiceModelMap is service-ID → grouped candidate models. Serialized to
	// JSON and embedded in the page so the model_override <select> can
	// repopulate (as <optgroup>s) when the preferred_service changes.
	ServiceModelMap map[string]ServiceModelGroup
	// CurrentGroup is the pre-resolved group for PreferredService so the
	// initial <select> can be fully server-rendered with optgroups — the
	// JS hydrator only needs to kick in when preferred_service changes.
	CurrentGroup ServiceModelGroup
}

// OverrideInCurrentGroup reports whether ModelOverride already appears in
// either the default or the allowed list of CurrentGroup. Templates use
// this to decide whether to render a separate "(현재 값)" header option.
func (c *ClientVM) OverrideInCurrentGroup() bool {
	if c.ModelOverride == "" {
		return false
	}
	if c.ModelOverride == c.CurrentGroup.Default {
		return true
	}
	for _, m := range c.CurrentGroup.Allowed {
		if m == c.ModelOverride {
			return true
		}
	}
	return false
}
