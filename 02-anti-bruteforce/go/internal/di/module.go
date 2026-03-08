package di

import (
	"context"
	"log"

	"go.uber.org/fx"

	httpadapter "anti-bruteforce/internal/adapter/inbound/http"
	"anti-bruteforce/internal/adapter/outbound/memory"
	"anti-bruteforce/internal/config"
	"anti-bruteforce/internal/usecase"
)

// Module is the fx module: service registration and app wiring.
var Module = fx.Options(
	fx.Provide(config.Load),
	fx.Provide(memory.NewStore),
	fx.Provide(NewNaiveLoginChecker),
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

// NewNaiveLoginChecker creates the LoginChecker for the naive (in-memory) strategy.
func NewNaiveLoginChecker(cfg *config.Config, store *memory.Store) usecase.LoginChecker {
	return usecase.NewLoginChecker(store, cfg.RateLimitMax, cfg.RateLimitWindow)
}

// NewHTTPDeps builds HTTP adapter deps from config and registered checkers.
func NewHTTPDeps(cfg *config.Config, naiveChecker usecase.LoginChecker) *httpadapter.Deps {
	return &httpadapter.Deps{
		Config:       cfg,
		NaiveChecker: naiveChecker,
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
