package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"idempotency/internal/di"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config := di.DefaultConfig()
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.RedisAddr = addr
	}
	flag.StringVar(&config.Provider, "provider", config.Provider, "noop | memory | redis | advanced")
	flag.StringVar(&config.Addr, "addr", config.Addr, "HTTP listen address")
	flag.StringVar(&config.RedisAddr, "redis-addr", config.RedisAddr, "Redis address (or REDIS_ADDR)")
	flag.StringVar(&config.RedisPrefix, "redis-prefix", config.RedisPrefix, "Redis key namespace shared by matching replicas")
	flag.DurationVar(&config.LeaseTTL, "lease-ttl", config.LeaseTTL, "ownership lease; expiry behavior depends on provider")
	flag.DurationVar(&config.ResultTTL, "result-ttl", config.ResultTTL, "completed-response retention window")
	flag.DurationVar(&config.CompleteTimeout, "complete-timeout", config.CompleteTimeout, "detached completion timeout")
	flag.Int64Var(&config.MaxBodyBytes, "max-body-bytes", config.MaxBodyBytes, "maximum request body bytes")
	flag.Int64Var(&config.MaxResponseBytes, "max-response-bytes", config.MaxResponseBytes, "maximum stored response bytes including allowed headers")
	flag.IntVar(&config.MaxEntries, "max-entries", config.MaxEntries, "memory-provider and per-process effect-ledger capacity")
	flag.DurationVar(&config.PaymentDelay, "payment-delay", config.PaymentDelay, "deterministic lab operation delay (0..10s)")
	flag.Parse()
	app, err := di.NewApp(config)
	if err != nil {
		return err
	}
	defer app.Close()

	// Шаг 1: SIGTERM/SIGINT закрывают приём новых запросов, сохраняя время для текущих.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	failures := make(chan error, 1)
	go func() { failures <- app.Server.Start() }()
	slog.Info("idempotency lab listening", "provider", config.Provider, "addr", config.Addr)
	select {
	case err := <-failures:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}
	// Шаг 2: Redis закрывается после остановки HTTP, чтобы завершения ещё могли сохранить ответ.
	shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(shutdown); err != nil {
		_ = app.Server.Close()
		return err
	}
	return nil
}
