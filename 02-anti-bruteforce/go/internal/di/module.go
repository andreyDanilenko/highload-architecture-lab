package di

import (
	"context"
	"log"

	"go.uber.org/fx"

	httpadapter "anti-bruteforce/internal/adapter/inbound/http"
	"anti-bruteforce/internal/adapter/outbound/memory"
	redisadapter "anti-bruteforce/internal/adapter/outbound/redis"
	"anti-bruteforce/internal/config"
	"anti-bruteforce/internal/usecase"
	"github.com/redis/go-redis/v9"
)

// Module is the fx module: service registration and app wiring.
var Module = fx.Options(
	fx.Provide(config.Load),
	fx.Provide(memory.NewStore),
	fx.Provide(NewRedisClient),
	fx.Provide(
		fx.Annotate(
			NewNaiveLoginChecker,
			fx.ResultTags(`name:"naiveChecker"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewPessimisticLoginChecker,
			fx.ResultTags(`name:"pessimisticChecker"`),
		),
	),
	fx.Provide(NewHTTPDeps),
	fx.Provide(httpadapter.New),
	fx.Invoke(printBanner),
	fx.Invoke(registerHTTPServerLifecycle),
)

func printBanner(cfg *config.Config) {
	log.Printf("🚀 Anti-Bruteforce Vault")
	log.Printf("📝 Port: %s, Host: %s, Redis: %s", cfg.Port, cfg.Host, cfg.RedisURL)
	log.Printf("   Rate limit: %d req / %ds", cfg.RateLimitMax, cfg.RateLimitWindow)
}

// NewRedisClient builds a Redis client from configuration.
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opts), nil
}

// NewNaiveLoginChecker creates the LoginChecker for the naive (in-memory) strategy.
func NewNaiveLoginChecker(cfg *config.Config, store *memory.Store) usecase.LoginChecker {
	return usecase.NewLoginChecker(store, cfg.RateLimitMax, cfg.RateLimitWindow)
}

// NewPessimisticLoginChecker creates the LoginChecker for the pessimistic (Redis lock) strategy.
func NewPessimisticLoginChecker(cfg *config.Config, client *redis.Client) usecase.LoginChecker {
	limiter := redisadapter.NewPessimisticLimiter(client)
	return usecase.NewLoginChecker(limiter, cfg.RateLimitMax, cfg.RateLimitWindow)
}

// httpDepsParams are the dependencies for NewHTTPDeps (for fx.In).
type httpDepsParams struct {
	fx.In

	Config             *config.Config
	NaiveChecker       usecase.LoginChecker `name:"naiveChecker"`
	PessimisticChecker usecase.LoginChecker `name:"pessimisticChecker"`
}

// NewHTTPDeps builds HTTP adapter deps from config and registered checkers.
func NewHTTPDeps(p httpDepsParams) *httpadapter.Deps {
	return &httpadapter.Deps{
		Config:             p.Config,
		NaiveChecker:       p.NaiveChecker,
		PessimisticChecker: p.PessimisticChecker,
	}
}

// registerHTTPServerLifecycle registers HTTP server start/stop in fx.Lifecycle.
func registerHTTPServerLifecycle(lc fx.Lifecycle, srv *httpadapter.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.Start(); err != nil {
					log.Printf("HTTP server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Stop()
		},
	})
}
