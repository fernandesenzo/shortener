package shortener

import (
	"errors"
	"fmt"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Shorten(originalURL string) (*domain.Link, error) {
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrLinkCreationFailed, err)
	}
	link := &domain.Link{
		Code:        code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	if len(link.OriginalURL) > 100 {
		return nil, domain.ErrURLTooLong
	}
	if err := s.repo.Save(link); err != nil {

		return nil, fmt.Errorf("%w: %v", domain.ErrLinkCreationFailed, err)
	}
	return link, nil
}

func (s *Service) Get(code string) (*domain.Link, error) {
	link, err := s.repo.Get(code)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, domain.ErrLinkNotFound
		}
		return nil, fmt.Errorf("could not get link: %w", err)
	}

	return link, nil
}
