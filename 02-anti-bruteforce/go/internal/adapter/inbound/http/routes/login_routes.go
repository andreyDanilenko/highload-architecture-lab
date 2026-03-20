package routes

import (
	"net/http"

	"anti-bruteforce/internal/adapter/inbound/http/handlers"
	"anti-bruteforce/internal/usecase"
)

func registerLoginRoutes(
	mux *http.ServeMux,
	naive, pessimistic, optimistic, vault usecase.LoginChecker,
) {
	login := handlers.NewLoginHandlers(naive, pessimistic, optimistic, vault)

	mux.HandleFunc("/login", login.Login)
	mux.HandleFunc("/vault/login", login.VaultLogin)
}

