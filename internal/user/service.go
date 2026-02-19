package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/password"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, nickname string, pass string) (*domain.User, error) {
	if err := password.Validate(pass); err != nil {
		slog.InfoContext(ctx, "user sent a password that doesnt match requirements.",
			"nickname", nickname)
		return nil, err
	}
	hashed, err := password.Hash(pass)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password during user creation",
			"nickname", nickname,
			"err:", err,
		)
		return nil, err
	}
	user := &domain.User{
		Nickname:     nickname,
		PasswordHash: hashed,
	}
	if err := s.repo.Save(ctx, user); err != nil {
		if errors.Is(err, ErrRecordAlreadyExists) {
			slog.InfoContext(ctx, "attempt to create an user with already existing nickname",
				"nickname", nickname)
			return nil, domain.ErrNicknameAlreadyUsed
		}
		slog.ErrorContext(ctx, "unknown db error when saving user", "error", err)
		return nil, err
	}
	return user, nil
}
