package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-url-shortener/internal/models"
	"go-url-shortener/internal/repository"
)

type memoryRepository struct {
	items      map[string]*models.URL
	saveErr    error
	increments int
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{items: make(map[string]*models.URL)}
}

func (r *memoryRepository) Save(item *models.URL) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	if _, exists := r.items[item.Code]; exists {
		return repository.ErrConflict
	}
	copy := *item
	r.items[item.Code] = &copy
	return nil
}

func (r *memoryRepository) Get(code string) (*models.URL, error) {
	item, ok := r.items[code]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copy := *item
	return &copy, nil
}

func (r *memoryRepository) Increment(code string) error {
	item, ok := r.items[code]
	if !ok {
		return sql.ErrNoRows
	}
	item.Clicks++
	r.increments++
	return nil
}

type memoryCache map[string]string

func (c memoryCache) Get(key string) (string, error) {
	value, ok := c[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return value, nil
}

func (c memoryCache) Set(key, value string) error {
	c[key] = value
	return nil
}

func TestShortenValidatesAndCreatesLink(t *testing.T) {
	repo := newMemoryRepository()
	cache := memoryCache{}
	handler := NewHandler(repo, cache)
	request := httptest.NewRequest(http.MethodPost, "/api/urls", strings.NewReader(`{"url":"https://example.com/a","custom_alias":"launch"}`))
	request.Host = "short.test"
	response := httptest.NewRecorder()

	handler.Shorten(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("Shorten() status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	var result struct {
		Code     string `json:"code"`
		ShortURL string `json:"short_url"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Code != "launch" || result.ShortURL != "http://short.test/launch" {
		t.Fatalf("unexpected response: %+v", result)
	}
	if cache["launch"] != "https://example.com/a" {
		t.Fatalf("cache value = %q, want destination URL", cache["launch"])
	}
}

func TestShortenRejectsInvalidURLAndAliasConflict(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		repo := newMemoryRepository()
		handler := NewHandler(repo, memoryCache{})
		request := httptest.NewRequest(http.MethodPost, "/api/urls", strings.NewReader(`{"url":"javascript:alert(1)"}`))
		response := httptest.NewRecorder()
		handler.Shorten(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("Shorten() status = %d, want %d", response.Code, http.StatusBadRequest)
		}
	})

	t.Run("alias conflict", func(t *testing.T) {
		repo := newMemoryRepository()
		repo.saveErr = repository.ErrConflict
		handler := NewHandler(repo, memoryCache{})
		request := httptest.NewRequest(http.MethodPost, "/api/urls", strings.NewReader(`{"url":"https://example.com","custom_alias":"launch"}`))
		response := httptest.NewRecorder()
		handler.Shorten(response, request)
		if response.Code != http.StatusConflict {
			t.Fatalf("Shorten() status = %d, want %d", response.Code, http.StatusConflict)
		}
	})
}

func TestRedirectCountsCachedVisits(t *testing.T) {
	repo := newMemoryRepository()
	repo.items["launch"] = &models.URL{Code: "launch", Original: "https://example.com"}
	handler := NewHandler(repo, memoryCache{"launch": "https://example.com"})
	request := httptest.NewRequest(http.MethodGet, "/launch", nil)
	response := httptest.NewRecorder()

	handler.Redirect(response, request)
	if response.Code != http.StatusFound {
		t.Fatalf("Redirect() status = %d, want %d", response.Code, http.StatusFound)
	}
	if response.Header().Get("Location") != "https://example.com" {
		t.Fatalf("Location = %q, want destination URL", response.Header().Get("Location"))
	}
	if repo.increments != 1 {
		t.Fatalf("visit increments = %d, want 1", repo.increments)
	}
}
