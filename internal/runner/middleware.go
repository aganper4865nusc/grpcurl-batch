package runner

import (
	"context"
	"errors"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// Middleware wraps a CallFunc with additional behaviour.
type Middleware func(CallFunc) CallFunc

// CallFunc is a function that executes a single gRPC call.
type CallFunc func(ctx context.Context, call manifest.Call) (string, error)

// Chain applies middlewares in order (first middleware is outermost).
func Chain(base CallFunc, mws ...Middleware) CallFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		base = mws[i](base)
	}
	return base
}

// WithRateLimit wraps a CallFunc with rate limiting.
func WithRateLimit(rl *RateLimiter) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			if err := rl.Wait(ctx); err != nil {
				return "", err
			}
			return next(ctx, call)
		}
	}
}

// WithCircuitBreaker wraps a CallFunc with circuit breaker protection.
func WithCircuitBreaker(cb *CircuitBreaker) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			return cb.Do(ctx, func(ctx context.Context) (string, error) {
				return next(ctx, call)
			})
		}
	}
}

// WithTimeout wraps a CallFunc with a timeout policy.
func WithTimeout(tp *TimeoutPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			var result string
			err := tp.Do(ctx, func(ctx context.Context) error {
				var e error
				result, e = next(ctx, call)
				return e
			})
			return result, err
		}
	}
}

// WithBulkhead wraps a CallFunc with bulkhead isolation.
func WithBulkhead(b *BulkheadPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			var result string
			err := b.Do(ctx, func(ctx context.Context) error {
				var e error
				result, e = next(ctx, call)
				return e
			})
			if errors.Is(err, ErrBulkheadFull) {
				return "", err
			}
			return result, err
		}
	}
}
