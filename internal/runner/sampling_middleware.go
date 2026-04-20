package runner

import (
	"context"
	"errors"

	"github.com/yourorg/grpcurl-batch/internal/manifest"
)

// ErrCallSkipped is returned when a call is dropped by the sampling policy.
var ErrCallSkipped = errors.New("call skipped by sampling policy")

// WithSampling returns a middleware that probabilistically skips calls
// according to the given SamplingPolicy. Skipped calls return ErrCallSkipped
// so the caller can distinguish them from genuine failures.
func WithSampling(p *SamplingPolicy) func(CallFunc) CallFunc {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			if !p.Sampled() {
				return "", ErrCallSkipped
			}
			return next(ctx, call)
		}
	}
}
