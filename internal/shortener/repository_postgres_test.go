package shortener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
	"github.com/fernandesenzo/shortener/internal/testutil"
	_ "github.com/lib/pq"
)

func TestPostgresRepository(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := shortener.NewPostgresRepository(db)
	ctx := context.Background()

	query := `
			INSERT INTO users (nickname, password_hash)
			VALUES ($1, $2)
			RETURNING id
			`
	var userID string
	err := db.QueryRow(query, "seeduser", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("error inserting seed user: %v", err)
	}

	t.Run("Save", func(t *testing.T) {
		tests := []struct {
			name    string
			link    *domain.PermanentLink
			wantErr error
		}{
			{
				name: "Success saving new link",
				link: &domain.PermanentLink{
					Code:        "654321",
					OriginalURL: "https://github.com",
					UserID:      userID,
				},
				wantErr: nil,
			},
			{
				name: "Fail on duplicate code",
				link: &domain.PermanentLink{
					Code:        "654321",
					OriginalURL: "https://another.com",
					UserID:      userID,
				},
				wantErr: shortener.ErrRecordAlreadyExists,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Save(ctx, tt.link)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		seed := &domain.PermanentLink{
			Code:        "found1",
			OriginalURL: "https://google.com",
			UserID:      userID,
		}
		_ = repo.Save(ctx, seed)

		tests := []struct {
			name    string
			code    string
			wantURL string
			wantErr error
		}{
			{
				name:    "Found existing link",
				code:    "found1",
				wantURL: "https://google.com",
				wantErr: nil,
			},
			{
				name:    "Link not found",
				code:    "000000",
				wantURL: "",
				wantErr: shortener.ErrRecordNotFound,
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
					if got.GetOriginalURL() != tt.wantURL {
						t.Errorf("Get() URL = %v, want %v", got.GetOriginalURL(), tt.wantURL)
					}
				}
			})
		}
	})
	t.Run("Exists", func(t *testing.T) {
		tests := []struct {
			name       string
			code       string
			wantExists bool
			wantErr    error
		}{
			{
				name:       "Link exists in database",
				code:       "found1",
				wantExists: true,
				wantErr:    nil,
			},
			{
				name:       "Link code from Save test exists",
				code:       "654321",
				wantExists: true,
				wantErr:    nil,
			},
			{
				name:       "Link does not exist",
				code:       "not_real_code",
				wantExists: false,
				wantErr:    nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				exists, err := repo.Exists(ctx, tt.code)

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if exists != tt.wantExists {
					t.Errorf("Exists() = %v, want %v", exists, tt.wantExists)
				}
			})
		}
	})
}
