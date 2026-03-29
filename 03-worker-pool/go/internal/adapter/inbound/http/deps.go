package http

import (
	"worker-pool/internal/config"
	"worker-pool/internal/usecase"
)

// Deps aggregates inbound HTTP dependencies (plain DI).
type Deps struct {
	Config   *config.Config
	Naive    usecase.TaskDispatcher
	Bounded  usecase.TaskDispatcher
	Reliable usecase.TaskDispatcher
	Advanced usecase.TaskDispatcher
}
