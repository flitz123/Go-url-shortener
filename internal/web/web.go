package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed assets/*
var assets embed.FS

func Handler(fallback http.Handler) http.Handler {
	files, _ := fs.Sub(assets, "assets")
	fileServer := http.FileServer(http.FS(files))
	static := http.StripPrefix("/assets/", fileServer)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			static.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/" {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				w.Header().Set("Allow", "GET, HEAD")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			page, err := assets.ReadFile("assets/index.html")
			if err != nil {
				http.Error(w, "Dashboard unavailable", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			if r.Method == http.MethodGet {
				_, _ = w.Write(page)
			}
			return
		}
		fallback.ServeHTTP(w, r)
	})
}
