package runner

import (
	"context"
	"sync"
	"time"
)

// CooldownPolicy enforces a minimum wait period after a failed call before
// allowing the next call to proceed.
type CooldownPolicy struct {
	mu       sync.Mutex
	duration time.Duration
	lastFail time.Time
}

// NewCooldownPolicy creates a CooldownPolicy with the given cooldown duration.
// If duration is zero or negative, no cooldown is enforced.
func NewCooldownPolicy(d time.Duration) *CooldownPolicy {
	return &CooldownPolicy{duration: d}
}

// RecordFailure marks the current time as the last failure.
func (c *CooldownPolicy) RecordFailure() {
	if c.duration <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastFail = time.Now()
}

// Wait blocks until the cooldown period has elapsed or ctx is cancelled.
// Returns ctx.Err() if the context is cancelled while waiting.
func (c *CooldownPolicy) Wait(ctx context.Context) error {
	if c.duration <= 0 {
		return nil
	}
	c.mu.Lock()
	remaining := time.Until(c.lastFail.Add(c.duration))
	c.mu.Unlock()

	if remaining <= 0 {
		return nil
	}

	select {
	case <-time.After(remaining):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wrap executes fn, enforcing the cooldown wait before the call and recording
// a failure if fn returns an error.
func (c *CooldownPolicy) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := c.Wait(ctx); err != nil {
		return err
	}
	err := fn(ctx)
	if err != nil {
		c.RecordFailure()
	}
	return err
}
