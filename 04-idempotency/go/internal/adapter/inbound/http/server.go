package httpadapter

import (
	"context"
	"net/http"
	"time"

	"idempotency/internal/adapter/inbound/http/handlers"
	"idempotency/internal/adapter/inbound/http/routes"
	"idempotency/pkg/idempotency"
)

type Server struct{ httpServer *http.Server }

// New ограничивает время чтения и размер заголовков; тело ограничено middleware.
func New(addr string, payment *handlers.PaymentHandler, middleware *idempotency.HTTPMiddleware, ready func(context.Context) error) *Server {
	mux := http.NewServeMux()
	routes.Register(mux, payment, middleware, ready)
	return &Server{httpServer: &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}}
}
func (s *Server) Start() error                       { return s.httpServer.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }
func (s *Server) Close() error                       { return s.httpServer.Close() }
func (s *Server) Handler() http.Handler              { return s.httpServer.Handler }
