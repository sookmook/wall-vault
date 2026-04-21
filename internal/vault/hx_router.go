package vault

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sookmook/wall-vault/internal/models"
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
	// /hx/{services|clients|keys}/(new|{id}/edit) use sub-routing
	mux.HandleFunc("/hx/services/", s.hxServiceSubroute)
	mux.HandleFunc("/hx/clients/", s.hxClientSubroute)
	mux.HandleFunc("/hx/keys/", s.hxKeySubroute)
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
		out[i] = &sidebar.ClientVM{ID: c.ID, Name: c.Name, Avatar: c.Avatar}
	}
	return out
}

// toMainServices converts vault ServiceConfig slice to main-view service VMs,
// enriching each entry with live per-service key counts and aggregate usage.
func (s *Server) toMainServices(svcs []*ServiceConfig) []*mainview.ServiceVM {
	keys := s.store.ListKeys()
	counts := map[string]int{}
	usage := map[string]int{}
	limits := map[string]int{}
	for _, k := range keys {
		counts[k.Service]++
		usage[k.Service] += k.TodayUsage
		if k.DailyLimit > 0 {
			limits[k.Service] += k.DailyLimit
		}
	}
	out := make([]*mainview.ServiceVM, len(svcs))
	for i, sv := range svcs {
		out[i] = &mainview.ServiceVM{
			ID:            sv.ID,
			Name:          sv.Name,
			DefaultModel:  sv.DefaultModel,
			LocalURL:      sv.LocalURL,
			Enabled:       sv.Enabled,
			ProxyEnabled:  sv.ProxyEnabled,
			SortOrder:     sv.SortOrder,
			AllowedModels: sv.AllowedModels,
			KeyCount:      counts[sv.ID],
			TodayUsage:    usage[sv.ID],
			DailyLimit:    limits[sv.ID],
		}
	}
	return out
}

// toMainClients converts vault Client slice to main-view agent VMs,
// folding in live proxy heartbeat state (online flag, current model, age).
func (s *Server) toMainClients(clients []*Client) []*mainview.ClientVM {
	proxies := s.store.ListProxies()
	byID := make(map[string]*ProxyStatus, len(proxies))
	for _, p := range proxies {
		byID[p.ClientID] = p
	}
	out := make([]*mainview.ClientVM, len(clients))
	now := time.Now()
	for i, c := range clients {
		vm := &mainview.ClientVM{
			ID:               c.ID,
			Name:             c.Name,
			AgentType:        c.AgentType,
			PreferredService: c.PreferredService,
			ModelOverride:    c.ModelOverride,
			Enabled:          c.Enabled,
			Avatar:           c.Avatar,
		}
		if p := byID[c.ID]; p != nil {
			vm.Online = !p.UpdatedAt.IsZero() && now.Sub(p.UpdatedAt) < 90*time.Second
			vm.RemoteModel = p.Model
			if !p.UpdatedAt.IsZero() {
				vm.LastHeartbeat = relativeTime(now.Sub(p.UpdatedAt))
			}
			if !p.StartedAt.IsZero() && vm.Online {
				vm.Uptime = relativeUptime(now.Sub(p.StartedAt))
			}
		}
		out[i] = vm
	}
	return out
}

// toMainKeys converts vault APIKey slice to dashboard KeyVM slice. Sensitive
// material (encrypted_key) is dropped. Cooldown text is computed from
// CooldownUntil, status from IsAvailable / IsExhausted / IsOnCooldown.
func (s *Server) toMainKeys(keys []*APIKey) []*mainview.KeyVM {
	now := time.Now()
	out := make([]*mainview.KeyVM, len(keys))

	// First pass: find max usage among unlimited keys so we can show
	// relative bars ("how much compared to the busiest key").
	maxUsage := 0
	for _, k := range keys {
		if k.DailyLimit == 0 && k.TodayUsage > maxUsage {
			maxUsage = k.TodayUsage
		}
	}

	for i, k := range keys {
		short := k.ID
		if len(short) > 10 {
			short = short[:10]
		}
		pct := k.UsagePct() // uses DailyLimit when set
		if pct == 0 && k.DailyLimit == 0 && maxUsage > 0 {
			pct = k.TodayUsage * 100 / maxUsage
		}
		vm := &mainview.KeyVM{
			ID:            k.ID,
			IDShort:       short,
			Service:       k.Service,
			Label:         k.Label,
			TodayUsage:    k.TodayUsage,
			TodayAttempts: k.TodayAttempts,
			DailyLimit:    k.DailyLimit,
			UsagePct:      pct,
		}
		switch {
		case k.IsExhausted():
			vm.Status = "exhausted"
		case k.IsOnCooldown():
			vm.Status = "cooldown"
			vm.Cooldown = remainingLabel(k.CooldownUntil.Sub(now))
		default:
			vm.Status = "active"
		}
		out[i] = vm
	}
	return out
}

// remainingLabel formats a positive Duration as "Nd Nh", "Nh Nm", "Nm Ns", or "Ns left".
func remainingLabel(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	secs := int(d.Seconds())
	switch {
	case secs >= 86400:
		return fmt.Sprintf("%dd %dh left", secs/86400, (secs%86400)/3600)
	case secs >= 3600:
		return fmt.Sprintf("%dh %dm left", secs/3600, (secs%3600)/60)
	case secs >= 60:
		return fmt.Sprintf("%dm %ds left", secs/60, secs%60)
	default:
		return fmt.Sprintf("%ds left", secs)
	}
}

// relativeUptime formats a Duration as a compact uptime label without "ago".
func relativeUptime(d time.Duration) string {
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%dm", mins)
	}
	hrs := mins / 60
	if hrs < 24 {
		return fmt.Sprintf("%dh %dm", hrs, mins%60)
	}
	days := hrs / 24
	return fmt.Sprintf("%dd %dh", days, hrs%24)
}

// relativeTime formats a Duration as a short Korean/unit-agnostic label
// (e.g. "3s ago", "12m ago"). Kept ASCII so it renders consistently across locales.
func relativeTime(d time.Duration) string {
	secs := int(d.Seconds())
	if secs < 2 {
		return "just now"
	}
	if secs < 60 {
		return fmt.Sprintf("%ds ago", secs)
	}
	if mins := secs / 60; mins < 60 {
		return fmt.Sprintf("%dm ago", mins)
	}
	if hrs := secs / 3600; hrs < 24 {
		return fmt.Sprintf("%dh ago", hrs)
	}
	return fmt.Sprintf("%dd ago", secs/86400)
}

// toSlideoverService converts a vault ServiceConfig to a slideover ServiceVM,
// enriching the VM with live model options pulled from the shared registry so
// the edit form can render a default_model dropdown. Ensures the registry is
// populated first — the dashboard shouldn't rely on a separate /admin/models
// round-trip just to see model choices in the initial HTML.
func (s *Server) toSlideoverService(sv *ServiceConfig) *slideover.ServiceVM {
	s.ensureRegistry()
	var modelNames []string
	if s.registry != nil {
		for _, m := range s.registry.All(sv.ID) {
			if m.ID != "" {
				modelNames = append(modelNames, m.ID)
			}
		}
	}
	// ensureRegistry only refreshes ENABLED services, so a disabled service
	// (e.g. a newly-added llama.cpp the user hasn't enabled yet) would yield
	// an empty dropdown. Fall back to an on-demand single-service fetch so
	// the edit form can always show model choices.
	if len(modelNames) == 0 && s.registry != nil {
		// Decrypt the first available key for this service so providers
		// with a live model-list endpoint (openrouter, anthropic) can
		// populate the edit slideover dropdown instead of showing only
		// the static known-list fallback.
		keys := models.ServiceKeys{}
		for _, k := range s.store.ListKeys() {
			if k.Service == sv.ID && k.IsAvailable() {
				if plain, err := decryptKey(k.EncryptedKey, s.cfg.Vault.MasterPass); err == nil {
					keys[sv.ID] = plain
					break
				}
			}
		}
		s.registry.RefreshService(sv.ID, sv.LocalURL, keys)
		for _, m := range s.registry.All(sv.ID) {
			if m.ID != "" {
				modelNames = append(modelNames, m.ID)
			}
		}
	}
	inUse := map[string]bool{}
	if sv.DefaultModel != "" {
		inUse[sv.DefaultModel] = true
	}
	for _, m := range sv.AllowedModels {
		inUse[m] = true
	}
	var catalogUnused []string
	for _, m := range modelNames {
		if !inUse[m] {
			catalogUnused = append(catalogUnused, m)
		}
	}
	return &slideover.ServiceVM{
		ID:            sv.ID,
		Name:          sv.Name,
		DefaultModel:  sv.DefaultModel,
		LocalURL:      sv.LocalURL,
		Enabled:       sv.Enabled,
		ProxyEnabled:  sv.ProxyEnabled,
		ReasoningMode: sv.ReasoningMode,
		SortOrder:     sv.SortOrder,
		AllowedModels: sv.AllowedModels,
		IsLocal:       sv.IsLocal(),
		Models:        modelNames,
		CatalogUnused: catalogUnused,
	}
}

// toSlideoverClient converts a vault Client to a slideover ClientVM,
// folding in the service→candidate-models map so the edit form's
// model_override dropdown can repopulate when preferred_service changes.
// Models are split into default_model and the admin-curated allowed_models
// so the <select> can render them as separate <optgroup>s. Allowed entries
// that match the default are dropped from the allowed list to avoid dupes.
func (s *Server) toSlideoverClient(c *Client) *slideover.ClientVM {
	svcMap := make(map[string]slideover.ServiceModelGroup)
	for _, sv := range s.store.ListServices() {
		grp := slideover.ServiceModelGroup{Default: sv.DefaultModel}
		for _, m := range sv.AllowedModels {
			if m == "" || m == sv.DefaultModel {
				continue
			}
			grp.Allowed = append(grp.Allowed, m)
		}
		svcMap[sv.ID] = grp
	}
	return &slideover.ClientVM{
		ID:               c.ID,
		Name:             c.Name,
		AgentType:        c.AgentType,
		PreferredService: c.PreferredService,
		ModelOverride:    c.ModelOverride,
		Enabled:          c.Enabled,
		WorkDir:          c.WorkDir,
		IPWhitelist:      strings.Join(c.IPWhitelist, ", "),
		Avatar:           c.Avatar,
		ServiceModelMap:  svcMap,
		CurrentGroup:     svcMap[c.PreferredService],
	}
}

// emptyClientVM builds a ClientVM used by the create form — carries only the
// ServiceModelMap so the model_override dropdown is pre-populated.
func (s *Server) emptyClientVM() *slideover.ClientVM {
	return s.toSlideoverClient(&Client{})
}

func (s *Server) hxSidebar(w http.ResponseWriter, r *http.Request) {
	svcVMs := toSidebarServices(s.store.ListServices())
	clientVMs := toSidebarClients(s.store.ListClients())
	sidebar.Sidebar(svcVMs, clientVMs).Render(r.Context(), w) //nolint:errcheck
}

func (s *Server) hxServicesGrid(w http.ResponseWriter, r *http.Request) {
	mainview.ServicesGrid(s.toMainServices(s.store.ListServices())).Render(r.Context(), w) //nolint:errcheck
}

func (s *Server) hxAgentsGrid(w http.ResponseWriter, r *http.Request) {
	mainview.AgentsGrid(s.toMainClients(s.store.ListClients())).Render(r.Context(), w) //nolint:errcheck
}

func (s *Server) hxKeysList(w http.ResponseWriter, r *http.Request) {
	mainview.KeysGrid(s.toMainKeys(s.store.ListKeys())).Render(r.Context(), w) //nolint:errcheck
}

// hxServiceSubroute handles two routes:
//   /hx/services/new        → blank create form
//   /hx/services/{id}/edit  → edit form for a specific service
func (s *Server) hxServiceSubroute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/hx/services/")
	if path == "new" {
		slideover.Frame("새 서비스", slideover.ServiceCreate()).Render(r.Context(), w) //nolint:errcheck
		return
	}
	id := strings.TrimSuffix(path, "/edit")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	svc := s.store.GetService(id)
	if svc == nil {
		http.NotFound(w, r)
		return
	}
	slideover.Frame(svc.Name, slideover.ServiceEdit(s.toSlideoverService(svc))).Render(r.Context(), w) //nolint:errcheck
}

// hxKeySubroute handles /hx/keys/new → blank key create form.
func (s *Server) hxKeySubroute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/hx/keys/")
	if path != "new" {
		http.NotFound(w, r)
		return
	}
	svcVMs := make([]*slideover.ServiceVM, 0)
	for _, svc := range s.store.ListServices() {
		if svc.Enabled {
			svcVMs = append(svcVMs, s.toSlideoverService(svc))
		}
	}
	slideover.Frame("새 API 키", slideover.KeyCreate(&slideover.KeyCreateVM{Services: svcVMs})).Render(r.Context(), w) //nolint:errcheck
}

// hxClientSubroute matches both:
//   /hx/clients/new        → blank create form
//   /hx/clients/{id}/edit  → edit form for a specific client
// Both render inside the slideover frame so the layout stays consistent.
func (s *Server) hxClientSubroute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/hx/clients/")

	// build service VMs reused by both create and edit
	svcVMs := make([]*slideover.ServiceVM, 0)
	for _, svc := range s.store.ListServices() {
		svcVMs = append(svcVMs, s.toSlideoverService(svc))
	}

	if path == "new" {
		slideover.Frame("새 에이전트", slideover.ClientCreate(svcVMs, s.emptyClientVM().ServiceModelMap)).Render(r.Context(), w) //nolint:errcheck
		return
	}
	if path == "" || path == "/" {
		http.NotFound(w, r)
		return
	}

	id := strings.TrimSuffix(path, "/edit")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	c := s.store.GetClient(id)
	if c == nil {
		http.NotFound(w, r)
		return
	}
	slideover.Frame(c.Name, slideover.ClientEdit(s.toSlideoverClient(c), svcVMs)).Render(r.Context(), w) //nolint:errcheck
}
