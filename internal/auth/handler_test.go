package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fernandesenzo/shortener/internal/auth"
	"github.com/fernandesenzo/shortener/internal/jwt"
	"github.com/fernandesenzo/shortener/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestHandlerLogin(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	password := "secret123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost) // MinCost para o teste ser rápido
	if err != nil {
		t.Fatalf("failed to generate hash: %v", err)
	}

	nickname := "enzo"
	_, err = db.ExecContext(context.Background(), `INSERT INTO users (nickname, password_hash) VALUES ($1, $2)`, nickname, string(hash))
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	repo := auth.NewPostgresRepository(db)
	jwtManager := jwt.NewManager("test-secret", time.Hour)
	service := auth.NewService(repo, jwtManager)
	handler := auth.NewHandler(service)

	tests := []struct {
		name           string
		reqBody        string
		contentType    string
		expectedStatus int
		expectedInBody string
	}{
		{
			name:           "Success",
			reqBody:        `{"nickname": "enzo", "password": "secret123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedInBody: `"token":`,
		},
		{
			name:           "Invalid Password",
			reqBody:        `{"nickname": "enzo", "password": "wrongpassword"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnauthorized,
			expectedInBody: "invalid credentials",
		},
		{
			name:           "User Not Found",
			reqBody:        `{"nickname": "ghost", "password": "secret123"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnauthorized,
			expectedInBody: "invalid credentials",
		},
		{
			name:           "Invalid Content-Type",
			reqBody:        `{"nickname": "enzo", "password": "secret123"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedInBody: "unsupported content type",
		},
		{
			name:           "Invalid JSON Format",
			reqBody:        `{"nickname": "enzo", "password":`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "invalid request",
		},
		{
			name:           "Unknown Fields",
			reqBody:        `{"nickname": "enzo", "password": "secret123", "admin": true}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.reqBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedInBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedInBody, w.Body.String())
			}
		})
	}
}
