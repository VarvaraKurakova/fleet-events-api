package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type recoverErrorResponse struct {
	Error string `json:"error"`
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error(
						"panic recovered",
						"request_id", GetRequestID(r.Context()),
						"method", r.Method,
						"path", r.URL.Path,
						"panic", recovered,
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					_ = json.NewEncoder(w).Encode(recoverErrorResponse{
						Error: "internal server error",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
