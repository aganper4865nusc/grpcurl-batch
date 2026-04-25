package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// PassthroughPolicy conditionally bypasses all middleware and executes a call
// directly when a predicate is satisfied. Useful for health-check or dry-run
// calls that must not be throttled, rate-limited, or circuit-broken.
type PassthroughPolicy struct {
	predicate func(Call) bool
	bypassed   atomic.Int64
}

// NewPassthroughPolicy creates a PassthroughPolicy that bypasses middleware
// for any call that satisfies predicate. If predicate is nil, no calls are
// bypassed.
func NewPassthroughPolicy(predicate func(Call) bool) *PassthroughPolicy {
	if predicate == nil {
		predicate = func(Call) bool { return false }
	}
	return &PassthroughPolicy{predicate: predicate}
}

// Bypassed returns the total number of calls that were passed through.
func (p *PassthroughPolicy) Bypassed() int64 {
	return p.bypassed.Load()
}

// Wrap returns a CallFunc that either invokes direct (bypassing the wrapped
// middleware chain) or falls through to next, depending on the predicate.
func (p *PassthroughPolicy) Wrap(direct, next CallFunc) CallFunc {
	if direct == nil {
		return next
	}
	if next == nil {
		return direct
	}
	return func(ctx context.Context, c Call) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if p.predicate(c) {
			p.bypassed.Add(1)
			return direct(ctx, c)
		}
		return next(ctx, c)
	}
}

// ErrPassthroughDirect is returned when a passthrough call is attempted but
// no direct executor is configured.
var ErrPassthroughDirect = errors.New("passthrough: no direct executor configured")
