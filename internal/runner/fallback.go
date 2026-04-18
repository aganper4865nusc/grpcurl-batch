package runner

import (
	"context"
	"errors"
)

// FallbackPolicy executes a primary call and, on failure, delegates to a
// fallback handler that may return a cached or default response.
type FallbackPolicy struct {
	fallbackFn func(ctx context.Context, call Call, err error) (string, error)
}

// NewFallbackPolicy creates a FallbackPolicy with the given fallback function.
// If fallbackFn is nil, the original error is always returned.
func NewFallbackPolicy(fallbackFn func(ctx context.Context, call Call, err error) (string, error)) *FallbackPolicy {
	if fallbackFn == nil {
		fallbackFn = func(_ context.Context, _ Call, err error) (string, error) {
			return "", err
		}
	}
	return &FallbackPolicy{fallbackFn: fallbackFn}
}

// Execute runs fn and invokes the fallback if fn returns an error.
// If the context is already cancelled the fallback is skipped.
func (f *FallbackPolicy) Execute(ctx context.Context, call Call, fn func(ctx context.Context) (string, error)) (string, error) {
	result, err := fn(ctx)
	if err == nil {
		return result, nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "", err
	}
	return f.fallbackFn(ctx, call, err)
}

// StaticFallback returns a FallbackPolicy that always responds with a fixed
// payload regardless of the original error.
func StaticFallback(payload string) *FallbackPolicy {
	return NewFallbackPolicy(func(_ context.Context, _ Call, _ error) (string, error) {
		return payload, nil
	})
}
