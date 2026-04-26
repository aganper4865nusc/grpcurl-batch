package runner

import (
	"context"
	"fmt"
)

// WithWarmup returns a middleware that ensures the WarmupPolicy has completed
// before forwarding the call. If warmup fails the call is rejected with the
// warmup error. Once warmed, subsequent calls pass through with zero overhead
// beyond the IsWarmed boolean check.
func WithWarmup(w *WarmupPolicy) func(CallFunc) CallFunc {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			if !w.IsWarmed() {
				if err := w.Warm(ctx); err != nil {
					return "", fmt.Errorf("warmup middleware: %w", err)
				}
			}
			return next(ctx, call)
		}
	}
}

// WarmupStatus returns a snapshot of the warmup state suitable for logging
// or metrics emission.
type WarmupStatus struct {
	Warmed bool
}

// WarmupSnapshot captures the current warmup state from the policy.
func WarmupSnapshot(w *WarmupPolicy) WarmupStatus {
	return WarmupStatus{Warmed: w.IsWarmed()}
}
