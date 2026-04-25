package runner

import (
	"sync"
	"testing"
)

func TestRetryBudget_ZeroMax_Unlimited(t *testing.T) {
	rb := NewRetryBudget(0)
	for i := 0; i < 1000; i++ {
		if err := rb.Consume(); err != nil {
			t.Fatalf("expected no error on unlimited budget, got %v", err)
		}
	}
}

func TestRetryBudget_NegativeMax_Unlimited(t *testing.T) {
	rb := NewRetryBudget(-5)
	if err := rb.Consume(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRetryBudget_AllowsUpToMax(t *testing.T) {
	rb := NewRetryBudget(3)
	for i := 0; i < 3; i++ {
		if err := rb.Consume(); err != nil {
			t.Fatalf("attempt %d: expected success, got %v", i, err)
		}
	}
}

func TestRetryBudget_ExhaustedReturnsError(t *testing.T) {
	rb := NewRetryBudget(2)
	_ = rb.Consume()
	_ = rb.Consume()
	if err := rb.Consume(); err != ErrRetryBudgetExhausted {
		t.Fatalf("expected ErrRetryBudgetExhausted, got %v", err)
	}
}

func TestRetryBudget_Remaining(t *testing.T) {
	rb := NewRetryBudget(5)
	if rb.Remaining() != 5 {
		t.Fatalf("expected 5, got %d", rb.Remaining())
	}
	_ = rb.Consume()
	_ = rb.Consume()
	if rb.Remaining() != 3 {
		t.Fatalf("expected 3, got %d", rb.Remaining())
	}
}

func TestRetryBudget_Remaining_Unlimited(t *testing.T) {
	rb := NewRetryBudget(0)
	if rb.Remaining() != -1 {
		t.Fatalf("expected -1 sentinel for unlimited, got %d", rb.Remaining())
	}
}

func TestRetryBudget_Reset_ClearsUsed(t *testing.T) {
	rb := NewRetryBudget(2)
	_ = rb.Consume()
	_ = rb.Consume()
	rb.Reset()
	if rb.Used() != 0 {
		t.Fatalf("expected 0 after reset, got %d", rb.Used())
	}
	if err := rb.Consume(); err != nil {
		t.Fatalf("expected success after reset, got %v", err)
	}
}

func TestRetryBudget_ConcurrentSafety(t *testing.T) {
	rb := NewRetryBudget(50)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rb.Consume()
		}()
	}
	wg.Wait()
	if rb.Used() > 50 {
		t.Fatalf("used %d exceeds max 50", rb.Used())
	}
}
