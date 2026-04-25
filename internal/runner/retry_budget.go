package runner

import (
	"errors"
	"sync/atomic"
)

// ErrRetryBudgetExhausted is returned when the retry budget has been consumed.
var ErrRetryBudgetExhausted = errors.New("retry budget exhausted")

// RetryBudget limits the total number of retries allowed across all calls
// within a batch run, preventing retry storms.
type RetryBudget struct {
	max      int64
	used     atomic.Int64
	disabled bool
}

// NewRetryBudget creates a RetryBudget with the given maximum retry count.
// A max of zero disables the budget (unlimited retries).
func NewRetryBudget(max int) *RetryBudget {
	if max < 0 {
		max = 0
	}
	return &RetryBudget{
		max:      int64(max),
		disabled: max == 0,
	}
}

// Consume attempts to consume one retry token.
// Returns ErrRetryBudgetExhausted if the budget is depleted.
func (rb *RetryBudget) Consume() error {
	if rb.disabled {
		return nil
	}
	next := rb.used.Add(1)
	if next > rb.max {
		rb.used.Add(-1)
		return ErrRetryBudgetExhausted
	}
	return nil
}

// Remaining returns the number of retry tokens still available.
func (rb *RetryBudget) Remaining() int64 {
	if rb.disabled {
		return -1 // sentinel: unlimited
	}
	r := rb.max - rb.used.Load()
	if r < 0 {
		return 0
	}
	return r
}

// Reset clears all consumed tokens.
func (rb *RetryBudget) Reset() {
	rb.used.Store(0)
}

// Used returns the number of retries consumed so far.
func (rb *RetryBudget) Used() int64 {
	return rb.used.Load()
}
