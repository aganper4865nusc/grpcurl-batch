package runner

import (
	"context"
	"fmt"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// Middleware wraps a CallFunc with additional behaviour.
type Middleware func(next CallFunc) CallFunc

// CallFunc is a function that executes a single gRPC call.
type CallFunc func(ctx context.Context, call manifest.Call) (string, error)

// Chain applies middlewares in order (first = outermost).
func Chain(base CallFunc, middlewares ...Middleware) CallFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		base = middlewares[i](base)
	}
	return base
}

// WithRateLimit wraps a CallFunc with rate limiting.
func WithRateLimit(rl *RateLimiter) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			if err := rl.Wait(ctx); err != nil {
				return "", fmt.Errorf("rate limit: %w", err)
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
			return tp.Do(ctx, func(ctx context.Context) (string, error) {
				return next(ctx, call)
			})
		}
	}
}

// WithBulkhead wraps a CallFunc with bulkhead concurrency control.
func WithBulkhead(b *BulkheadPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			return b.Do(ctx, func(ctx context.Context) (string, error) {
				return next(ctx, call)
			})
		}
	}
}

// WithDrain wraps a CallFunc tracking in-flight calls via DrainPolicy.
func WithDrain(d *DrainPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			d.Acquire()
			defer d.Release()
			return next(ctx, call)
		}
	}
}
