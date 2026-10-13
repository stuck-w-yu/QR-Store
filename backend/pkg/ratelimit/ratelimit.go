// Package ratelimit provides a tiny in-memory, per-key token bucket limiter
// suitable for protecting sensitive endpoints (login, webhook, public order
// endpoints) from brute-force and abuse without external dependencies.
//
// NOTE: state is per-process. When running multiple replicas, put a shared
// limiter (e.g. Redis) in front; this is a defense-in-depth baseline.
package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"qr-store/backend/pkg/response"
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// Limiter is a simple token-bucket limiter keyed by an arbitrary string.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64 // max burst
	lastGC   time.Time
}

// New creates a limiter allowing `burst` requests, refilling at `perSecond`.
func New(perSecond float64, burst int) *Limiter {
	if perSecond <= 0 {
		perSecond = 1
	}
	if burst <= 0 {
		burst = 1
	}
	return &Limiter{
		buckets:  make(map[string]*bucket),
		rate:     perSecond,
		capacity: float64(burst),
		lastGC:   time.Now(),
	}
}

func (l *Limiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, lastSeen: now}
		l.buckets[key] = b
	} else {
		elapsed := now.Sub(b.lastSeen).Seconds()
		b.tokens = min(l.capacity, b.tokens+elapsed*l.rate)
		b.lastSeen = now
	}

	// Opportunistic GC of stale buckets.
	if now.Sub(l.lastGC) > 10*time.Minute {
		for k, v := range l.buckets {
			if now.Sub(v.lastSeen) > 10*time.Minute {
				delete(l.buckets, k)
			}
		}
		l.lastGC = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware returns a Gin middleware enforcing the limit per client IP.
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please slow down")
			c.Abort()
			return
		}
		c.Next()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
