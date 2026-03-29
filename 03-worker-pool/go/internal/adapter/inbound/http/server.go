package http

import (
	"fmt"
	"net/http"
	"time"

	"worker-pool/internal/adapter/inbound/http/routes"
)

// Server is the inbound HTTP API adapter.
type Server struct {
	httpServer *http.Server
	deps       *Deps
}

// New creates the HTTP server and registers routes.
func New(deps *Deps) *Server {
	mux := http.NewServeMux()
	readTO := time.Duration(deps.Config.HTTPReadTimeoutSec) * time.Second
	writeTO := time.Duration(deps.Config.HTTPWriteTimeoutSec) * time.Second
	if readTO <= 0 {
		readTO = 600 * time.Second
	}
	if writeTO <= 0 {
		writeTO = 600 * time.Second
	}
	srv := &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", deps.Config.Host, deps.Config.Port),
			Handler:      mux,
			ReadTimeout:  readTO,
			WriteTimeout: writeTO,
			IdleTimeout:  120 * time.Second,
		},
		deps: deps,
	}
	routes.Register(
		mux,
		deps.Config,
		deps.Naive,
		deps.Bounded,
		deps.Reliable,
		deps.Advanced,
	)
	return srv
}

// Start blocks until the server stops.
func (s *Server) Start() error {
	fmt.Printf("Server starting on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop closes the HTTP server (stops accepting new connections).
func (s *Server) Stop() error {
	return s.httpServer.Close()
}
