package user_test

import (
	"context"
	"errors"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/user"
)

var ErrMockedError = errors.New("forced db error")

type MockRepository struct {
	users       []*domain.User
	shouldError bool
}

func (m *MockRepository) Save(ctx context.Context, usr *domain.User) error {
	if m.shouldError {
		return ErrMockedError
	}

	for _, u := range m.users {
		if u.Nickname == usr.Nickname {
			return user.ErrRecordAlreadyExists
		}
	}

	if usr.ID == "" {
		usr.ID = "uuid-123"
	}

	m.users = append(m.users, usr)
	return nil
}
