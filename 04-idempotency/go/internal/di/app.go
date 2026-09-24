package di

import (
	"flag"
	"log"
	"time"

	httpadapter "idempotency/internal/adapter/inbound/http"
	"idempotency/internal/adapter/inbound/http/handlers"
	"idempotency/internal/usecase"
	"idempotency/pkg/idempotency"
)

const defaultLockTTL = 60 * time.Second

// App — composition root: здесь собираются все зависимости (как в 03-worker-pool).
type App struct {
	Server *httpadapter.Server
}

// NewApp читает флаги, выбирает провайдер идемпотентности, поднимает HTTP.
func NewApp() *App {
	providerType := flag.String("provider", "noop", "idempotency provider: noop | memory | redis")
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	provider := selectProvider(*providerType, defaultLockTTL)
	mw := idempotency.NewMiddleware(provider, defaultLockTTL)

	payments := usecase.NewPaymentService()
	paymentHandler := handlers.NewPaymentHandler(payments, mw)
	srv := httpadapter.New(*addr, paymentHandler)

	log.Printf("idempotency provider=%s addr=%s", provider.Name(), *addr)
	return &App{Server: srv}
}

func selectProvider(name string, ttl time.Duration) idempotency.Provider {
	switch name {
	case "memory":
		log.Println("provider: in-memory (single instance only)")
		return idempotency.NewMemoryProvider(ttl)
	case "redis":
		log.Fatal("RedisProvider: implement in docs/subtask-3-redis-provider.md")
		return nil
	default:
		log.Println("provider: noop — каждый retry создаёт новый side-effect (демо бага)")
		return idempotency.NewNoopProvider()
	}
}
