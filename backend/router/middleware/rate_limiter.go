package middleware

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Shiva936/code-review-agent/backend/config"
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type memoryRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

func NewMemoryRateLimiter() *memoryRateLimiter {
	return &memoryRateLimiter{
		buckets: map[string]*tokenBucket{},
	}
}

func (rl *memoryRateLimiter) allow(key string, rule config.RateLimitRule) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{
			tokens:     float64(rule.BucketSize),
			lastRefill: now,
		}
		rl.buckets[key] = b
	}

	// Refill
	if rule.RefillDuration > 0 && rule.RefillSize > 0 {
		elapsed := now.Sub(b.lastRefill)
		if elapsed > 0 {
			refills := float64(elapsed) / float64(rule.RefillDuration)
			add := refills * float64(rule.RefillSize)
			if add > 0 {
				b.tokens = math.Min(float64(rule.BucketSize), b.tokens+add)
				b.lastRefill = now
			}
		}
	}

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

func RateLimit(cfg *config.Config, route string, rl *memoryRateLimiter, next http.Handler) http.Handler {
	if cfg == nil || !cfg.RateLimit.Enabled {
		return next
	}

	if cfg.RateLimit.Storage != "" && cfg.RateLimit.Storage != "memory" {
		log.Printf("warning: rate_limit.storage=%q not supported, using memory", cfg.RateLimit.Storage)
	}

	rule := cfg.RateLimit.DefaultRule
	if cfg.RateLimit.Routes != nil {
		if r, ok := cfg.RateLimit.Routes[route]; ok {
			rule = r
		}
	}

	if !rule.Enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := identifyRequest(r, rule.IdentifyBy)
		// Route-scoped key so per-route limits don't collide.
		key = route + ":" + key

		if !rl.allow(key, rule) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func identifyRequest(r *http.Request, identifyBy string) string {
	switch strings.ToLower(strings.TrimSpace(identifyBy)) {
	case "api_key":
		// Minimal support: treat Authorization header as identifier.
		// (No new auth mechanisms introduced.)
		return r.Header.Get("Authorization")
	default:
		// Default: IP address.
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && host != "" {
			return host
		}
		return r.RemoteAddr
	}
}
