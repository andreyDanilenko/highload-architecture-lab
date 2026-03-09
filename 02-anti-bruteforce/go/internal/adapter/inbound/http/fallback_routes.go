package http

import "net/http"

// registerFallbackRoutes wires catch-all / not-found style routes.
func (s *Server) registerFallbackRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.notFound)
}

