package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"idempotency/internal/business"
	"idempotency/internal/handlers"
	"idempotency/pkg/idempotency"
)

func main() {
	// Флаг для выбора провайдера (можно заменить на переменную окружения)
	providerType := flag.String("provider", "noop", "idempotency provider: noop, memory, redis")
	flag.Parse()

	// Инициализация провайдера
	var provider idempotency.Provider
	switch *providerType {
	case "memory":
		provider = idempotency.NewMemoryProvider(60 * time.Second)
		log.Println("Using MemoryProvider")
	case "redis":
		// TODO: subtask 3
		log.Fatal("RedisProvider not implemented yet")
	default:
		provider = idempotency.NewNoopProvider()
		log.Println("Using NoopProvider (no idempotency)")
	}

	// Инициализация зависимостей
	idempotencyMw := idempotency.NewMiddleware(provider, 60*time.Second)
	paymentService := business.NewPaymentService()
	paymentHandler := handlers.NewPaymentHandler(paymentService, idempotencyMw)

	// Роутинг
	http.HandleFunc("POST /api/v1/payments", paymentHandler.CreatePayment)

	// Запуск
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
