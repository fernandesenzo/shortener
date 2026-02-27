package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(next http.Handler, client *redis.Client, ipLimit int, window time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		var ip string
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ips := strings.Split(forwarded, ",")
			ip = strings.TrimSpace(ips[0])
		} else {
			var err error
			ip, _, err = net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
		}

		now := time.Now().UTC()
		windowStart := now.Truncate(window)
		key := fmt.Sprintf("rl:w:ip:%s:%d", ip, windowStart.Unix())

		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			slog.ErrorContext(ctx, "ratelimiter: redis failed", "error", err, "ip", ip)
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			client.Expire(ctx, key, window)
		}

		if count > int64(ipLimit) {
			remaining := int(windowStart.Add(window).Sub(now).Seconds())
			w.Header().Set("Retry-After", fmt.Sprintf("%d", remaining))
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
