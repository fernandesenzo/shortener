package user

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

func (r *PostgresRepository) Save(ctx context.Context, usr *domain.User) error {
	query := `INSERT INTO users (nickname,password_hash) VALUES ($1,$2) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, usr.Nickname, usr.PasswordHash).Scan(&usr.ID, &usr.CreatedAt)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrRecordAlreadyExists
			}
			return fmt.Errorf("postgres error code %s: %w", pgErr.Code, err)
		}
		return err
	}

	return nil
}
