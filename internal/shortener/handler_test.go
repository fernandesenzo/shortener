package shortener_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

func TestHandlerGet(t *testing.T) {
	tests := []struct {
		name           string
		codeParam      string
		setupLink      *domain.Link
		shouldError    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			codeParam:      "abcdef",
			setupLink:      &domain.Link{Code: "abcdef", OriginalURL: "https://google.com"},
			shouldError:    false,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedBody:   ``,
		},
		{
			name:           "Not Found",
			codeParam:      "missing",
			setupLink:      nil,
			shouldError:    false,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "link not found",
		},
		{
			name:           "Internal Error",
			codeParam:      "any",
			setupLink:      nil,
			shouldError:    true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}

			repo.SetShouldError(tt.shouldError)

			if tt.setupLink != nil {
				_ = repo.Save(context.Background(), tt.setupLink)
			}

			service := shortener.NewService(repo)
			handler := shortener.NewHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.codeParam, nil)
			req.SetPathValue("code", tt.codeParam)
			w := httptest.NewRecorder()

			handler.Get(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandlerShorten(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		contentType    string
		expectedStatus int
		expectedInBody string
		shouldError    bool
	}{
		{
			name:           "Success",
			reqBody:        `{"url": "https://google.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectedInBody: `"code":`,
			shouldError:    false,
		},
		{
			name:           "Invalid Content-Type",
			reqBody:        `{"url": "https://google.com"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedInBody: "unsupported content type",
			shouldError:    false,
		},
		{
			name:           "Invalid JSON format",
			reqBody:        `{"url": "google.com"`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "invalid request",
			shouldError:    false,
		},
		{
			name:           "Unknown Fields",
			reqBody:        `{"url": "https://google.com", "foo": "bar"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "invalid request",
			shouldError:    false,
		},
		{
			name:           "Empty URL",
			reqBody:        `{"url": ""}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "invalid url",
			shouldError:    false,
		},
		{
			name:           "Malformed URL",
			reqBody:        `{"url": "h tp://broken"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "invalid url",
			shouldError:    false,
		},
		{
			name:           "URL Too Long",
			reqBody:        `{"url": "` + strings.Repeat("a", 2001) + `"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedInBody: "url too long",
			shouldError:    false,
		},
		{
			name:           "Link Creation Failed",
			reqBody:        `{"url": "https://google.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "internal server error",
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			if tt.shouldError {
				repo.SetShouldError(true)
			}

			service := shortener.NewService(repo)
			handler := shortener.NewHandler(service)

			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.reqBody))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.Shorten(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedInBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedInBody, w.Body.String())
			}
		})
	}
}
