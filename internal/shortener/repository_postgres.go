package shortener

import (
	"context"
	"database/sql"
	"errors"

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
	query := `INSERT INTO links (code, original_url, created_at) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, link.Code, link.OriginalURL, link.CreatedAt)
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

	query := `SELECT code, original_url, created_at FROM links WHERE code = $1`

	row := r.db.QueryRowContext(ctx, query, code)

	var link domain.Link

	if err := row.Scan(&link.Code, &link.OriginalURL, &link.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &link, nil
}
