package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRateLimiter(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(client, 2, 5)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("IP limit enforcement", func(t *testing.T) {
		ip := "1.1.1.1"
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":1234"

		for i := 0; i < 2; i++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rr.Code)
			}
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", rr.Code)
		}
	})

	t.Run("Global limit enforcement", func(t *testing.T) {
		mr.FastForward(2 * time.Second)

		for i := 1; i <= 5; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = fmt.Sprintf("1.1.1.%d:1234", i)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rr.Code)
			}
		}

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.1.1.99:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", rr.Code)
		}
	})

	t.Run("Fail-open on Redis error", func(t *testing.T) {
		mr.Close()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "2.2.2.2:1234"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 on redis error, got %d", rr.Code)
		}
	})
}
