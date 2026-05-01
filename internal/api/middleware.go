package api

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxRateBuckets = 500

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	maxTokens  int
	refillRate time.Duration
}

type bucket struct { tokens int; lastRefill time.Time }

func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*bucket, maxRateBuckets), maxTokens: maxTokens, refillRate: refillRate}
}

func (rl *RateLimiter) prune() {
	if len(rl.buckets) <= maxRateBuckets { return }
	var oldest string
	var oldestTime time.Time
	first := true
	for ip, b := range rl.buckets {
		if first || b.lastRefill.Before(oldestTime) {
			oldest = ip
			oldestTime = b.lastRefill
			first = false
		}
	}
	if !first { delete(rl.buckets, oldest) }
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.prune()
	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &bucket{tokens: rl.maxTokens, lastRefill: now}
		return true
	}
	elapsed := now.Sub(b.lastRefill)
	add := int(elapsed / rl.refillRate)
	if add > 0 {
		b.tokens += add
		if b.tokens > rl.maxTokens { b.tokens = rl.maxTokens }
		b.lastRefill = now
	}
	if b.tokens > 0 { b.tokens--; return true }
	return false
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" { ip = strings.Split(fwd, ",")[0] }
		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate_limit_exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "path", r.URL.Path, "err", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal_server_error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
		slog.Debug("response", "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
