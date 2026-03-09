package http

import "net/http"

// registerHealthRoutes groups all health/diagnostic routes.
func (s *Server) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.healthCheck)
}

