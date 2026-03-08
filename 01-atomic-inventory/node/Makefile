.PHONY: dev build start clean check

dev:
	npm run dev

build:
	npm run build

start:
	npm run start

clean:
	rm -rf dist

check:
	@curl -s http://localhost:3000/health | head -1 || echo "❌ Server not running"
