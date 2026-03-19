package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                 string
	VisualCrossingAPIKey string
	VisualCrossingBaseURL string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	CacheTTL             time.Duration
	RateLimitRPS         float64
	RateLimitBurst       int
}

func Load() (Config, error) {
	cfg := Config{}

	cfg.Port = getEnv("PORT", "8080")
	cfg.VisualCrossingBaseURL = getEnv("VISUAL_CROSSING_BASE_URL", "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline")
	cfg.RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")

	apiKey, ok := os.LookupEnv("VISUAL_CROSSING_API_KEY")
	if !ok || apiKey == "" {
		return cfg, fmt.Errorf("missing required environment variable: VISUAL_CROSSING_API_KEY")
	}
	cfg.VisualCrossingAPIKey = apiKey

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return cfg, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	cfg.RedisDB = redisDB

	cacheTTLHours, err := strconv.Atoi(getEnv("CACHE_TTL_HOURS", "12"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CACHE_TTL_HOURS: %w", err)
	}
	cfg.CacheTTL = time.Duration(cacheTTLHours) * time.Hour

	rateLimitRPS, err := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "5"), 64)
	if err != nil {
		return cfg, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
	}
	cfg.RateLimitRPS = rateLimitRPS

	rateLimitBurst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "10"))
	if err != nil {
		return cfg, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}
	cfg.RateLimitBurst = rateLimitBurst

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
