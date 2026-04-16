package runner

import "context"

// CallFunc represents a gRPC call execution function.
type CallFunc func(ctx context.Context) (string, error)

// Middleware wraps a CallFunc with additional behavior.
type Middleware func(next CallFunc) CallFunc

// Chain applies middlewares in order (first middleware is outermost).
func Chain(middlewares ...Middleware) Middleware {
	return func(next CallFunc) CallFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// WithRateLimit returns a Middleware that applies rate limiting.
func WithRateLimit(rl *RateLimiter) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context) (string, error) {
			if err := rl.Wait(ctx); err != nil {
				return "", err
			}
			return next(ctx)
		}
	}
}

// WithCircuitBreaker returns a Middleware that applies circuit breaking.
func WithCircuitBreaker(cb *CircuitBreaker) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context) (string, error) {
			var result string
			err := cb.Execute(func() error {
				var callErr error
				result, callErr = next(ctx)
				return callErr
			})
			return result, err
		}
	}
}

// WithTimeout returns a Middleware that applies a timeout policy.
func WithTimeout(tp *TimeoutPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context) (string, error) {
			return tp.Execute(ctx, func(ctx context.Context) (string, error) {
				return next(ctx)
			})
		}
	}
}
