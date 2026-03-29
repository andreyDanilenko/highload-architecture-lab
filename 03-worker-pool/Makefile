.PHONY: dev build run clean deps

APP_NAME=worker-pool
MAIN_FILE=cmd/server/main.go

deps:
	cd go && go mod download && go mod verify

dev: deps
	cd go && go run $(MAIN_FILE)

build: deps
	cd go && go build -o bin/$(APP_NAME) $(MAIN_FILE)

run: build
	cd go && ./bin/$(APP_NAME)

clean:
	rm -rf go/bin/
