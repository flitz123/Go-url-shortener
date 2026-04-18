package router

import (
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/middleware"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Setup(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", h.Shorten)
	mux.HandleFunc("/", h.Redirect)

	mux.Handle("/secure", middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected Route"))
	})))

	mux.Handle("/metrics", promhttp.Handler())

	return middleware.Metrics(
		middleware.RateLimit(mux),
	)
}
