package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicRoute(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status %d", w.Code)
	}
}
