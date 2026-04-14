package vault

import (
	"net/http"
	"strings"

	mainview "github.com/sookmook/wall-vault/internal/vault/views/main"
	sidebar "github.com/sookmook/wall-vault/internal/vault/views/sidebar"
	slideover "github.com/sookmook/wall-vault/internal/vault/views/slideover"
)

// RegisterHXRoutes wires /hx/* fragment endpoints backed by templ components.
func (s *Server) RegisterHXRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hx/sidebar", s.hxSidebar)
	mux.HandleFunc("/hx/services/grid", s.hxServicesGrid)
	mux.HandleFunc("/hx/agents/grid", s.hxAgentsGrid)
	mux.HandleFunc("/hx/keys/list", s.hxKeysList)
	// /hx/services/{id}/edit and /hx/clients/{id}/edit use sub-routing
	mux.HandleFunc("/hx/services/", s.hxServiceSubroute)
	mux.HandleFunc("/hx/clients/", s.hxClientSubroute)
}

// toSidebarServices converts vault ServiceConfig slice to sidebar view models.
func toSidebarServices(svcs []*ServiceConfig) []*sidebar.ServiceVM {
	out := make([]*sidebar.ServiceVM, len(svcs))
	for i, s := range svcs {
		out[i] = &sidebar.ServiceVM{ID: s.ID, Name: s.Name}
	}
	return out
}

// toSidebarClients converts vault Client slice to sidebar view models.
func toSidebarClients(clients []*Client) []*sidebar.ClientVM {
	out := make([]*sidebar.ClientVM, len(clients))
	for i, c := range clients {
		out[i] = &sidebar.ClientVM{ID: c.ID, Name: c.Name}
	}
	return out
}

// toMainServices converts vault ServiceConfig slice to main-view service VMs.
func toMainServices(svcs []*ServiceConfig) []*mainview.ServiceVM {
	out := make([]*mainview.ServiceVM, len(svcs))
	for i, s := range svcs {
		out[i] = &mainview.ServiceVM{
			ID:            s.ID,
			Name:          s.Name,
			DefaultModel:  s.DefaultModel,
			LocalURL:      s.LocalURL,
			Enabled:       s.Enabled,
			ProxyEnabled:  s.ProxyEnabled,
			SortOrder:     s.SortOrder,
			AllowedModels: s.AllowedModels,
		}
	}
	return out
}

// toMainClients converts vault Client slice to main-view agent VMs.
func toMainClients(clients []*Client) []*mainview.ClientVM {
	out := make([]*mainview.ClientVM, len(clients))
	for i, c := range clients {
		out[i] = &mainview.ClientVM{
			ID:               c.ID,
			Name:             c.Name,
			AgentType:        c.AgentType,
			PreferredService: c.PreferredService,
			ModelOverride:    c.ModelOverride,
			Enabled:          c.Enabled,
		}
	}
	return out
}

// toSlideoverService converts a vault ServiceConfig to a slideover ServiceVM.
func toSlideoverService(s *ServiceConfig) *slideover.ServiceVM {
	return &slideover.ServiceVM{
		ID:            s.ID,
		Name:          s.Name,
		DefaultModel:  s.DefaultModel,
		LocalURL:      s.LocalURL,
		ProxyEnabled:  s.ProxyEnabled,
		SortOrder:     s.SortOrder,
		AllowedModels: s.AllowedModels,
	}
}

// toSlideoverClient converts a vault Client to a slideover ClientVM.
func toSlideoverClient(c *Client) *slideover.ClientVM {
	return &slideover.ClientVM{
		ID:               c.ID,
		Name:             c.Name,
		AgentType:        c.AgentType,
		PreferredService: c.PreferredService,
		ModelOverride:    c.ModelOverride,
		Enabled:          c.Enabled,
	}
}

func (s *Server) hxSidebar(w http.ResponseWriter, r *http.Request) {
	svcVMs := toSidebarServices(s.store.ListServices())
	clientVMs := toSidebarClients(s.store.ListClients())
	sidebar.Sidebar(svcVMs, clientVMs).Render(r.Context(), w) //nolint:errcheck
}

func (s *Server) hxServicesGrid(w http.ResponseWriter, r *http.Request) {
	mainview.ServicesGrid(toMainServices(s.store.ListServices())).Render(r.Context(), w) //nolint:errcheck
}

func (s *Server) hxAgentsGrid(w http.ResponseWriter, r *http.Request) {
	mainview.AgentsGrid(toMainClients(s.store.ListClients())).Render(r.Context(), w) //nolint:errcheck
}

func (s *Server) hxKeysList(w http.ResponseWriter, r *http.Request) {
	// Placeholder: keys list UI lands in a later round; for now show a note.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<p>Keys list UI coming in a later round.</p>`)) //nolint:errcheck
}

// hxServiceSubroute matches /hx/services/{id}/edit — renders slideover frame around ServiceEdit.
func (s *Server) hxServiceSubroute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/hx/services/")
	id = strings.TrimSuffix(id, "/edit")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	svc := s.store.GetService(id)
	if svc == nil {
		http.NotFound(w, r)
		return
	}
	slideover.Frame(svc.Name, slideover.ServiceEdit(toSlideoverService(svc))).Render(r.Context(), w) //nolint:errcheck
}

// hxClientSubroute matches /hx/clients/{id}/edit — renders slideover frame around ClientEdit.
func (s *Server) hxClientSubroute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/hx/clients/")
	id = strings.TrimSuffix(id, "/edit")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	c := s.store.GetClient(id)
	if c == nil {
		http.NotFound(w, r)
		return
	}
	// build service VMs for the select dropdown
	svcVMs := make([]*slideover.ServiceVM, 0)
	for _, svc := range s.store.ListServices() {
		svcVMs = append(svcVMs, toSlideoverService(svc))
	}
	slideover.Frame(c.Name, slideover.ClientEdit(toSlideoverClient(c), svcVMs)).Render(r.Context(), w) //nolint:errcheck
}
