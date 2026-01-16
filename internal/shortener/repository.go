package shortener

import "github.com/fernandesenzo/shortener/internal/domain"

type Repository interface {
	Save(link *domain.Link) error
	Get(code string) (*domain.Link, error)
}
