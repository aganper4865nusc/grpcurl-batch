package runner

import (
	"context"
	"fmt"
	"time"
)

// TimeoutPolicy wraps a call with a per-call deadline.
type TimeoutPolicy struct {
	Timeout time.Duration
}

// DefaultTimeoutPolicy returns a TimeoutPolicy with a 30-second deadline.
func DefaultTimeoutPolicy() TimeoutPolicy {
	return TimeoutPolicy{Timeout: 30 * time.Second}
}

// Apply executes fn within the configured timeout. If the timeout is zero or
// negative the call is executed without a deadline. Returns an error that
// wraps context.DeadlineExceeded when the deadline is exceeded.
func (tp TimeoutPolicy) Apply(ctx context.Context, fn func(ctx context.Context) error) error {
	if tp.Timeout <= 0 {
		return fn(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, tp.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("call timed out after %s: %w", tp.Timeout, ctx.Err())
	}
}
