package user

import (
	"context"
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type Repository interface {
	Save(ctx context.Context, user *domain.User) error
}

var ErrRecordNotFound = errors.New("record not found")
var ErrRecordAlreadyExists = errors.New("record already exists")
