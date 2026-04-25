package runner

import (
	"context"
	"fmt"
)

// WithConcurrencyLimit wraps a CallFunc with a ConcurrencyLimiter, rejecting
// calls that would exceed the configured maximum in-flight count.
//
// Example:
//
//	limiter := NewConcurrencyLimiter(20)
//	chain := Chain(base, WithConcurrencyLimit(limiter))
func WithConcurrencyLimit(cl *ConcurrencyLimiter) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (Result, error) {
			if err := cl.Acquire(ctx); err != nil {
				return Result{
					Call:  call,
					Error: fmt.Errorf("concurrency limiter: %w", err),
				}, err
			}
			defer cl.Release()
			return next(ctx, call)
		}
	}
}

// ConcurrencyLimiterStatus returns a snapshot of the limiter's current state
// suitable for logging or metrics emission.
type ConcurrencyLimiterStatus struct {
	Inflight int64 `json:"inflight"`
	Max      int64 `json:"max"`
	Avail    int64 `json:"available"`
}

// ConcurrencyStatus snapshots the given limiter.
func ConcurrencyStatus(cl *ConcurrencyLimiter) ConcurrencyLimiterStatus {
	inflight := cl.Inflight()
	max := cl.Max()
	avail := max - inflight
	if avail < 0 {
		avail = 0
	}
	return ConcurrencyLimiterStatus{
		Inflight: inflight,
		Max:      max,
		Avail:    avail,
	}
}
