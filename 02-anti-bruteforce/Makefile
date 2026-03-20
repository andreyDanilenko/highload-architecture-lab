.PHONY: dev build run clean deps

APP_NAME=anti-bruteforce
MAIN_FILE=cmd/server/main.go

deps:
	go mod download
	go mod verify

dev: deps
	go run $(MAIN_FILE)

build: deps
	go build -o bin/$(APP_NAME) $(MAIN_FILE)

run: build
	./bin/$(APP_NAME)

clean:
	rm -rf bin/
