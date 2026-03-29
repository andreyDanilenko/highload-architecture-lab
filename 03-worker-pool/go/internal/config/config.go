package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Host               string
	LogLevel           string
	Workers            int
	QueueSize          int
	AdvancedWorkers    int
	AdvancedQueueSize  int
	TaskSimulateMs     int
	NaiveMode          string // simulate | ffmpeg | upload (real multipart → transcode)
	FFmpegPath         string
	NaiveFFmpegWorkDir string
	NaiveFFmpegInput   string
	NaiveFFmpegCopy    bool
	PoolTaskMode       string // simulate | ffmpeg | upload
	PoolFFmpegWorkDir  string
	VideoUploadDir     string
	VideoOutputDir     string
	UploadMaxMB        int
	HTTPReadTimeoutSec int
	HTTPWriteTimeoutSec int
}

func Load() *Config {
	if err := godotenv.Load("../.env"); err != nil {
		godotenv.Load()
		log.Println("Warning: no .env file found, using defaults")
	}

	return &Config{
		Port:               getEnv("PORT", "3000"),
		Host:               getEnv("HOST", "0.0.0.0"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		Workers:            getEnvAsInt("WORKER_POOL_WORKERS", 4),
		QueueSize:          getEnvAsInt("WORKER_POOL_QUEUE_SIZE", 32),
		AdvancedWorkers:    getEnvAsInt("WORKER_POOL_ADVANCED_WORKERS", 8),
		AdvancedQueueSize:  getEnvAsInt("WORKER_POOL_ADVANCED_QUEUE_SIZE", 64),
		TaskSimulateMs:     getEnvAsInt("TASK_SIMULATE_MS", 50),
		NaiveMode:          getEnv("NAIVE_MODE", "simulate"),
		FFmpegPath:         getEnv("FFMPEG_PATH", "ffmpeg"),
		NaiveFFmpegWorkDir: getEnv("NAIVE_FFMPEG_WORKDIR", "tmp/naive-ffmpeg"),
		NaiveFFmpegInput:   getEnv("NAIVE_FFMPEG_INPUT", ""),
		NaiveFFmpegCopy:    getEnvAsBool("NAIVE_FFMPEG_STREAM_COPY", false),
		PoolTaskMode:        getEnv("POOL_TASK_MODE", "simulate"),
		PoolFFmpegWorkDir:   getEnv("POOL_FFMPEG_WORKDIR", "tmp/pool-ffmpeg"),
		VideoUploadDir:      getEnv("VIDEO_UPLOAD_DIR", "tmp/video-uploads"),
		VideoOutputDir:      getEnv("VIDEO_OUTPUT_DIR", "tmp/video-outputs"),
		UploadMaxMB:         getEnvAsInt("UPLOAD_MAX_MB", 200),
		HTTPReadTimeoutSec:  getEnvAsInt("HTTP_READ_TIMEOUT_SEC", 600),
		HTTPWriteTimeoutSec: getEnvAsInt("HTTP_WRITE_TIMEOUT_SEC", 600),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}
