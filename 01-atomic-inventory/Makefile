.PHONY: infra-up infra-down infra-logs infra-ps infra-reset reset-db run-node run-go

# Shared infrastructure (Postgres + Redis) for Node and Go
infra-up:
	docker-compose up -d
	@echo "Waiting for services..."
	@sleep 3
	@echo "✅ PostgreSQL: localhost:5432"
	@echo "✅ Redis: localhost:6379"

infra-down:
	docker-compose down

infra-logs:
	docker-compose logs -f

infra-ps:
	docker-compose ps

infra-reset:
	docker-compose down -v
	docker-compose up -d

reset-db:
	./scripts/reset-db.sh

run-node:
	cd implementation-node && $(MAKE) dev

run-go:
	@echo "Go implementation not yet available. Use: cd implementation-go && go run ."
	@exit 1
