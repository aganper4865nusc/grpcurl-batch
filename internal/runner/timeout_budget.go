package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrTimeoutBudgetExhausted is returned when the shared timeout budget is depleted.
var ErrTimeoutBudgetExhausted = errors.New("timeout budget exhausted")

// TimeoutBudget tracks a shared wall-clock deadline across multiple calls.
// Once the deadline passes, all subsequent calls are rejected immediately.
type TimeoutBudget struct {
	deadline time.Time
	exhausted atomic.Bool
}

// NewTimeoutBudget creates a TimeoutBudget that expires after the given duration.
// A zero or negative duration creates a budget that never expires.
func NewTimeoutBudget(total time.Duration) *TimeoutBudget {
	tb := &TimeoutBudget{}
	if total > 0 {
		tb.deadline = time.Now().Add(total)
	}
	return tb
}

// Remaining returns how much time is left in the budget.
// Returns a large duration if no deadline is set.
func (tb *TimeoutBudget) Remaining() time.Duration {
	if tb.deadline.IsZero() {
		return 24 * time.Hour
	}
	remaining := time.Until(tb.deadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsExhausted reports whether the budget deadline has passed.
func (tb *TimeoutBudget) IsExhausted() bool {
	if tb.exhausted.Load() {
		return true
	}
	if !tb.deadline.IsZero() && time.Now().After(tb.deadline) {
		tb.exhausted.Store(true)
		return true
	}
	return false
}

// Wrap executes fn within a context bounded by the remaining budget.
// Returns ErrTimeoutBudgetExhausted if the budget is already depleted.
func (tb *TimeoutBudget) Wrap(ctx context.Context, fn func(context.Context) error) error {
	if tb.IsExhausted() {
		return ErrTimeoutBudgetExhausted
	}
	if tb.deadline.IsZero() {
		return fn(ctx)
	}
	ctx, cancel := context.WithDeadline(ctx, tb.deadline)
	defer cancel()
	return fn(ctx)
}

// WithTimeoutBudget returns a middleware that enforces a shared TimeoutBudget.
func WithTimeoutBudget(tb *TimeoutBudget) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (Result, error) {
			var res Result
			err := tb.Wrap(ctx, func(bctx context.Context) error {
				var e error
				res, e = next(bctx, call)
				return e
			})
			return res, err
		}
	}
}
