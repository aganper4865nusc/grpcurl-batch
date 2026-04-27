package runner

import (
	"context"
	"fmt"
	"time"
)

// AdaptiveBreakerMiddlewareOption configures WithAdaptiveBreaker.
type AdaptiveBreakerMiddlewareOption struct {
	Breaker *AdaptiveBreaker
}

// AdaptiveBreakerStatus is returned by AdaptiveBreakerSnapshot.
type AdaptiveBreakerStatus struct {
	State       string        `json:"state"`
	WindowStats WindowStats   `json:"window_stats"`
	HalfOpenAfter time.Duration `json:"half_open_after_ns"`
}

// WithAdaptiveBreaker wraps a CallFunc with an adaptive circuit breaker that
// opens based on rolling-window error rate rather than a fixed failure count.
func WithAdaptiveBreaker(ab *AdaptiveBreaker) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (Result, error) {
			if err := ab.Allow(); err != nil {
				return Result{
					Call:    call,
					Err:     err,
					Elapsed: 0,
				}, err
			}
			res, err := next(ctx, call)
			ab.Record(err == nil)
			return res, err
		}
	}
}

// DefaultAdaptiveBreaker returns an AdaptiveBreaker with sensible defaults.
func DefaultAdaptiveBreaker() *AdaptiveBreaker {
	return NewAdaptiveBreaker(AdaptiveBreakerConfig{
		MinRequests:        10,
		ErrorRateThreshold: 0.5,
		HalfOpenAfter:      15 * time.Second,
		WindowSize:         30 * time.Second,
	})
}

// AdaptiveBreakerSnapshot returns a human-readable status string for the breaker.
func AdaptiveBreakerSnapshot(ab *AdaptiveBreaker) string {
	stats := ab.window.Stats()
	return fmt.Sprintf(
		"adaptive_breaker state=%s total=%d errors=%d error_rate=%.2f",
		ab.State(), stats.Total, stats.Errors, stats.ErrorRate(),
	)
}
