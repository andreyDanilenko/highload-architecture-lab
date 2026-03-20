package routes

import (
	"net/http"

	"anti-bruteforce/internal/adapter/inbound/http/handlers"
	"anti-bruteforce/internal/usecase"
)

func registerResourceRoutes(
	mux *http.ServeMux,
	naive, pessimistic, optimistic usecase.LoginChecker,
) {
	login := handlers.NewLoginHandlers(naive, pessimistic, optimistic, nil)

	mux.HandleFunc("/resource/naive", login.NaiveResource)
	mux.HandleFunc("/resource/pessimistic", login.PessimisticResource)
	mux.HandleFunc("/resource/optimistic", login.OptimisticResource)
}

