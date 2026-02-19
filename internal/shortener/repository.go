package shortener

import (
	"context"
	"errors"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type LinkRepository interface {
	TempSave(ctx context.Context, link *domain.TemporaryLink, ttl time.Duration) error
	PermSave(ctx context.Context, link *domain.PermanentLink) error
	Get(ctx context.Context, code string) (domain.Link, error)
}

var ErrRecordNotFound = errors.New("record not found")
var ErrRecordAlreadyExists = errors.New("record already exists")
