package shortener

import (
	"context"
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type LinkRepository interface {
	Save(ctx context.Context, link *domain.Link) error
	Get(ctx context.Context, code string) (*domain.Link, error)
}

var ErrRecordNotFound = errors.New("record not found")
var ErrRecordAlreadyExists = errors.New("record already exists")
