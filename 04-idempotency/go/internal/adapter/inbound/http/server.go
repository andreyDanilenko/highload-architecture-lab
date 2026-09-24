package httpadapter

import (
	"fmt"
	"net/http"

	"idempotency/internal/adapter/inbound/http/handlers"
	"idempotency/internal/adapter/inbound/http/routes"
)

// Server — inbound adapter: принимает HTTP, отдаёт handlers.
type Server struct {
	httpServer *http.Server
}

// New собирает mux и http.Server.
func New(addr string, payment *handlers.PaymentHandler) *Server {
	mux := http.NewServeMux()
	routes.Register(mux, payment)
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.httpServer.Close()
}
