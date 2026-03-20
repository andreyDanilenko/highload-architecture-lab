package http

import (
	"anti-bruteforce/internal/config"
	"anti-bruteforce/internal/usecase"
)

type Deps struct {
	Config             *config.Config
	NaiveChecker       usecase.LoginChecker // POST /login, POST /resource/naive
	PessimisticChecker usecase.LoginChecker // POST /resource/pessimistic
	OptimisticChecker  usecase.LoginChecker // POST /resource/optimistic
	VaultChecker       usecase.LoginChecker // POST /vault/login
}
