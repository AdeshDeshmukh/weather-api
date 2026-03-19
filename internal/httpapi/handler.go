package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"weatherapi/internal/weather"
)

type Handler struct {
	weatherService *weather.Service
}

func NewHandler(weatherService *weather.Service) *Handler {
	return &Handler{weatherService: weatherService}
}

func (h *Handler) Routes(rps float64, burst int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/weather/hardcoded", h.hardcodedWeather)
	mux.HandleFunc("/weather", h.weather)

	limiter := newIPRateLimiter(rps, burst)
	return limiter.Middleware(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) hardcodedWeather(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"city":          "london",
		"source":        "hardcoded",
		"temperature_c": 17,
		"conditions":    "Partly Cloudy",
	})
}

func (h *Handler) weather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'city' is required"})
		return
	}

	result, err := h.weatherService.GetWeather(r.Context(), city)
	if err != nil {
		if errors.Is(err, weather.ErrCityNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("city '%s' not found", city)})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch weather from provider"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
