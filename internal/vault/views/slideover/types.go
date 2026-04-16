package slideover

// ServiceVM is the view model for the service edit form.
type ServiceVM struct {
	ID            string
	Name          string
	DefaultModel  string
	LocalURL      string
	Enabled       bool
	ProxyEnabled  bool
	SortOrder     int
	AllowedModels []string
	IsLocal       bool     // ollama / lmstudio / vllm — show LocalURL input
	Models        []string // live-queried model options for the default_model datalist
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
	IPWhitelist      string // comma-joined for single-line form input
	Avatar           string // current avatar data URI (preview only)
	// ServiceModelMap is service-ID → grouped candidate models. Serialized to
	// JSON and embedded in the page so the model_override <select> can
	// repopulate (as <optgroup>s) when the preferred_service changes.
	ServiceModelMap map[string]ServiceModelGroup
}
