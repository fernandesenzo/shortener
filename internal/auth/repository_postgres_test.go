package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fernandesenzo/shortener/internal/auth"
	"github.com/fernandesenzo/shortener/internal/testutil"
)

func TestPostgresRepository_GetByNickname(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := auth.NewPostgresRepository(db)
	ctx := context.Background()

	t.Run("should return user successfully when nickname exists", func(t *testing.T) {
		nickname := "enzo_test"
		passwordHash := "$2a$10$dummyhash1234567890123"

		_, err := db.ExecContext(ctx, `INSERT INTO users (nickname, password_hash) VALUES ($1, $2)`, nickname, passwordHash)
		if err != nil {
			t.Fatalf("failed to insert test user: %v", err)
		}

		user, err := repo.GetByNickname(ctx, nickname)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user == nil {
			t.Fatalf("expected user, got nil")
		}
		if user.Nickname != nickname {
			t.Errorf("expected nickname %s, got %s", nickname, user.Nickname)
		}
		if user.PasswordHash != passwordHash {
			t.Errorf("expected password hash %s, got %s", passwordHash, user.PasswordHash)
		}
	})

	t.Run("should return ErrRecordNotFound when nickname does not exist", func(t *testing.T) {
		user, err := repo.GetByNickname(ctx, "ghost_user")

		if user != nil {
			t.Errorf("expected nil user, got %+v", user)
		}
		if !errors.Is(err, auth.ErrRecordNotFound) {
			t.Errorf("expected error %v, got %v", auth.ErrRecordNotFound, err)
		}
	})
}
