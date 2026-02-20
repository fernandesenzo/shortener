package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/jwt"
	"github.com/fernandesenzo/shortener/internal/password"
)

type Service struct {
	repo       Repository
	jwtManager jwt.Manager
}

func NewService(repo Repository, jwtManager jwt.Manager) *Service {
	return &Service{repo: repo, jwtManager: jwtManager}
}

func (s *Service) Authenticate(ctx context.Context, nickname string, pswd string) (string, error) {
	user, err := s.repo.GetByNickname(ctx, nickname)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			password.CompareDummy(pswd)
			return "", domain.ErrNicknameNotFound
		}
		slog.ErrorContext(ctx, "unknown db error when trying to get user by nickname", "nickname", nickname, "error", err)
		return "", err
	}

	if err := password.Compare(user.PasswordHash, pswd); err != nil {
		return "", domain.ErrInvalidPassword
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "unknown error when generating jwt token", "error", err)
		return "", err
	}
	return token, nil
}
