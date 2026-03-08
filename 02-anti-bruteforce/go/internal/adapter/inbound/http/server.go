package http

import (
	"fmt"
	"net/http"
	"time"
)

// Server is the inbound HTTP API adapter.
type Server struct {
	httpServer *http.Server
	deps       *Deps
}

// New creates the HTTP server and registers routes.
func New(deps *Deps) *Server {
	mux := http.NewServeMux()
	srv := &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", deps.Config.Host, deps.Config.Port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		deps: deps,
	}
	srv.routes(mux)
	return srv
}

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.healthCheck)
	mux.HandleFunc("/login", s.handleLoginRequired(s.deps.NaiveChecker))
	mux.HandleFunc("/resource/naive", s.handleLoginRequired(s.deps.NaiveChecker))
	mux.HandleFunc("/resource/pessimistic", s.handleLoginRequired(s.deps.PessimisticChecker))
	mux.HandleFunc("/resource/optimistic", s.handleLoginRequired(s.deps.OptimisticChecker))
	mux.HandleFunc("/vault/login", s.handleLoginRequired(s.deps.VaultChecker))
	mux.HandleFunc("/", s.notFound)
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.httpServer.Close()
}
