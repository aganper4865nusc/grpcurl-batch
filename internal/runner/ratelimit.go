package runner

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate of gRPC calls using a token bucket approach.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
}

// NewRateLimiter creates a RateLimiter that allows up to rps calls per second.
// If rps is zero or negative, no rate limiting is applied.
func NewRateLimiter(rps float64) *RateLimiter {
	if rps <= 0 {
		return &RateLimiter{rate: 0}
	}
	return &RateLimiter{
		tokens:   rps,
		max:      rps,
		rate:     rps,
		lastTick: time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	if r.rate <= 0 {
		return nil
	}

	for {
		r.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(r.lastTick).Seconds()
		r.tokens += elapsed * r.rate
		if r.tokens > r.max {
			r.tokens = r.max
		}
		r.lastTick = now

		if r.tokens >= 1.0 {
			r.tokens -= 1.0
			r.mu.Unlock()
			return nil
		}

		// Calculate how long until next token is available.
		waitDuration := time.Duration((1.0-r.tokens)/r.rate*1000) * time.Millisecond
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}
}
