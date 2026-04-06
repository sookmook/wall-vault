package rtk

// Registry maps command names to their filters.
type Registry struct {
	filters map[string]CommandFilter
}

// NewRegistry creates a registry with all built-in command filters.
func NewRegistry() *Registry {
	r := &Registry{filters: make(map[string]CommandFilter)}
	r.Register(&GitFilter{})
	r.Register(&GoFilter{})
	r.Register(&GeneralFilter{})
	return r
}

// Register adds a command filter.
func (r *Registry) Register(f CommandFilter) {
	r.filters[f.Name()] = f
}

// Lookup returns the filter for a command, or nil for passthrough.
func (r *Registry) Lookup(cmd string) CommandFilter {
	return r.filters[cmd]
}
