package shortener_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

func TestServiceShorten(t *testing.T) {
	tests := []struct {
		name            string
		originalURL     string
		mockCollisions  int
		mockShouldError bool
		expectedError   error
	}{
		{
			name:            "Success",
			originalURL:     "https://google.com",
			mockCollisions:  0,
			mockShouldError: false,
			expectedError:   nil,
		},
		{
			name:            "Success with collisions",
			originalURL:     "https://google.com",
			mockCollisions:  3,
			mockShouldError: false,
			expectedError:   nil,
		},
		{
			name:            "Error by exhausting",
			originalURL:     "https://google.com",
			mockCollisions:  10,
			mockShouldError: false,
			expectedError:   domain.ErrLinkCreationFailed,
		},
		{
			name:            "Error by fatal",
			originalURL:     "https://google.com",
			mockCollisions:  0,
			mockShouldError: true,
			expectedError:   domain.ErrLinkCreationFailed,
		},
		{
			name:            "Bad URL error",
			originalURL:     "something",
			mockCollisions:  0,
			mockShouldError: false,
			expectedError:   domain.ErrInvalidURL,
		},
	}
	{
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				repo := &MockRepository{}
				repo.SetCollisionCounter(tt.mockCollisions)
				repo.SetShouldError(tt.mockShouldError)

				service := shortener.NewService(repo)
				_, err := service.Shorten(context.Background(), tt.originalURL)

				if !errors.Is(err, tt.expectedError) {
					t.Errorf("Test %s failed: expected error %v, got %v", tt.name, tt.expectedError, err)
				}
			})
		}
	}
}

func TestServiceGet(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		setupLink     *domain.Link
		shouldError   bool
		expectedError error
	}{
		{
			name:          "Success",
			code:          "abcdef",
			setupLink:     &domain.Link{Code: "abcdef", OriginalURL: "https://google.com"},
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
					_ = repo.Save(context.Background(), tt.setupLink)
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
