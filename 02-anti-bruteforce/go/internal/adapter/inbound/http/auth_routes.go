package http

import "net/http"

// registerAuthRoutes groups all login/authentication-style routes.
// In a larger project сюда бы попали /login, /register, /refresh, /logout и т.п.
func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	// Naive login (basic anti-bruteforce strategy)
	mux.HandleFunc("/login", s.loginHandler())

	// Vault login (separate strategy, e.g. more strict checks)
	mux.HandleFunc("/vault/login", s.vaultLoginHandler())
}

