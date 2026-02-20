package auth

import (
	"context"
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
)

var ErrRecordNotFound = errors.New("nickname does not exist")

type Repository interface {
	GetByNickname(ctx context.Context, nickname string) (*domain.User, error)
}
