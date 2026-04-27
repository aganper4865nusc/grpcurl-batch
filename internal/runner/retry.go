package runner

import (
	"context"
	"fmt"
	"time"
)

// RetryPolicy defines the retry behaviour for a single call.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

// RetryableFunc is a function that can be retried.
type RetryableFunc func(ctx context.Context, attempt int) error

// Execute runs fn up to policy.MaxAttempts times, sleeping policy.Delay
// between failures. It returns nil on the first success, or the last error
// if all attempts are exhausted.
func (p RetryPolicy) Execute(ctx context.Context, fn RetryableFunc) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled before attempt %d: %w", attempt, err)
		}

		lastErr = fn(ctx, attempt)
		if lastErr == nil {
			return nil
		}

		if attempt < p.MaxAttempts {
			select {
			case <-time.After(p.Delay):
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("all %d attempt(s) failed, last error: %w", p.MaxAttempts, lastErr)
}

// WithMaxAttempts returns a copy of the policy with MaxAttempts set to n.
// If n is less than 1, it is clamped to 1.
func (p RetryPolicy) WithMaxAttempts(n int) RetryPolicy {
	if n < 1 {
		n = 1
	}
	p.MaxAttempts = n
	return p
}

// WithDelay returns a copy of the policy with Delay set to d.
func (p RetryPolicy) WithDelay(d time.Duration) RetryPolicy {
	p.Delay = d
	return p
}
