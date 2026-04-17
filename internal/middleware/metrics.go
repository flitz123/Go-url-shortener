package middleware

import (
	"go-url-shortener/internal/metrics"
	"net/http"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.Requests.WithLabelValues(r.URL.Path).Inc()
		next.ServeHTTP(w, r)
	})
}
