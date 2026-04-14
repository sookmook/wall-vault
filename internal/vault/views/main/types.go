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
}
