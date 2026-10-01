package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go-url-shortener/internal/models"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service"
)

var codePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type Handler struct {
	repo  urlRepository
	cache urlCache
}

type urlRepository interface {
	Save(*models.URL) error
	Get(string) (*models.URL, error)
	Increment(string) error
}

type urlCache interface {
	Get(string) (string, error)
	Set(string, string) error
}

func NewHandler(repo urlRepository, cache urlCache) *Handler {
	return &Handler{repo: repo, cache: cache}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL         string `json:"url"`
		CustomAlias string `json:"custom_alias"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || len(req.URL) > 2048 || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Error(w, "Enter a valid HTTP or HTTPS URL", http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(req.CustomAlias)
	if code != "" && !codePattern.MatchString(code) {
		http.Error(w, "Custom aliases must be 3-32 letters, numbers, hyphens, or underscores", http.StatusBadRequest)
		return
	}
	if code == "" {
		code, err = service.GenerateCode()
		if err != nil {
			http.Error(w, "Could not generate a short code", http.StatusInternalServerError)
			return
		}
	}

	url := &models.URL{
		Code:     code,
		Original: req.URL,
		Clicks:   0,
	}

	if err := h.repo.Save(url); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			http.Error(w, "That short code is already in use", http.StatusConflict)
			return
		}
		http.Error(w, "Could not save the short link", http.StatusInternalServerError)
		return
	}
	_ = h.cache.Set(code, req.URL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(struct {
		Code     string `json:"code"`
		Original string `json:"original"`
		ShortURL string `json:"short_url"`
		Clicks   int    `json:"clicks"`
	}{url.Code, url.Original, shortURL(r, code), url.Clicks})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/")
	if !codePattern.MatchString(code) {
		http.NotFound(w, r)
		return
	}

	val, err := h.cache.Get(code)
	if err != nil {
		storedURL, repoErr := h.repo.Get(code)
		if errors.Is(repoErr, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if repoErr != nil {
			http.Error(w, "Could not look up the short link", http.StatusInternalServerError)
			return
		}
		val = storedURL.Original
		_ = h.cache.Set(code, val)
	}

	if err := h.repo.Increment(code); err != nil {
		http.Error(w, "Could not record the visit", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, val, http.StatusFound)
}

func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/api/urls/")
	if !codePattern.MatchString(code) {
		http.NotFound(w, r)
		return
	}

	item, err := h.repo.Get(code)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Could not look up the short link", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Code     string `json:"code"`
		Original string `json:"original"`
		ShortURL string `json:"short_url"`
		Clicks   int    `json:"clicks"`
	}{item.Code, item.Original, shortURL(r, item.Code), item.Clicks})
}

func shortURL(r *http.Request, code string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s", scheme, r.Host, code)
}
