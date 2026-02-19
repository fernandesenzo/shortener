package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client      *redis.Client
	ipLimit     int
	globalLimit int
}

func NewRateLimiter(client *redis.Client, ipLimit, globalLimit int) *RateLimiter {
	return &RateLimiter{
		client:      client,
		ipLimit:     ipLimit,
		globalLimit: globalLimit,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		now := time.Now().Unix()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		globalKey := fmt.Sprintf("rl:g:%d", now)
		ipKey := fmt.Sprintf("rl:i:%s:%d", ip, now)

		pipe := rl.client.Pipeline()
		gCmd := pipe.Incr(ctx, globalKey)
		iCmd := pipe.Incr(ctx, ipKey)

		_, err := pipe.Exec(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "ratelimiter: redis pipeline failed", "error", err, "ip", ip)
			next.ServeHTTP(w, r)
			return
		}

		if v, _ := gCmd.Result(); v == 1 {
			rl.client.Expire(ctx, globalKey, 2*time.Second)
		}
		if v, _ := iCmd.Result(); v == 1 {
			rl.client.Expire(ctx, ipKey, 2*time.Second)
		}

		if gCmd.Val() > int64(rl.globalLimit) {
			slog.WarnContext(ctx, "ratelimiter: global limit exceeded", "limit", rl.globalLimit, "current", gCmd.Val())
			http.Error(w, "server capacity exceeded", http.StatusServiceUnavailable)
			return
		}

		if iCmd.Val() > int64(rl.ipLimit) {
			slog.DebugContext(ctx, "ratelimiter: ip limit exceeded", "ip", ip, "limit", rl.ipLimit)
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
