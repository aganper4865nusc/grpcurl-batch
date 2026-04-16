package runner

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate of gRPC calls using a token bucket approach.
type RateLimiter struct {
	rps      float64
	tokens   float64
	max      float64
	lastTick time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a RateLimiter allowing rps calls per second.
// rps <= 0 disables rate limiting.
func NewRateLimiter(rps float64) *RateLimiter {
	return &RateLimiter{
		rps:      rps,
		tokens:   rps,
		max:      rps,
		lastTick: time.Now(),
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	if rl.rps <= 0 {
		return nil
	}
	for {
		rl.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(rl.lastTick).Seconds()
		rl.tokens += elapsed * rl.rps
		if rl.tokens > rl.max {
			rl.tokens = rl.max
		}
		rl.lastTick = now
		if rl.tokens >= 1 {
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}
		wait := time.Duration(float64(time.Second) / rl.rps)
		rl.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}
