package shortener

import (
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type Repository interface {
	Save(link *domain.Link) error
	Get(code string) (*domain.Link, error)
}

var ErrRecordNotFound = errors.New("record not found")
