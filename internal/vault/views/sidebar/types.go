package sidebar

// ServiceVM is the sidebar view model for a service entry.
type ServiceVM struct {
	ID   string
	Name string
}

// ClientVM is the sidebar view model for a client entry.
type ClientVM struct {
	ID     string
	Name   string
	Avatar string
}
