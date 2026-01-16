package shortener_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresRepoSaveAndGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := shortener.NewPostgresRepository(db)

	link := &domain.Link{
		Code:        "123456",
		OriginalURL: "https://google.com",
		CreatedAt:   time.Now().Truncate(time.Second),
	}

	t.Run("save new link", func(t *testing.T) {
		err := repo.Save(link)
		if err != nil {
			t.Fatalf("failed to save link: %v", err)
		}
	})

	t.Run("get existing link", func(t *testing.T) {
		savedLink, err := repo.Get(link.Code)
		if err != nil {
			t.Fatalf("failed to get saved link: %v", err)
		}
		if savedLink.OriginalURL != link.OriginalURL {
			t.Fatalf("saved link has wrong original url")
		}
	})
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	migrationPath := filepath.Join("..", "..", "db", "migrations", "0001_init_schema.up.sql")

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(migrationPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)

	if err != nil {
		t.Fatalf("failed to initialize postgres container: %v", err)
	}

	connectionString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get database connection string: %v", err)
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		t.Fatalf("failed to open database connection: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping database connection: %v", err)
	}

	cleanup := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %v", err)
		}
	}
	return db, cleanup
}
