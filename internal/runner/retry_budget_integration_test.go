package runner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

// TestRetryBudget_ConcurrentConsume_NeverExceedsMax verifies that under high
// concurrency the budget is never over-consumed.
func TestRetryBudget_ConcurrentConsume_NeverExceedsMax(t *testing.T) {
	const max = 100
	rb := NewRetryBudget(max)
	var wg sync.WaitGroup
	var succeeded atomic.Int64
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rb.Consume() == nil {
				succeeded.Add(1)
			}
		}()
	}
	wg.Wait()
	if succeeded.Load() > max {
		t.Fatalf("succeeded %d exceeds max %d", succeeded.Load(), max)
	}
	if rb.Used() != succeeded.Load() {
		t.Fatalf("Used() %d != succeeded %d", rb.Used(), succeeded.Load())
	}
}

// TestRetryBudget_MiddlewareChain_StopsRetries verifies that the middleware
// correctly blocks further retries once the budget is exhausted in a chain.
func TestRetryBudget_MiddlewareChain_StopsRetries(t *testing.T) {
	rb := NewRetryBudget(2)
	mw := WithRetryBudget(rb)
	failFn := func(_ context.Context, _ Call) (string, error) {
		return "", errors.New("always fails")
	}
	wrapped := mw(failFn)
	ctx := context.Background()
	// First two failures consume the budget.
	_, _ = wrapped(ctx, Call{})
	_, _ = wrapped(ctx, Call{})
	// Third should be rejected by budget.
	_, err := wrapped(ctx, Call{})
	if !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("expected budget exhausted on 3rd call, got %v", err)
	}
}
