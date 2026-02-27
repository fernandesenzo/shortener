package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fernandesenzo/shortener/internal/auth"
	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/jwt"
)

func TestService_Authenticate(t *testing.T) {
	validPasswordHash := "$2a$12$AhH7jtSk/Y5hkWtkNW3ygePi./4IzRr6F3ocXwLiev5BawuUef9wq"
	realJwtManager := jwt.NewManager("secret-key-test", 1*time.Hour)

	tests := []struct {
		name          string
		nickname      string
		password      string
		mockRepo      *MockRepository
		wantToken     bool
		expectedError error
	}{
		{
			name:     "success",
			nickname: "enzo",
			password: "valid_password",
			mockRepo: &MockRepository{
				shouldError: false,
				users: []*domain.User{
					{ID: "123", Nickname: "enzo", PasswordHash: validPasswordHash},
				},
			},
			wantToken:     true,
			expectedError: nil,
		},
		{
			name:     "failure - invalid user",
			nickname: "invalid_user",
			password: "invalid_password",
			mockRepo: &MockRepository{
				shouldError: false,
				users:       []*domain.User{},
			},
			wantToken:     false,
			expectedError: domain.ErrNicknameNotFound,
		},
		{
			name:     "failure - invalid password",
			nickname: "enzo",
			password: "invalid_password",
			mockRepo: &MockRepository{
				shouldError: false,
				users: []*domain.User{
					{ID: "123", Nickname: "enzo", PasswordHash: validPasswordHash},
				},
			},
			wantToken:     false,
			expectedError: domain.ErrInvalidPassword,
		},
		{
			name:     "failure - internal error",
			nickname: "enzo",
			password: "valid_password",
			mockRepo: &MockRepository{
				shouldError: true,
				users:       []*domain.User{{ID: "123", Nickname: "enzo", PasswordHash: validPasswordHash}},
			},
			wantToken:     false,
			expectedError: ErrMockedRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := auth.NewService(tt.mockRepo, realJwtManager)

			token, err := svc.Authenticate(context.Background(), tt.nickname, tt.password)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("error = %q, expectedError %q", err, tt.expectedError)
			}

			hasToken := token != ""
			if hasToken != tt.wantToken {
				t.Errorf("got token (not empty) = %v, wantToken %v", hasToken, tt.wantToken)
			}
		})
	}
}
