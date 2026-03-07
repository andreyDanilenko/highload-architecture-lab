.PHONY: infra-up infra-down infra-logs infra-ps dev build clean

infra-up:
	docker-compose up -d
	@echo "Waiting for services to be ready..."
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

dev:
	npm run dev

build:
	npm run build

start:
	npm run start

clean:
	rm -rf dist

# Новый helper для проверки подключения
check:
	@echo "Checking connections..."
	@curl -s http://localhost:3000/health | json_pp || echo "❌ Server not running"
