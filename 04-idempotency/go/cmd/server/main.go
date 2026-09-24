package main

import (
	"log"

	"idempotency/internal/di"
)

func main() {
	app := di.NewApp()
	log.Fatal(app.Server.Start())
}
