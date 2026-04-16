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
	// ServiceModelMap is service-ID → candidate models (default_model first,
	// then allowed_models). Serialized to JSON and embedded in the page so the
	// model_override <select> can repopulate when the preferred_service changes.
	ServiceModelMap map[string][]string
}
