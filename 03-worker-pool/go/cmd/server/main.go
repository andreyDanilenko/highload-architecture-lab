package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"worker-pool/internal/di"
)

func main() {
	app := di.NewApp()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Server.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("shutdown signal: %v", sig)
	case err := <-errCh:
		if err != nil {
			log.Printf("server stopped: %v", err)
		}
	}

	if err := app.Server.Stop(); err != nil {
		log.Printf("http close: %v", err)
	}
	app.StopPools()
}
