package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeoutBudget_ZeroDuration_NeverExpires(t *testing.T) {
	tb := NewTimeoutBudget(0)
	if tb.IsExhausted() {
		t.Fatal("expected budget with zero duration to never expire")
	}
	if tb.Remaining() < time.Hour {
		t.Fatalf("expected large remaining, got %v", tb.Remaining())
	}
}

func TestTimeoutBudget_FutureDuration_NotExhausted(t *testing.T) {
	tb := NewTimeoutBudget(10 * time.Second)
	if tb.IsExhausted() {
		t.Fatal("expected budget to not be exhausted")
	}
	if tb.Remaining() <= 0 {
		t.Fatal("expected positive remaining time")
	}
}

func TestTimeoutBudget_Expired_IsExhausted(t *testing.T) {
	tb := NewTimeoutBudget(1 * time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if !tb.IsExhausted() {
		t.Fatal("expected budget to be exhausted after expiry")
	}
	if tb.Remaining() != 0 {
		t.Fatalf("expected zero remaining, got %v", tb.Remaining())
	}
}

func TestTimeoutBudget_Wrap_ExhaustedReturnsError(t *testing.T) {
	tb := NewTimeoutBudget(1 * time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	err := tb.Wrap(context.Background(), func(_ context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrTimeoutBudgetExhausted) {
		t.Fatalf("expected ErrTimeoutBudgetExhausted, got %v", err)
	}
}

func TestTimeoutBudget_Wrap_PropagatesError(t *testing.T) {
	tb := NewTimeoutBudget(5 * time.Second)
	sentinel := errors.New("inner error")
	err := tb.Wrap(context.Background(), func(_ context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestTimeoutBudget_Wrap_NoBudget_PassesThrough(t *testing.T) {
	tb := NewTimeoutBudget(0)
	called := false
	_ = tb.Wrap(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})
	if !called {
		t.Fatal("expected fn to be called")
	}
}

func TestWithTimeoutBudget_Middleware_AllowsCall(t *testing.T) {
	tb := NewTimeoutBudget(5 * time.Second)
	mw := WithTimeoutBudget(tb)
	inner := mw(func(ctx context.Context, c Call) (Result, error) {
		return Result{CallName: c.Name, Success: true}, nil
	})
	res, err := inner(context.Background(), Call{Name: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatal("expected success")
	}
}

func TestWithTimeoutBudget_Middleware_RejectsWhenExhausted(t *testing.T) {
	tb := NewTimeoutBudget(1 * time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	mw := WithTimeoutBudget(tb)
	inner := mw(func(ctx context.Context, c Call) (Result, error) {
		return Result{Success: true}, nil
	})
	_, err := inner(context.Background(), Call{Name: "test"})
	if !errors.Is(err, ErrTimeoutBudgetExhausted) {
		t.Fatalf("expected ErrTimeoutBudgetExhausted, got %v", err)
	}
}
