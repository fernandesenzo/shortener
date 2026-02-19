package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernandesenzo/shortener/internal/user"
)

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		repoMock       *MockRepository
		expectedStatus int
		expectedID     string
		expectedNick   string
		expectedError  string
	}{
		{
			name:           "must error with invalid json",
			reqBody:        `{invalid json]}`,
			repoMock:       &MockRepository{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "successfully create user",
			reqBody:        `{"nickname":"enzo","password":"strongpassword"}`,
			repoMock:       &MockRepository{},
			expectedStatus: http.StatusCreated,
			expectedID:     "uuid-123",
			expectedNick:   "enzo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := user.NewService(tt.repoMock)
			handler := user.NewHandler(srv)

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected code %d, got %d", tt.expectedStatus, rr.Code)
			}

			var resp struct {
				ID        string `json:"id"`
				Nickname  string `json:"nickname"`
				CreatedAt string `json:"createdAt"`
				Error     string `json:"error"`
			}

			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("error reading response json %v", err)
			}
			if tt.expectedError != "" {
				if resp.Error != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, resp.Error)
				}
			} else {
				if resp.ID != tt.expectedID {
					t.Errorf("expected id %q, got %q", tt.expectedID, resp.ID)
				}
				if resp.Nickname != tt.expectedNick {
					t.Errorf("expected nickname %q, got %q", tt.expectedNick, resp.Nickname)
				}
				if resp.CreatedAt == "" {
					t.Error("expected non empty created-at")
				}
			}
		})
	}
}
