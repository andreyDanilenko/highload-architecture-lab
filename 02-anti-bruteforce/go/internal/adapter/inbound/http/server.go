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
	srv.registerRoutes(mux)
	return srv
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.httpServer.Close()
}
