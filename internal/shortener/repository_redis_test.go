package shortener

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/redis/go-redis/v9"
)

func TestRedisRepository_Table(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)
	ctx := context.Background()

	t.Run("Save", func(t *testing.T) {
		tests := []struct {
			name    string
			link    *domain.TemporaryLink
			ttl     time.Duration
			wantErr bool
		}{
			{
				name:    "save new temporary link",
				link:    &domain.TemporaryLink{Code: "abc", OriginalURL: "https://test.com"},
				ttl:     time.Hour,
				wantErr: false,
			},
			{
				name:    "overwrite existing link (Reset TTL)",
				link:    &domain.TemporaryLink{Code: "abc", OriginalURL: "https://new-url.com"},
				ttl:     time.Minute,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Save(ctx, tt.link, tt.ttl)
				if (err != nil) != tt.wantErr {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				}

				val, _ := s.Get(linkPrefix + tt.link.Code)
				if val != tt.link.OriginalURL {
					t.Errorf("Value mismatch: got %v, want %v", val, tt.link.OriginalURL)
				}
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		_ = s.Set("link:exists", "https://exists.com")

		tests := []struct {
			name    string
			code    string
			wantURL string
			wantErr error
		}{
			{
				name:    "link exists",
				code:    "exists",
				wantURL: "https://exists.com",
				wantErr: nil,
			},
			{
				name:    "link does not exist",
				code:    "notfound",
				wantURL: "",
				wantErr: ErrRecordNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.Get(ctx, tt.code)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr == nil {
					if got.OriginalURL != tt.wantURL {
						t.Errorf("Get() URL = %v, want %v", got.OriginalURL, tt.wantURL)
					}
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		_ = s.Set(linkPrefix+"delete_me", "https://todelete.com")

		tests := []struct {
			name    string
			code    string
			wantErr error
		}{
			{
				name:    "delete existing link",
				code:    "delete_me",
				wantErr: nil,
			},
			{
				name:    "delete non-existing link (Redis does not error on this)",
				code:    "ghost_code",
				wantErr: nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Delete(ctx, tt.code)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("delete() error = %v, want %v", err, tt.wantErr)
				}

				if s.Exists(linkPrefix + tt.code) {
					t.Errorf("expected key %v to be deleted from redis", linkPrefix+tt.code)
				}
			})
		}
	})
}
