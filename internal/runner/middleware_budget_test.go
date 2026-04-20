package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithBudget_AllowsCallWhenBudgetAvailable(t *testing.T) {
	b := NewBudgetPolicy(5, time.Minute)
	mw := WithBudget(b)
	fn := mw(func(_ context.Context, _ Call) (string, error) {
		return "ok", nil
	})
	res, err := fn(context.Background(), Call{})
	if err != nil || res != "ok" {
		t.Fatalf("unexpected err=%v res=%s", err, res)
	}
}

func TestWithBudget_RejectsWhenExhausted(t *testing.T) {
	b := NewBudgetPolicy(2, time.Minute)
	b.Record()
	b.Record()
	mw := WithBudget(b)
	fn := mw(func(_ context.Context, _ Call) (string, error) {
		return "ok", nil
	})
	_, err := fn(context.Background(), Call{})
	if !errors.Is(err, ErrBudgetExhausted) {
		t.Fatalf("expected ErrBudgetExhausted, got %v", err)
	}
}

func TestWithBudget_RecordsFailure(t *testing.T) {
	b := NewBudgetPolicy(3, time.Minute)
	mw := WithBudget(b)
	fn := mw(func(_ context.Context, _ Call) (string, error) {
		return "", errors.New("fail")
	})
	_, _ = fn(context.Background(), Call{})
	if got := b.Remaining(); got != 2 {
		t.Fatalf("expected remaining=2, got %d", got)
	}
}

func TestWithBudget_SuccessDoesNotConsume(t *testing.T) {
	b := NewBudgetPolicy(3, time.Minute)
	mw := WithBudget(b)
	fn := mw(func(_ context.Context, _ Call) (string, error) {
		return "ok", nil
	})
	_, _ = fn(context.Background(), Call{})
	if got := b.Remaining(); got != 3 {
		t.Fatalf("expected remaining=3, got %d", got)
	}
}
