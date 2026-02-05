package shortener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener/utils"
)

type Service struct {
	repo LinkRepository
}

func NewService(repo LinkRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Shorten(ctx context.Context, originalURL string) (*domain.Link, error) {
	if len(originalURL) > 100 {
		return nil, domain.ErrURLTooLong
	}

	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, domain.ErrInvalidURL
	}

	for i := 0; i < 5; i++ {
		code, err := utils.GenerateCode(6)
		if err != nil {
			continue
		}

		link := &domain.Link{
			Code:        code,
			OriginalURL: originalURL,
		}

		if err := s.repo.Save(ctx, link); err != nil {
			if errors.Is(err, ErrRecordAlreadyExists) {
				continue
			}
			return nil, fmt.Errorf("%w: %v", domain.ErrLinkCreationFailed, err)
		}

		return link, nil
	}
	return nil, fmt.Errorf("%w: failed to generate unique code after 5 attempts", domain.ErrLinkCreationFailed)

}

func (s *Service) Get(ctx context.Context, code string) (*domain.Link, error) {
	link, err := s.repo.Get(ctx, code)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, domain.ErrLinkNotFound
		}
		slog.ErrorContext(ctx, "failed to get link", "error", err, "code", code)
		return nil, fmt.Errorf("unexpected database error: %w", err)
	}

	return link, nil
}
