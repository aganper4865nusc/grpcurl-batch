package runner

import (
	"context"
	"errors"
	"testing"
)

func okRetryCall(_ context.Context, _ Call) (string, error) {
	return "ok", nil
}

func errRetryCall(_ context.Context, _ Call) (string, error) {
	return "", errors.New("rpc failed")
}

func TestWithRetryBudget_SuccessDoesNotConsume(t *testing.T) {
	rb := NewRetryBudget(3)
	mw := WithRetryBudget(rb)
	_, err := mw(okRetryCall)(context.Background(), Call{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rb.Used() != 0 {
		t.Fatalf("expected 0 consumed, got %d", rb.Used())
	}
}

func TestWithRetryBudget_FailureConsumesToken(t *testing.T) {
	rb := NewRetryBudget(5)
	mw := WithRetryBudget(rb)
	_, _ = mw(errRetryCall)(context.Background(), Call{})
	if rb.Used() != 1 {
		t.Fatalf("expected 1 consumed, got %d", rb.Used())
	}
}

func TestWithRetryBudget_ExhaustedReturnsWrappedError(t *testing.T) {
	rb := NewRetryBudget(1)
	_ = rb.Consume() // exhaust
	mw := WithRetryBudget(rb)
	_, err := mw(errRetryCall)(context.Background(), Call{})
	if !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("expected ErrRetryBudgetExhausted, got %v", err)
	}
}

func TestWithRetryBudget_UnlimitedNeverBlocks(t *testing.T) {
	rb := NewRetryBudget(0)
	mw := WithRetryBudget(rb)
	for i := 0; i < 200; i++ {
		_, err := mw(errRetryCall)(context.Background(), Call{})
		if errors.Is(err, ErrRetryBudgetExhausted) {
			t.Fatal("unlimited budget should never return ErrRetryBudgetExhausted")
		}
	}
}

func TestRetryBudgetSnapshot_Values(t *testing.T) {
	rb := NewRetryBudget(10)
	_ = rb.Consume()
	_ = rb.Consume()
	snap := RetryBudgetSnapshot(rb)
	if snap.Used != 2 {
		t.Fatalf("expected Used=2, got %d", snap.Used)
	}
	if snap.Remaining != 8 {
		t.Fatalf("expected Remaining=8, got %d", snap.Remaining)
	}
	if snap.Unlimited {
		t.Fatal("expected Unlimited=false")
	}
}

func TestRetryBudgetSnapshot_Unlimited(t *testing.T) {
	rb := NewRetryBudget(0)
	snap := RetryBudgetSnapshot(rb)
	if !snap.Unlimited {
		t.Fatal("expected Unlimited=true for zero-max budget")
	}
}
