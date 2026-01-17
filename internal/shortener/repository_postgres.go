package shortener

import (
	"context"
	"database/sql"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db}
}
func (r *PostgresRepository) Save(link *domain.Link) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO links (code, original_url, created_at) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, link.Code, link.OriginalURL, link.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) Get(code string) (*domain.Link, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT code, original_url, created_at FROM links WHERE code = $1`

	row := r.db.QueryRowContext(ctx, query, code)

	var link domain.Link

	if err := row.Scan(&link.Code, &link.OriginalURL, &link.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &link, nil
}
