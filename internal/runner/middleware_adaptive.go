package runner

import (
	"context"
	"fmt"
	"time"
)

// WithAdaptiveThrottle returns a Middleware that gates calls through an
// AdaptiveThrottle. The throttle adjusts its concurrency ceiling automatically
// based on the error rate reported to the supplied WindowPolicy.
//
// Example:
//
//	window := NewWindowPolicy(30 * time.Second)
//	mw := WithAdaptiveThrottle(NewAdaptiveThrottle(16, 0.4, 0.05, window), window)
func WithAdaptiveThrottle(at *AdaptiveThrottle, window *WindowPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (Response, error) {
			var resp Response
			err := at.Wrap(ctx, func(ctx context.Context) error {
				var callErr error
				resp, callErr = next(ctx, call)
				window.Record(callErr != nil)
				return callErr
			})
			return resp, err
		}
	}
}

// DefaultAdaptiveThrottle builds an AdaptiveThrottle with production-ready
// defaults: max=32, high-watermark=0.4, low-watermark=0.05, 30 s window.
func DefaultAdaptiveThrottle() (*AdaptiveThrottle, *WindowPolicy) {
	window := NewWindowPolicy(30 * time.Second)
	at := NewAdaptiveThrottle(32, 0.4, 0.05, window)
	return at, window
}

// AdaptiveThrottleStatus returns a human-readable status string for logging.
func AdaptiveThrottleStatus(at *AdaptiveThrottle, window *WindowPolicy) string {
	stats := window.Stats()
	var errRate float64
	if stats.Total > 0 {
		errRate = float64(stats.Failures) / float64(stats.Total)
	}
	return fmt.Sprintf("adaptive_throttle current=%d err_rate=%.2f total=%d",
		at.Current(), errRate, stats.Total)
}
