package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type requestStep struct {
	ip           string
	expectedCode int
}

func repeatSteps(count int, ip string, expectedCode int) []requestStep {
	steps := make([]requestStep, count)
	for i := 0; i < count; i++ {
		steps[i] = requestStep{ip: ip, expectedCode: expectedCode}
	}
	return steps
}

func TestRateLimiterMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		globalRPS   int
		globalBurst int
		ipRPS       int
		ipBurst     int
		steps       []requestStep
	}{
		{
			name:        "should allow single request within limits",
			globalRPS:   100,
			globalBurst: 100,
			ipRPS:       1,
			ipBurst:     1,
			steps: []requestStep{
				{ip: "192.168.0.1", expectedCode: http.StatusOK},
			},
		},
		{
			name:        "should block IP after exceeding IP burst",
			globalRPS:   100,
			globalBurst: 100,
			ipRPS:       1,
			ipBurst:     3,
			steps: append(
				repeatSteps(3, "192.168.0.1", http.StatusOK),
				requestStep{ip: "192.168.0.1", expectedCode: http.StatusTooManyRequests},
			),
		},
		{
			name:        "should block globally when server capacity is full",
			globalRPS:   1,
			globalBurst: 5,
			ipRPS:       100,
			ipBurst:     100,
			steps: append(
				repeatSteps(5, "10.0.0.1", http.StatusOK),
				requestStep{ip: "10.0.0.2", expectedCode: http.StatusServiceUnavailable},
			),
		},
		{
			name:        "different IPs have different rates",
			globalRPS:   100,
			globalBurst: 100,
			ipRPS:       1,
			ipBurst:     1,
			steps: []requestStep{
				{ip: "192.168.0.1", expectedCode: http.StatusOK},
				{ip: "192.168.0.1", expectedCode: http.StatusTooManyRequests},
				{ip: "192.168.0.2", expectedCode: http.StatusOK},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.globalRPS, tt.globalBurst, tt.ipRPS, tt.ipBurst, 0)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middlewareToTest := rl.Middleware(nextHandler)

			for i, step := range tt.steps {
				req := httptest.NewRequest("GET", "http://localhost:8080/", nil)
				req.RemoteAddr = step.ip + ":1234"

				rr := httptest.NewRecorder()
				middlewareToTest.ServeHTTP(rr, req)

				if rr.Code != step.expectedCode {
					t.Errorf("step %d (%s): got status %d, expected %d",
						i+1, step.ip, rr.Code, step.expectedCode)
				}
			}
		})
	}
}
