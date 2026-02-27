package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRateLimitMiddleware(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limit := 1
	window := time.Minute

	t.Run("X-Forwarded-For Priority", func(t *testing.T) {
		mr.FlushAll()
		mw := RateLimitMiddleware(nextHandler, client, limit, window)

		req1 := httptest.NewRequest("POST", "/api/links", nil)
		req1.Header.Set("X-Forwarded-For", "1.2.3.4")
		rr1 := httptest.NewRecorder()
		mw.ServeHTTP(rr1, req1)

		if rr1.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr1.Code)
		}

		req2 := httptest.NewRequest("POST", "/api/links", nil)
		req2.Header.Set("X-Forwarded-For", "1.2.3.4")
		rr2 := httptest.NewRecorder()
		mw.ServeHTTP(rr2, req2)

		if rr2.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", rr2.Code)
		}
	})

	t.Run("RemoteAddr Fallback", func(t *testing.T) {
		mr.FlushAll()
		mw := RateLimitMiddleware(nextHandler, client, limit, window)

		req1 := httptest.NewRequest("POST", "/api/links", nil)
		req1.RemoteAddr = "127.0.0.1:1234"
		rr1 := httptest.NewRecorder()
		mw.ServeHTTP(rr1, req1)

		if rr1.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr1.Code)
		}

		req2 := httptest.NewRequest("POST", "/api/links", nil)
		req2.RemoteAddr = "127.0.0.1:1234"
		rr2 := httptest.NewRecorder()
		mw.ServeHTTP(rr2, req2)

		if rr2.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", rr2.Code)
		}
	})

	t.Run("Skip Rate Limit on GET", func(t *testing.T) {
		mr.FlushAll()
		mw := RateLimitMiddleware(nextHandler, client, 0, window)

		req := httptest.NewRequest("GET", "/abc", nil)
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("Retry-After Header", func(t *testing.T) {
		mr.FlushAll()
		mw := RateLimitMiddleware(nextHandler, client, 1, window)

		req1 := httptest.NewRequest("POST", "/api/links", nil)
		req1.RemoteAddr = "8.8.8.8:1234"
		mw.ServeHTTP(httptest.NewRecorder(), req1)

		req2 := httptest.NewRequest("POST", "/api/links", nil)
		req2.RemoteAddr = "8.8.8.8:1234"
		rr2 := httptest.NewRecorder()
		mw.ServeHTTP(rr2, req2)

		retryAfter := rr2.Header().Get("Retry-After")
		if retryAfter == "" {
			t.Error("expected Retry-After header, got empty string")
		}
	})
}
