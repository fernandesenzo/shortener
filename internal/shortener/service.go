package shortener

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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

func (s *Service) Shorten(ctx context.Context, originalURL string) (*domain.Link, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, domain.ErrInvalidURL
	}
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrLinkCreationFailed, err)
	}
	if len(originalURL) > 100 {
		return nil, domain.ErrURLTooLong
	}

	link := &domain.Link{
		Code:        code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Save(ctx, link); err != nil {

		return nil, fmt.Errorf("%w: %v", domain.ErrLinkCreationFailed, err)
	}
	return link, nil
}

func (s *Service) Get(ctx context.Context, code string) (*domain.Link, error) {
	link, err := s.repo.Get(ctx, code)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, domain.ErrLinkNotFound
		}
		return nil, fmt.Errorf("could not get link: %w", err)
	}

	return link, nil
}
