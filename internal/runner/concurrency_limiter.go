package runner

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

// ErrConcurrencyLimitExceeded is returned when the concurrency cap is reached.
var ErrConcurrencyLimitExceeded = errors.New("concurrency limit exceeded")

// ConcurrencyLimiter tracks the number of in-flight calls and rejects new ones
// once the configured maximum is reached.
type ConcurrencyLimiter struct {
	max     int64
	inflight atomic.Int64
}

// NewConcurrencyLimiter creates a ConcurrencyLimiter with the given maximum.
// If max <= 0 it defaults to 1.
func NewConcurrencyLimiter(max int) *ConcurrencyLimiter {
	if max <= 0 {
		max = 1
	}
	return &ConcurrencyLimiter{max: int64(max)}
}

// Acquire attempts to reserve a slot. It returns an error if the limit has
// been reached or the context is already done.
func (c *ConcurrencyLimiter) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	current := c.inflight.Add(1)
	if current > c.max {
		c.inflight.Add(-1)
		return fmt.Errorf("%w: max=%d", ErrConcurrencyLimitExceeded, c.max)
	}
	return nil
}

// Release decrements the in-flight counter. It must be called after every
// successful Acquire.
func (c *ConcurrencyLimiter) Release() {
	c.inflight.Add(-1)
}

// Inflight returns the current number of in-flight calls.
func (c *ConcurrencyLimiter) Inflight() int64 {
	return c.inflight.Load()
}

// Max returns the configured maximum concurrency.
func (c *ConcurrencyLimiter) Max() int64 {
	return c.max
}

// Wrap executes fn inside the limiter, acquiring before the call and releasing
// after it completes.
func (c *ConcurrencyLimiter) Wrap(ctx context.Context, fn func(context.Context) error) error {
	if err := c.Acquire(ctx); err != nil {
		return err
	}
	defer c.Release()
	return fn(ctx)
}
