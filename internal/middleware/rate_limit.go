package middleware

import (
	"net/http"
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		mu.lock()
		clients[ip]++
		if clients[ip] > 20 {
			mu.Unlock()
			http.Error(w, "Too many requests", 429)
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
