package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"go-url-shortener/internal/cache"
	"go-url-shortener/internal/handlers"
	_ "go-url-shortener/internal/metrics"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://user:password@localhost:5432/urlshort?sslmode=disable"
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewPostgresRepo(db)
	if err := repo.EnsureSchema(); err != nil {
		log.Fatalf("initialize database schema: %v", err)
	}
	cache := cache.NewRedis()
	handler := handlers.NewHandler(repo, cache)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router.Setup(handler),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Server running on port 8080")
	log.Fatal(server.ListenAndServe())
}
