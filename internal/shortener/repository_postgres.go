package shortener

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db}
}

func (r *PostgresRepository) Save(ctx context.Context, link *domain.Link) error {
	query := `INSERT INTO links (code, original_url) VALUES ($1,$2) RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, link.Code, link.OriginalURL).Scan(&link.ID, &link.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrRecordAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, code string) (*domain.Link, error) {
	query := `SELECT id, code, original_url, created_at FROM links WHERE code = $1`

	var link domain.Link
	err := r.db.QueryRowContext(ctx, query, code).Scan(&link.ID, &link.Code, &link.OriginalURL, &link.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &link, nil
}

// TODO: create tests for this function
func (r *PostgresRepository) PruneExpired(ctx context.Context, ageLimit time.Duration) error {
	cutoff := time.Now().Add(-ageLimit)
	query := `DELETE FROM links WHERE created_at < $1`
	_, err := r.db.ExecContext(ctx, query, cutoff)
	return err
}
