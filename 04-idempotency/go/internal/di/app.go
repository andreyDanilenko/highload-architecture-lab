package di

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	httpadapter "idempotency/internal/adapter/inbound/http"
	"idempotency/internal/adapter/inbound/http/handlers"
	"idempotency/internal/usecase"
	"idempotency/pkg/idempotency"
)

// Config читается в main: сборка зависимостей не меняет глобальные flags и не завершает процесс.
type Config struct {
	Provider         string
	Addr             string
	RedisAddr        string
	RedisPrefix      string
	LeaseTTL         time.Duration
	ResultTTL        time.Duration
	CompleteTimeout  time.Duration
	MaxBodyBytes     int64
	MaxResponseBytes int64
	MaxEntries       int
	PaymentDelay     time.Duration
}

func DefaultConfig() Config {
	return Config{Provider: "memory", Addr: ":8080", RedisAddr: "localhost:6379", RedisPrefix: "lab04:idempotency:", LeaseTTL: 30 * time.Second, ResultTTL: 24 * time.Hour, CompleteTimeout: 2 * time.Second, MaxBodyBytes: 64 << 10, MaxResponseBytes: 64 << 10, MaxEntries: 10000}
}
func (c Config) Validate() error {
	switch c.Provider {
	case "noop", "memory", "redis", "advanced":
	default:
		return fmt.Errorf("unknown provider %q (use noop, memory, redis or advanced)", c.Provider)
	}
	if c.Addr == "" || c.MaxEntries <= 0 || c.PaymentDelay < 0 || c.PaymentDelay > 10*time.Second {
		return errors.New("address and positive capacity are required; payment delay must be between 0 and 10s")
	}
	if c.LeaseTTL < time.Millisecond || c.ResultTTL < time.Millisecond || c.CompleteTimeout < time.Millisecond {
		return errors.New("lease, retention and completion timeout must be at least 1ms")
	}
	if c.MaxBodyBytes <= 0 || c.MaxResponseBytes <= 0 || c.MaxBodyBytes > 64<<20 || c.MaxResponseBytes > 64<<20 {
		return errors.New("body limits must be in 1..64 MiB")
	}
	if (c.Provider == "redis" || c.Provider == "advanced") && (c.RedisAddr == "" || c.RedisPrefix == "") {
		return errors.New("Redis address and key prefix are required")
	}
	return nil
}

// App — единственная точка сборки и владения Redis-клиентом.
type App struct {
	Server      *httpadapter.Server
	redisClient *redis.Client
}

func NewApp(config Config) (*App, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	var provider idempotency.Provider
	var client *redis.Client
	var ready func(context.Context) error
	switch config.Provider {
	case "noop":
		provider = idempotency.NewNoopProvider()
	case "memory":
		provider = idempotency.NewMemoryProvider(config.MaxEntries)
	case "redis", "advanced":
		// Не повторяем мутацию незаметно после потери ответа Redis: её исход может быть неизвестен.
		client = redis.NewClient(&redis.Options{Addr: config.RedisAddr, MaxRetries: -1, DialTimeout: 2 * time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second, PoolTimeout: 2 * time.Second, ContextTimeoutEnabled: true})
		ready = func(ctx context.Context) error { return client.Ping(ctx).Err() }
		if config.Provider == "redis" {
			provider = idempotency.NewRedisProvider(client, config.RedisPrefix)
		} else {
			provider = idempotency.NewAdvancedProvider(client, config.RedisPrefix)
		}
	}
	middleware, err := idempotency.NewHTTPMiddleware(provider, idempotency.HTTPOptions{LeaseTTL: config.LeaseTTL, ResultTTL: config.ResultTTL, CompleteTimeout: config.CompleteTimeout, MaxBodyBytes: config.MaxBodyBytes, MaxResponseBytes: config.MaxResponseBytes, Logger: slog.Default()})
	if err != nil {
		if client != nil {
			_ = client.Close()
		}
		return nil, err
	}
	// Журнал тоже ограничен: заполнение возвращает явную ошибку, а не бесконечно наращивает память.
	payments := usecase.NewPaymentService(config.PaymentDelay, config.MaxEntries)
	handler := handlers.NewPaymentHandler(payments)
	return &App{Server: httpadapter.New(config.Addr, handler, middleware, ready), redisClient: client}, nil
}
func (a *App) Close() error {
	if a.redisClient != nil {
		return a.redisClient.Close()
	}
	return nil
}
