package auth_test

import (
	"context"
	"errors"

	"github.com/fernandesenzo/shortener/internal/auth"
	"github.com/fernandesenzo/shortener/internal/domain"
)

var ErrMockedRepo = errors.New("forced mockdb error")

type MockRepository struct {
	users       []*domain.User
	shouldError bool
}

func (m *MockRepository) GetByNickname(ctx context.Context, nickname string) (*domain.User, error) {
	if m.shouldError {
		return nil, ErrMockedRepo
	}

	for _, u := range m.users {
		if u.Nickname == nickname {
			return u, nil
		}
	}
	return nil, auth.ErrRecordNotFound
}
