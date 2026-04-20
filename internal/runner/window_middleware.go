package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/isobit/grpcurl-batch/internal/manifest"
)

// WithWindowBreaker wraps a CallFunc and opens a circuit when the failure
// rate within the window exceeds maxFailRate (0.0–1.0).
func WithWindowBreaker(w *WindowPolicy, minCalls int, maxFailRate float64) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, c manifest.Call) (string, error) {
			s := w.Stats()
			if s.Total >= minCalls {
				rate := float64(s.Failures) / float64(s.Total)
				if rate > maxFailRate {
					return "", fmt.Errorf("window breaker open: failure rate %.2f exceeds %.2f", rate, maxFailRate)
				}
			}
			out, err := next(ctx, c)
			w.Record(err == nil)
			return out, err
		}
	}
}

// NewWindowPolicy30s is a convenience constructor for a 30-second window.
func NewWindowPolicy30s() *WindowPolicy {
	return NewWindowPolicy(30 * time.Second)
}
