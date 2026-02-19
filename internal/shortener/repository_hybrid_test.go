package shortener_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
	"github.com/fernandesenzo/shortener/internal/testutil"
	"github.com/redis/go-redis/v9"
)

func TestHybridRepository(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	pgRepo := shortener.NewPostgresRepository(db)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	redisRepo := shortener.NewRedisRepository(redisClient)

	hybrid := shortener.NewHybridLinkRepository(pgRepo, redisRepo)
	ctx := context.Background()

	var testUserID string
	err := db.QueryRow(`
		INSERT INTO users (nickname, password_hash) 
		VALUES ('enzo_test', 'hash123') 
		RETURNING id
	`).Scan(&testUserID)
	if err != nil {
		t.Fatalf("error creating seed user for test: %v", err)
	}

	t.Run("Collisions", func(t *testing.T) {
		tests := []struct {
			name        string
			linkType    string //perm or temp
			code        string
			originalURL string
			expectedErr error
		}{
			{
				name:        "PermSave_Success",
				linkType:    "perm",
				code:        "123456",
				originalURL: "https://internacional.com",
				expectedErr: nil,
			},
			{
				name:        "TempSave_Fail_ExistInPostgres",
				linkType:    "temp",
				code:        "123456",
				originalURL: "https://palmeiras.com",
				expectedErr: shortener.ErrRecordAlreadyExists,
			},
			{
				name:        "TempSave_Success",
				linkType:    "temp",
				code:        "321456",
				originalURL: "https://fluminense.com",
				expectedErr: nil,
			},
			{
				name:        "PermSave_Fail_ExistInRedis",
				linkType:    "perm",
				code:        "321456",
				originalURL: "https://fail.com",
				expectedErr: shortener.ErrRecordAlreadyExists,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var err error

				if tt.linkType == "perm" {
					err = hybrid.PermSave(ctx, &domain.PermanentLink{
						Code:        tt.code,
						OriginalURL: tt.originalURL,
						UserID:      testUserID,
					})
				} else {
					err = hybrid.TempSave(ctx, &domain.TemporaryLink{
						Code:        tt.code,
						OriginalURL: tt.originalURL,
					}, time.Hour)
				}

				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected %q, got %q", tt.expectedErr, err)
				}
			})
		}
	})

	t.Run("Get_CacheAside_Flow", func(t *testing.T) {
		code := "ABC123"
		url := "https://teste.com"

		err := hybrid.PermSave(ctx, &domain.PermanentLink{
			Code:        code,
			OriginalURL: url,
			UserID:      testUserID,
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		mr.Del("link:" + code)

		got, err := hybrid.Get(ctx, code)
		if err != nil {
			t.Fatalf("expected to find link, got %v", err)
		}

		if got.GetOriginalURL() != url {
			t.Errorf("got %s, want %s", got.GetOriginalURL(), url)
		}

		if !mr.Exists("link:" + code) {
			t.Error("expected redis to be repopulated after cache miss")
		}
	})
}
