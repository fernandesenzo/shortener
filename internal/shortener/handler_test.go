package shortener_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/shortener/internal/domain"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

func TestShortenHandlerSuccess(t *testing.T) {
	repo := &MockRepository{}
	service := shortener.NewService(repo)
	handler := shortener.NewHandler(service)

	reqBody := `{"url":"https://google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Shorten(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Errorf("incorrect status: got %d, wanted %d", recorder.Code, http.StatusCreated)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("incorrect content-type: got %s, wanted application/json", contentType)
	}

	var response domain.Link
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	expectedURL := "https://google.com"
	if response.OriginalURL != expectedURL {
		t.Errorf("handler changed original URL, expected %s and got %s", expectedURL, response.Code)
	}
}

func TestShortenEmptyBody(t *testing.T) {
	repo := &MockRepository{}
	service := shortener.NewService(repo)
	handler := shortener.NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.Shorten(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("an empty body was sent, expected status 400 and got %v", recorder.Code)
	}
}

func TestShortenInvalidURL(t *testing.T) {
	repo := &MockRepository{}
	service := shortener.NewService(repo)
	handler := shortener.NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"invalid url"}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.Shorten(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("invalid url was sent, expected status 400 and got %v", recorder.Code)
	}
}
