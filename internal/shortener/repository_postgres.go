package shortener

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (r *PostgresRepository) Save(ctx context.Context, link *domain.PermanentLink) error {
	query := `INSERT INTO links (code, original_url, user_id) VALUES ($1,$2,$3) RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, link.Code, link.OriginalURL, link.UserID).Scan(&link.ID, &link.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrRecordAlreadyExists
		}
		return fmt.Errorf("postgres error code %s: %w", pqErr.Code, err)
	}
	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, code string) (*domain.PermanentLink, error) {
	query := `SELECT id, code, original_url, created_at, user_id FROM links WHERE code = $1`

	var link domain.PermanentLink
	err := r.db.QueryRowContext(ctx, query, code).Scan(&link.ID, &link.Code, &link.OriginalURL, &link.CreatedAt, &link.UserID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return nil, fmt.Errorf("postgres error getting link (code %s): %w", pgErr.Code, err)
		}

		return nil, fmt.Errorf("unexpected error getting link: %w", err)
	}

	return &link, nil
}

func (r *PostgresRepository) Exists(ctx context.Context, code string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM links WHERE code = $1)`

	err := r.db.QueryRowContext(ctx, query, code).Scan(&exists)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return false, fmt.Errorf("postgres error checking existence (code %s): %w", pgErr.Code, err)
		}

		return false, fmt.Errorf("error checking link existence: %w", err)
	}

	return exists, nil
}
