package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCityNotFound = errors.New("city not found")

type Service struct {
	httpClient *http.Client
	redis      *redis.Client
	baseURL    string
	apiKey     string
	cacheTTL   time.Duration
}

type Response struct {
	City        string         `json:"city"`
	Source      string         `json:"source"`
	RetrievedAt time.Time      `json:"retrieved_at"`
	Data        map[string]any `json:"data"`
}

func NewService(httpClient *http.Client, redisClient *redis.Client, baseURL, apiKey string, cacheTTL time.Duration) *Service {
	return &Service{
		httpClient: httpClient,
		redis:      redisClient,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		cacheTTL:   cacheTTL,
	}
}

func (s *Service) GetWeather(ctx context.Context, city string) (Response, error) {
	normalizedCity := normalizeCity(city)
	if normalizedCity == "" {
		return Response{}, fmt.Errorf("city is required")
	}

	cacheKey := fmt.Sprintf("weather:%s", normalizedCity)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var payload map[string]any
		if unmarshalErr := json.Unmarshal([]byte(cached), &payload); unmarshalErr == nil {
			return Response{
				City:        normalizedCity,
				Source:      "cache",
				RetrievedAt: time.Now().UTC(),
				Data:        payload,
			}, nil
		}
	}

	apiURL := fmt.Sprintf("%s/%s?unitGroup=metric&key=%s&contentType=json", s.baseURL, url.PathEscape(normalizedCity), url.QueryEscape(s.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Response{}, fmt.Errorf("failed to build weather request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("failed to call weather provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return Response{}, ErrCityNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Response{}, fmt.Errorf("weather provider error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("failed to read weather provider response: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return Response{}, fmt.Errorf("invalid weather provider response: %w", err)
	}

	_ = s.redis.Set(ctx, cacheKey, body, s.cacheTTL).Err()

	return Response{
		City:        normalizedCity,
		Source:      "provider",
		RetrievedAt: time.Now().UTC(),
		Data:        payload,
	}, nil
}

func normalizeCity(city string) string {
	return strings.TrimSpace(strings.ToLower(city))
}
