package runner

import (
	"context"
	"fmt"
)

// WithRetryBudget wraps a call function so that each retry attempt consumes
// one token from the shared RetryBudget. The first attempt is free; only
// subsequent attempts (retries) require a token.
//
// Usage: compose with Chain alongside WithRetry or similar retry middleware.
func WithRetryBudget(rb *RetryBudget) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			result, err := next(ctx, call)
			if err != nil {
				// The call failed; consume a retry token before the caller retries.
				if budgetErr := rb.Consume(); budgetErr != nil {
					return "", fmt.Errorf("%w: original error: %v", ErrRetryBudgetExhausted, err)
				}
			}
			return result, err
		}
	}
}

// RetryBudgetStatus returns a snapshot of the current retry budget state
// for observability or logging purposes.
type RetryBudgetStatus struct {
	Used      int64
	Remaining int64
	Unlimited bool
}

// RetryBudgetSnapshot captures the current state of a RetryBudget.
func RetryBudgetSnapshot(rb *RetryBudget) RetryBudgetStatus {
	remaining := rb.Remaining()
	return RetryBudgetStatus{
		Used:      rb.Used(),
		Remaining: remaining,
		Unlimited: remaining == -1,
	}
}
