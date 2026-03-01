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
	Delete(ctx context.Context, code string, userId string) error
}

var ErrRecordNotFound = errors.New("record not found")
var ErrRecordAlreadyExists = errors.New("record already exists")
var ErrLimitExceeded = errors.New("user already exceeded link limit")
var ErrNoLinkDeleted = errors.New("query did not delete any links")
var ErrCouldNotUncache = errors.New("record was not deleted from redis")
