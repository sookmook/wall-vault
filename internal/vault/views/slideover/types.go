package slideover

// ServiceVM is the view model for the service edit form.
type ServiceVM struct {
	ID            string
	Name          string
	DefaultModel  string
	LocalURL      string
	ProxyEnabled  bool
	SortOrder     int
	AllowedModels []string
}

// ClientVM is the view model for the client edit form.
type ClientVM struct {
	ID               string
	Name             string
	AgentType        string
	PreferredService string
	ModelOverride    string
	Enabled          bool
}
