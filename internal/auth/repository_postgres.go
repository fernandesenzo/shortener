package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}
func (r PostgresRepository) GetByNickname(ctx context.Context, nickname string) (*domain.User, error) {
	query := `SELECT id, nickname, password_hash, created_at FROM users WHERE nickname = $1`
	var user domain.User
	row := r.db.QueryRowContext(ctx, query, nickname)
	err := row.Scan(
		&user.ID,
		&user.Nickname,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}
