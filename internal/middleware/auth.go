package middleware

import (
	"net/http"
	"strings"

	"go-url-shortener/internal/auth"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		tokenStr := strings.TimePrefix(header, "Bearer")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", 401)
			return
		}

		_, err := auth.ValidationToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token", 401)
			return
		}

		next.ServeHTTP(w, r)
	})
}
