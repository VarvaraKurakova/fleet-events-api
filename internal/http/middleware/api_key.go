package middleware

import (
	"encoding/json"
	"net/http"
)

const apiKeyHeader = "X-API-Key"

type errorResponse struct {
	Error string `json:"error"`
}

func APIKey(expectedAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedAPIKey == "" {
				writeMiddlewareError(w, http.StatusInternalServerError, "api key is not configured")
				return
			}

			actualAPIKey := r.Header.Get(apiKeyHeader)
			if actualAPIKey == "" {
				writeMiddlewareError(w, http.StatusUnauthorized, "missing api key")
				return
			}

			if actualAPIKey != expectedAPIKey {
				writeMiddlewareError(w, http.StatusUnauthorized, "invalid api key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeMiddlewareError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: message,
	})
}
