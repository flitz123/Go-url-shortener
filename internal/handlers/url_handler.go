package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"go-url-shortener/internal/cache"
	"go-url-shortener/internal/models"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service"
)

type Handler struct {
	repo  *repository.PostgresRepo
	cache *cache.RedisCache
}

func NewHandler(repo *repository.PostgresRepo, cache *cache.RedisCache) *Handler {
	return &Handler{repo: repo, cache: cache}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	code := service.GenerateCode(req.URL)
	url := &models.URL{
		Code:     code,
		Original: req.URL,
		Clicks:   0,
	}

	h.repo.Save(url)
	h.cache.Set(code, req.URL)

	json.NewEncoder(w).Encode(url)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	val, err := h.cache.Get(code)
	if err == nil {
		http.Redirect(w, r, val, http.StatusFound)
		return
	}

	url, err := h.repo.Get(code)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}

	h.cache.Set(code, url.Original)
	h.repo.Increment(code)

	http.Redirect(w, r, url.Original, http.StatusFound)
}
