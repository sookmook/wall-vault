package vault

import "net/http"

// RegisterHXRoutes wires /hx/* fragment endpoints. Real renderers land in
// Task 22 (Stage 6). This task installs the mux skeleton so later tasks can
// commit one handler at a time without fighting the router.
func (s *Server) RegisterHXRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hx/sidebar", s.hxNotImplemented)
	mux.HandleFunc("/hx/services/grid", s.hxNotImplemented)
	mux.HandleFunc("/hx/agents/grid", s.hxNotImplemented)
	mux.HandleFunc("/hx/keys/list", s.hxNotImplemented)
	// /hx/services/{id}/edit and /hx/clients/{id}/edit use sub-routing
	mux.HandleFunc("/hx/services/", s.hxServiceSubroute)
	mux.HandleFunc("/hx/clients/", s.hxClientSubroute)
}

func (s *Server) hxNotImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "hx endpoint not yet implemented", http.StatusNotImplemented)
}

// hxServiceSubroute handles /hx/services/{id}/edit — stub until Task 22.
func (s *Server) hxServiceSubroute(w http.ResponseWriter, r *http.Request) {
	s.hxNotImplemented(w, r)
}

// hxClientSubroute handles /hx/clients/{id}/edit — stub until Task 22.
func (s *Server) hxClientSubroute(w http.ResponseWriter, r *http.Request) {
	s.hxNotImplemented(w, r)
}
