package runner

import (
	"context"
	"errors"
)

// WithBudget wraps a call with a BudgetPolicy, rejecting calls when the error
// budget is exhausted and recording failures against the budget.
func WithBudget(b *BudgetPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			var (
				result string
				callErr error
			)
			err := b.Wrap(ctx, func(c context.Context) error {
				result, callErr = next(c, call)
				return callErr
			})
			if errors.Is(err, ErrBudgetExhausted) {
				return "", ErrBudgetExhausted
			}
			return result, callErr
		}
	}
}
