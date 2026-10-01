package router

import (
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/middleware"
	"go-url-shortener/internal/web"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Setup(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", h.Shorten)
	mux.HandleFunc("/api/urls", h.Shorten)
	mux.HandleFunc("/api/urls/", h.Lookup)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", web.Handler(http.HandlerFunc(h.Redirect)))

	mux.Handle("/secure", middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected Route"))
	})))

	mux.Handle("/metrics", promhttp.Handler())

	return middleware.Metrics(
		middleware.RateLimit(mux),
	)
}
