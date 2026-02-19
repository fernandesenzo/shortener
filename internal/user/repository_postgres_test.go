package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/testutil"
	"github.com/fernandesenzo/shortener/internal/user"
	_ "github.com/lib/pq"
)

func TestPostgresRepository_Save(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := user.NewPostgresRepository(db)
	ctx := context.Background()

	t.Run("Save User", func(t *testing.T) {
		tests := []struct {
			name    string
			user    *domain.User
			wantErr error
		}{
			{
				name: "Success saving user",
				user: &domain.User{
					Nickname:     "enzo_fernandes",
					PasswordHash: "hashed_pass_123",
				},
				wantErr: nil,
			},
			{
				name: "fail on duplicate nickname",
				user: &domain.User{
					Nickname:     "enzo_fernandes",
					PasswordHash: "another_hash",
				},
				wantErr: user.ErrRecordAlreadyExists,
			},
			{
				name: "success saving another user",
				user: &domain.User{
					Nickname:     "another_user",
					PasswordHash: "hashed_pass_456",
				},
				wantErr: nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Save(ctx, tt.user)

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.wantErr == nil {
					if tt.user.ID == "" {
						t.Error("Save() expected a generated ID, got empty string")
					}
					if tt.user.CreatedAt.IsZero() {
						t.Error("Save() expected a generated CreatedAt, got zero time")
					}
				}
			})
		}
	})
}
