package runner

import (
	"context"
	"time"
)

// TimeoutPolicy enforces a maximum duration on call execution.
type TimeoutPolicy struct {
	Timeout time.Duration
}

// DefaultTimeoutPolicy creates a TimeoutPolicy with the given duration.
// A zero timeout disables the deadline.
func DefaultTimeoutPolicy(timeout time.Duration) *TimeoutPolicy {
	return &TimeoutPolicy{Timeout: timeout}
}

// Execute runs fn within the configured timeout.
func (tp *TimeoutPolicy) Execute(ctx context.Context, fn func(context.Context) (string, error)) (string, error) {
	if tp.Timeout <= 0 {
		return fn(ctx)
	}
	ctx, cancel := context.WithTimeout(ctx, tp.Timeout)
	defer cancel()

	type result struct {
		val string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, e := fn(ctx)
		ch <- result{v, e}
	}()
	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
