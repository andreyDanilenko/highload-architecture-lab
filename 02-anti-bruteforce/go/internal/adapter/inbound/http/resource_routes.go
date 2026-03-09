package http

import "net/http"

// registerResourceRoutes groups routes that represent protected "resources"
// with different anti-bruteforce strategies.
func (s *Server) registerResourceRoutes(mux *http.ServeMux) {
	// Naive in-memory strategy
	mux.HandleFunc("/resource/naive", s.naiveResourceHandler())

	// Future Redis-based strategies (pessimistic/optimistic)
	mux.HandleFunc("/resource/pessimistic", s.pessimisticResourceHandler())
	mux.HandleFunc("/resource/optimistic", s.optimisticResourceHandler())
}

