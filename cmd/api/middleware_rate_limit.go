package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mutex         sync.Mutex
	globalLimiter *rate.Limiter
	clients       map[string]*client
	requestsPerS  rate.Limit
	burst         int
	ttl           time.Duration
}

func NewRateLimiter(globalRequestsPerS int, globalBurst int, requestsPerS int, burst int, ttl int) *RateLimiter {
	rl := &RateLimiter{
		clients:       make(map[string]*client),
		globalLimiter: rate.NewLimiter(rate.Limit(globalRequestsPerS), globalBurst),
		requestsPerS:  rate.Limit(requestsPerS),
		burst:         burst,
		ttl:           time.Duration(ttl) * time.Second,
	}
	if ttl > 0 {
		go rl.cleanup()
	}

	return rl
}

func (rl *RateLimiter) getClient(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.requestsPerS, rl.burst)
		rl.clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	c.lastSeen = time.Now()
	return c.limiter
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(1 * time.Minute)
		rl.mutex.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > rl.ttl {
				delete(rl.clients, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.globalLimiter.Allow() {
			http.Error(w, "service unavailable, try again later", http.StatusServiceUnavailable)
			return
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		limiter := rl.getClient(ip)

		if !limiter.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
