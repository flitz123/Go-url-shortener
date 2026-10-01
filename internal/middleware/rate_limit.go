package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientWindow struct {
	started  time.Time
	requests int
}

var (
	mu      sync.Mutex
	clients = make(map[string]clientWindow)
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		now := time.Now()

		mu.Lock()
		window := clients[ip]
		if now.Sub(window.started) >= time.Minute {
			window = clientWindow{started: now}
		}
		window.requests++
		clients[ip] = window
		if window.requests > 120 {
			mu.Unlock()
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		if len(clients) > 10000 {
			for address, entry := range clients {
				if now.Sub(entry.started) >= 2*time.Minute {
					delete(clients, address)
				}
			}
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
