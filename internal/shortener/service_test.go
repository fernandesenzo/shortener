package shortener_test

import (
	"testing"

	"github.com/fernandesenzo/shortener/internal/shortener"
)

func TestShortenSuccess(t *testing.T) {
	repo := &MockRepository{}

	service := shortener.NewService(repo)

	originalURL := "https://google.com"

	link, err := service.Shorten(originalURL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if link.Code == "" {
		t.Error("expected non-empty code")
	}

	if len(link.Code) != 6 {
		t.Errorf("expected length == 6, received %d", len(link.Code))
	}

	if link.OriginalURL != originalURL {
		t.Errorf("expected the same originalURL, should be %s - received %s", originalURL, link.OriginalURL)
	}
}
