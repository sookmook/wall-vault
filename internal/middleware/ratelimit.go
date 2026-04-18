package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a minimal per-IP token-bucket limiter kept small on purpose:
// wall-vault runs on a LAN with a tiny number of clients, so introducing
// golang.org/x/time/rate for the same behavior would be an unnecessary external
// dependency. A background janitor trims idle visitor state so the map stays
// bounded in long-running processes.
type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitorState
	ratePerSec float64
	burst      int
}

type visitorState struct {
	tokens    float64
	lastCheck time.Time
}

// NewRateLimiter creates a limiter that replenishes ratePerSec tokens per IP
// per second, with a burst cap of `burst`. A goroutine drops visitor entries
// that have been idle for >10 minutes to keep memory bounded.
func NewRateLimiter(ratePerSec float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitorState),
		ratePerSec: ratePerSec,
		burst:      burst,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if v.lastCheck.Before(cutoff) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an http middleware that rejects requests with HTTP 429
// when the caller's IP exceeds the limiter's rate. On reject we set a
// Retry-After header so compliant clients back off for a second.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(realIP(r)) {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &visitorState{tokens: float64(rl.burst) - 1, lastCheck: now}
		return true
	}
	elapsed := now.Sub(v.lastCheck).Seconds()
	v.tokens += elapsed * rl.ratePerSec
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.lastCheck = now
	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

func realIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || ip == "" {
		return r.RemoteAddr
	}
	return ip
}
