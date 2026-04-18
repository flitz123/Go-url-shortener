package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"go-url-shortener/internal/cache"
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/metrics"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
)

func main() {
	metrics.Init()

	db, err := sql.Open("postgres", "postgres://user:password@db:5432/urlshort?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresRepo(db)
	cache := cache.NewRedis()
	handler := handlers.New(repo, cache)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router.Setup(handler),
	}

	log.Println("Server running on port 8080")
	log.Fatal(server.ListenAndServe())
}
