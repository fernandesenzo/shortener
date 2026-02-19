package shortener_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

func TestServiceShorten_Validation(t *testing.T) {
	repo := &MockRepository{}
	service := shortener.NewService(repo)

	tests := []struct {
		name        string
		originalURL string
		expectedErr error
	}{
		{
			name:        "Invalid URL",
			originalURL: "invalid",
			expectedErr: domain.ErrInvalidURL,
		},
		{
			name:        "URL too long",
			originalURL: string(make([]byte, 101)),
			expectedErr: domain.ErrURLTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Shorten(context.Background(), tt.originalURL, "")
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestServiceShorten_SaveStrategy(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		expectTemp bool
		expectPerm bool
	}{
		{
			name:       "Unlogged user uses TempSave",
			userID:     "",
			expectTemp: true,
			expectPerm: false,
		},
		{
			name:       "Logged user uses PermSave",
			userID:     "123",
			expectTemp: false,
			expectPerm: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			service := shortener.NewService(repo)

			_, err := service.Shorten(context.Background(), "https://google.com", tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if repo.tempSaveCalled != tt.expectTemp {
				t.Errorf("expected TempSaveCalled=%v, got %v", tt.expectTemp, repo.tempSaveCalled)
			}

			if repo.permSaveCalled != tt.expectPerm {
				t.Errorf("expected PermSaveCalled=%v, got %v", tt.expectPerm, repo.permSaveCalled)
			}
		})
	}
}

func TestServiceShorten_Collisions(t *testing.T) {
	tests := []struct {
		name           string
		mockCollisions int
		expectedErr    error
	}{
		{
			name:           "Succeeds after collisions",
			mockCollisions: 3,
			expectedErr:    nil,
		},
		{
			name:           "Fails after exhausting attempts",
			mockCollisions: 10,
			expectedErr:    domain.ErrLinkCreationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			repo.SetCollisionCounter(tt.mockCollisions)

			service := shortener.NewService(repo)

			_, err := service.Shorten(context.Background(), "https://google.com", "")
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestServiceShorten_RepositoryError(t *testing.T) {
	repo := &MockRepository{}
	repo.SetShouldError(true)

	service := shortener.NewService(repo)

	_, err := service.Shorten(context.Background(), "https://google.com", "123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestServiceGet(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		setupLink     *domain.TemporaryLink
		shouldError   bool
		expectedError error
	}{
		{
			name:          "Success",
			code:          "abcdef",
			setupLink:     &domain.TemporaryLink{Code: "abcdef", OriginalURL: "https://google.com"},
			shouldError:   false,
			expectedError: nil,
		},
		{
			name:          "Not Found",
			code:          "123456",
			setupLink:     nil,
			shouldError:   false,
			expectedError: domain.ErrLinkNotFound,
		},
		{
			name:          "Database error",
			code:          "abcdef",
			setupLink:     nil,
			shouldError:   true,
			expectedError: errors.New("unexpected database error"),
		},
	}
	{
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := &MockRepository{}
				repo.SetShouldError(tt.shouldError)
				if tt.setupLink != nil {
					_ = repo.save(context.Background(), tt.setupLink)
				}
				service := shortener.NewService(repo)
				_, err := service.Get(context.Background(), tt.code)

				// workaround to catch non sentinel error
				//TODO: find a more elegant way to do this, this is ugly
				if err != nil {
					if strings.Contains(err.Error(), tt.expectedError.Error()) {
						return
					}
				}
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("Test %s failed: expected error %v, got %v", tt.name, tt.expectedError, err)
				}
			})
		}
	}
}
