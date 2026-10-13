package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLimiterAllowsBurstThenBlocks(t *testing.T) {
	l := New(0.0001, 3) // effectively no refill during test
	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("4th request should be blocked")
	}
}

func TestLimiterIsolatedPerKey(t *testing.T) {
	l := New(0.0001, 1)
	if !l.allow("a") {
		t.Fatal("first key should be allowed")
	}
	if !l.allow("b") {
		t.Fatal("second key should be allowed independently")
	}
	if l.allow("a") {
		t.Fatal("first key should now be blocked")
	}
}

func TestLimiterRefillsOverTime(t *testing.T) {
	// 1000 tokens/second so refill is effectively immediate.
	l := New(1000, 1)
	if !l.allow("k") {
		t.Fatal("first request should be allowed")
	}
	if l.allow("k") {
		t.Fatal("second immediate request should be blocked (burst exhausted)")
	}
	time.Sleep(5 * time.Millisecond) // 5ms * 1000/s = 5 tokens refilled, capped at 1
	if !l.allow("k") {
		t.Fatal("request after refill should be allowed")
	}
}

func TestNewClampsInvalidParams(t *testing.T) {
	// Non-positive rate/burst must be clamped to sane minimums (no panic, no divide-by-zero).
	l := New(0, 0)
	if !l.allow("x") {
		t.Fatal("clamped limiter must still allow the first request")
	}
	if l.allow("x") {
		t.Fatal("second request should be blocked with capacity clamped to 1")
	}
}

func TestMiddlewareBlocksWith429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := New(0.0001, 1)
	r := gin.New()
	r.Use(l.Middleware())
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	// First request passes.
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/x", nil))
	if w1.Code != 200 {
		t.Fatalf("first request: got %d, want 200", w1.Code)
	}

	// Second from the same IP is rate limited.
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("GET", "/x", nil))
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: got %d, want 429", w2.Code)
	}
}
