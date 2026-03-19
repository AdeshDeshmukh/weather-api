package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"weatherapi/internal/config"
	"weatherapi/internal/httpapi"
	"weatherapi/internal/weather"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis not reachable, caching may fail: %v", err)
	}

	weatherService := weather.NewService(
		&http.Client{Timeout: 10 * time.Second},
		redisClient,
		cfg.VisualCrossingBaseURL,
		cfg.VisualCrossingAPIKey,
		cfg.CacheTTL,
	)

	handler := httpapi.NewHandler(weatherService)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler.Routes(cfg.RateLimitRPS, cfg.RateLimitBurst),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("weather API listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
