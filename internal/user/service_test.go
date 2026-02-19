package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	user2 "github.com/fernandesenzo/shortener/internal/user"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name         string
		nickname     string
		password     string
		repoError    bool
		preExistUser bool
		wantErr      error
	}{
		{
			name:         "Success",
			nickname:     "enzo",
			password:     "password123",
			repoError:    false,
			preExistUser: false,
			wantErr:      nil,
		},
		{
			name:         "Duplicate Nickname",
			nickname:     "enzo",
			password:     "password123",
			preExistUser: true,
			wantErr:      domain.ErrNicknameAlreadyUsed,
		},
		{
			name:      "Unexpected Repository Error",
			nickname:  "database_fail",
			password:  "password123",
			repoError: true,
			wantErr:   ErrMockedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRepository{}

			if tt.repoError {
				mock.shouldError = true
			}

			if tt.preExistUser {
				mock.users = append(mock.users, &domain.User{Nickname: tt.nickname})
			}

			svc := user2.NewService(mock)

			_, err := svc.Create(context.Background(), tt.nickname, tt.password)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(mock.users) == 0 {
				t.Error("expected user to be saved in mock repository")
			}
		})
	}
}
