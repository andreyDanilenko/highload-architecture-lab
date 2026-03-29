package routes

import (
	"net/http"

	"worker-pool/internal/adapter/inbound/http/handlers"
	"worker-pool/internal/config"
	"worker-pool/internal/usecase"
)

func registerWorkRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	naive, bounded, reliable, advanced usecase.TaskDispatcher,
) {
	h := handlers.NewWorkHandler(cfg, naive, bounded, reliable, advanced)
	mux.HandleFunc("/work/naive", h.Naive)
	mux.HandleFunc("/work/bounded", h.Bounded)
	mux.HandleFunc("/work/reliable", h.Reliable)
	mux.HandleFunc("/work/advanced", h.Advanced)

	mux.HandleFunc("POST /work/naive/upload", h.NaiveUpload)
	mux.HandleFunc("POST /work/bounded/upload", h.BoundedUpload)
	mux.HandleFunc("POST /work/reliable/upload", h.ReliableUpload)
	mux.HandleFunc("POST /work/advanced/upload", h.AdvancedUpload)
	mux.HandleFunc("GET /work/result/{id}", h.VideoResult)
}
