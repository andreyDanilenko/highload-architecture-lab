package di

import (
	"log"
	"os"
	"path/filepath"
	"time"

	httpadapter "worker-pool/internal/adapter/inbound/http"
	"worker-pool/internal/adapter/outbound/pool/advanced"
	"worker-pool/internal/adapter/outbound/pool/bounded"
	"worker-pool/internal/adapter/outbound/pool/naive"
	"worker-pool/internal/adapter/outbound/pool/reliable"
	"worker-pool/internal/config"
	"worker-pool/internal/usecase"
	"worker-pool/internal/work"
)

// App holds the HTTP server and background pools for plain (non-fx) wiring.
type App struct {
	Server *httpadapter.Server
	pools  []usecase.Stopper
}

// NewApp constructs config, outbound pools, HTTP deps, and server in one place.
func NewApp() *App {
	cfg := config.Load()
	simulate := time.Duration(cfg.TaskSimulateMs) * time.Millisecond

	if cfg.NaiveMode == "upload" || cfg.PoolTaskMode == "upload" {
		if err := os.MkdirAll(cfg.VideoUploadDir, 0o755); err != nil {
			log.Printf("VIDEO_UPLOAD_DIR: %v", err)
		}
		if err := os.MkdirAll(cfg.VideoOutputDir, 0o755); err != nil {
			log.Printf("VIDEO_OUTPUT_DIR: %v", err)
		}
	}

	var naiveFF *work.FFmpegSegmentConfig
	switch cfg.NaiveMode {
	case "ffmpeg":
		naiveFF = &work.FFmpegSegmentConfig{
			Bin:        cfg.FFmpegPath,
			WorkDir:    cfg.NaiveFFmpegWorkDir,
			Input:      cfg.NaiveFFmpegInput,
			StreamCopy: cfg.NaiveFFmpegCopy,
		}
		if abs, err := filepath.Abs(naiveFF.WorkDir); err == nil {
			log.Printf("naive ffmpeg (lavfi/url): files under %s", abs)
		}
	case "upload":
		naiveFF = &work.FFmpegSegmentConfig{
			Bin:       cfg.FFmpegPath,
			UploadDir: cfg.VideoUploadDir,
			OutputDir: cfg.VideoOutputDir,
		}
		if absU, err := filepath.Abs(naiveFF.UploadDir); err == nil {
			if absO, err2 := filepath.Abs(naiveFF.OutputDir); err2 == nil {
				log.Printf("naive upload: originals %s → transcoded mp4 %s", absU, absO)
			}
		}
	}
	naiveDispatcher := naive.NewDispatcher(simulate, naiveFF)

	var poolFF *work.FFmpegSegmentConfig
	switch cfg.PoolTaskMode {
	case "ffmpeg":
		poolFF = &work.FFmpegSegmentConfig{
			Bin:        cfg.FFmpegPath,
			WorkDir:    cfg.PoolFFmpegWorkDir,
			Input:      cfg.NaiveFFmpegInput,
			StreamCopy: cfg.NaiveFFmpegCopy,
		}
		if abs, err := filepath.Abs(poolFF.WorkDir); err == nil {
			log.Printf("pool ffmpeg (lavfi/url): files under %s", abs)
		}
	case "upload":
		poolFF = &work.FFmpegSegmentConfig{
			Bin:       cfg.FFmpegPath,
			UploadDir: cfg.VideoUploadDir,
			OutputDir: cfg.VideoOutputDir,
		}
		if absU, err := filepath.Abs(poolFF.UploadDir); err == nil {
			if absO, err2 := filepath.Abs(poolFF.OutputDir); err2 == nil {
				log.Printf("pool upload: originals %s → transcoded mp4 %s", absU, absO)
			}
		}
	}

	boundedPool := bounded.New(cfg.Workers, cfg.QueueSize, simulate, poolFF)
	reliablePool := reliable.New(cfg.Workers, cfg.QueueSize, simulate, poolFF)
	advancedPool := advanced.New(cfg.AdvancedWorkers, cfg.AdvancedQueueSize, simulate, poolFF)

	deps := &httpadapter.Deps{
		Config:   cfg,
		Naive:    naiveDispatcher,
		Bounded:  boundedPool,
		Reliable: reliablePool,
		Advanced: advancedPool,
	}
	srv := httpadapter.New(deps)

	log.Printf("worker-pool: listening on %s:%s (workers=%d queue=%d, advanced workers=%d queue=%d, simulate=%s, naive_mode=%s, pool_task_mode=%s)",
		cfg.Host, cfg.Port,
		cfg.Workers, cfg.QueueSize,
		cfg.AdvancedWorkers, cfg.AdvancedQueueSize,
		simulate,
		cfg.NaiveMode,
		cfg.PoolTaskMode,
	)

	return &App{
		Server: srv,
		pools: []usecase.Stopper{
			naiveDispatcher,
			boundedPool,
			reliablePool,
			advancedPool,
		},
	}
}

// StopPools stops background workers (call after HTTP server is closed).
func (a *App) StopPools() {
	for i := len(a.pools) - 1; i >= 0; i-- {
		a.pools[i].Stop()
	}
}
