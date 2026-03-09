package http

import "net/http"

// registerRoutes is the top-level router composition entrypoint.
// It delegates route registration to semantic groups, similar to modules
// in a larger project (health, auth, resources, admin, etc.).
func (s *Server) registerRoutes(mux *http.ServeMux) {
	s.registerHealthRoutes(mux)
	s.registerAuthRoutes(mux)
	s.registerResourceRoutes(mux)
	s.registerFallbackRoutes(mux)
}
