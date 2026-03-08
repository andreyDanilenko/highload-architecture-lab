.PHONY: infra-up infra-down infra-logs infra-ps infra-reset reset-db run-node run-go

# Compose file dir so volume ./init.sql always resolves correctly (avoid "mount directory onto file")
COMPOSE_DIR := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

# Shared infrastructure (Postgres + Redis) for Node and Go
infra-up:
	@test -f $(COMPOSE_DIR)init.sql || (echo "❌ $(COMPOSE_DIR)init.sql not found. Run from 01-atomic-inventory." && exit 1)
	@test -f $(COMPOSE_DIR).env || (cp $(COMPOSE_DIR)env.example $(COMPOSE_DIR).env && echo "Created $(COMPOSE_DIR).env from env.example (set DB_PASSWORD etc.)")
	cd $(COMPOSE_DIR) && docker-compose up -d
	@echo "Waiting for services..."
	@sleep 3
	@echo "✅ PostgreSQL: localhost:5432"
	@echo "✅ Redis: localhost:6379"

infra-down:
	cd $(COMPOSE_DIR) && docker-compose down

infra-logs:
	cd $(COMPOSE_DIR) && docker-compose logs -f

infra-ps:
	cd $(COMPOSE_DIR) && docker-compose ps

infra-reset:
	cd $(COMPOSE_DIR) && docker-compose down -v
	cd $(COMPOSE_DIR) && docker-compose up -d

reset-db:
	./scripts/reset-db.sh

run-node:
	cd node && $(MAKE) dev

run-go:
	@echo "Go implementation not yet available. Use: cd implementation-go && go run ."
	@exit 1
