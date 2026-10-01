package service

import (
	"regexp"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{12}$`).MatchString(code) {
		t.Fatalf("GenerateCode() = %q, want 12 URL-safe characters", code)
	}
}
