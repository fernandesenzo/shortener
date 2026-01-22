package shortener_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

func TestShortenSuccess(t *testing.T) {
	repo := &MockRepository{}

	service := shortener.NewService(repo)

	originalURL := "https://google.com"

	link, err := service.Shorten(context.Background(), originalURL)

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

func TestGetExistingLink(t *testing.T) {
	repo := &MockRepository{}

	service := shortener.NewService(repo)

	_ = repo.Save(context.Background(), &domain.Link{
		Code:        "123456",
		OriginalURL: "https://google.com",
	})

	_, err := service.Get(context.Background(), "123456")
	if err != nil {
		t.Fatalf("should not return error, returned %v", err)
	}
}

func TestGetNonExistentLink(t *testing.T) {
	repo := &MockRepository{}

	service := shortener.NewService(repo)

	_, err := service.Get(context.Background(), "123456")
	if err == nil {
		t.Error("expected error")
	}
	if !errors.Is(err, domain.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestShortenLongURL(t *testing.T) {
	repo := &MockRepository{}

	service := shortener.NewService(repo)

	url := "https://google.com/" + strings.Repeat("a", 101)

	_, err := service.Shorten(context.Background(), url)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrURLTooLong) {
		t.Errorf("expected ErrURLTooLong, got: %v", err)
	}
}
