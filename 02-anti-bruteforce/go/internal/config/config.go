package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	Host            string
	RedisURL        string
	RateLimitMax    int
	RateLimitWindow int64
	LogLevel        string
}

func Load() *Config {
	if err := godotenv.Load("../.env"); err != nil {
		godotenv.Load()
		log.Println("Warning: no .env file found, using defaults")
	}

	return &Config{
		Port:            getEnv("PORT", "3000"),
		Host:            getEnv("HOST", "0.0.0.0"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		RateLimitMax:    getEnvAsInt("RATE_LIMIT_MAX", 5),
		RateLimitWindow: getEnvAsInt64("RATE_LIMIT_WINDOW", 60),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
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

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}
