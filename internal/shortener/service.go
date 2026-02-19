package shortener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type Service struct {
	repo LinkRepository
}

func NewService(repo LinkRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Shorten(ctx context.Context, originalURL string, userID string) (domain.Link, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	link, err := s.saveLink(ctx, userID, originalURL)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Service) Get(ctx context.Context, code string) (domain.Link, error) {
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

func (s *Service) saveLink(ctx context.Context, userID string, originalURL string) (domain.Link, error) {
	for i := 0; i < 10; i++ {
		code, err := GenerateCode(6)
		if err != nil {
			return nil, fmt.Errorf("internal error generating code: %w", err)
		}
		if userID == "" {
			link := &domain.TemporaryLink{
				OriginalURL: originalURL,
				Code:        code,
			}
			if err := s.repo.TempSave(ctx, link, 24*time.Hour); err != nil {
				if errors.Is(err, ErrRecordAlreadyExists) {
					continue
				}
				return nil, fmt.Errorf("failed to save link: %w", err)
			}
			return link, nil
		}
		link := &domain.PermanentLink{
			Code:        code,
			OriginalURL: originalURL,
			UserID:      userID,
			CreatedAt:   time.Now(),
		}
		if err := s.repo.PermSave(ctx, link); err != nil {
			if errors.Is(err, ErrRecordAlreadyExists) {
				continue
			}
			return nil, fmt.Errorf("failed to save link: %w", err)
		}
		return link, nil
	}
	return nil, domain.ErrLinkCreationFailed
}

func validateURL(originalURL string) error {
	if len(originalURL) > 100 {
		return domain.ErrURLTooLong
	}
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return domain.ErrInvalidURL
	}
	return nil
}
