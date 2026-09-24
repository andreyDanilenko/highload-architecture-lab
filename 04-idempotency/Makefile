GO ?= go
PROVIDER ?= memory
ADDR ?= 127.0.0.1:8084
LAB_REDIS_ADDR ?= 127.0.0.1:6384
COMPOSE = docker compose -f go/docker-compose.yml

.PHONY: run build test test-race test-integration experiment redis-up redis-down

run:
	cd go && GOWORK=off $(GO) run ./cmd/server -addr=$(ADDR) -provider=$(PROVIDER) -redis-addr=$(LAB_REDIS_ADDR)

build:
	cd go && GOWORK=off $(GO) build ./...

test:
	cd go && LAB_REDIS_ADDR= GOWORK=off $(GO) test ./...

test-race:
	cd go && LAB_REDIS_ADDR= GOWORK=off $(GO) test -race ./...

# Реальный Redis обязателен: неверный адрес должен завершить проверку ошибкой, а не skip.
test-integration:
	cd go && LAB_REDIS_ADDR=$(LAB_REDIS_ADDR) GOWORK=off $(GO) test -race -count=1 ./...

experiment:
	cd go && LAB_REDIS_ADDR=$(LAB_REDIS_ADDR) GOWORK=off $(GO) test -count=1 -v ./internal/di -run '^TestExperiment'

redis-up:
	$(COMPOSE) up -d --wait

# Данные сохраняются в volume; неизвестные исходы не исчезают при обычной остановке.
redis-down:
	$(COMPOSE) down
