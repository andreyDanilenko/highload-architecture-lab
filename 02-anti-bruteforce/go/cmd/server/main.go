package main

import (
	"go.uber.org/fx"

	"anti-bruteforce/internal/di"
)

func main() {
	fx.New(di.Module).Run()
}
