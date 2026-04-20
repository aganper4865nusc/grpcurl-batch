package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBudget_ZeroMax_DefaultsToOne(t *testing.T) {
	p := NewBudgetPolicy(0, time.Minute)
	if p.max != 1 {
		t.Fatalf("expected max=1, got %d", p.max)
	}
}

func TestBudget_InitialRemaining(t *testing.T) {
	p := NewBudgetPolicy(3, time.Minute)
	if got := p.Remaining(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestBudget_RecordDecrementsRemaining(t *testing.T) {
	p := NewBudgetPolicy(3, time.Minute)
	p.Record()
	if got := p.Remaining(); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestBudget_ExhaustedRejectsCall(t *testing.T) {
	p := NewBudgetPolicy(2, time.Minute)
	p.Record()
	p.Record()
	err := p.Wrap(context.Background(), func(_ context.Context) error { return nil })
	if !errors.Is(err, ErrBudgetExhausted) {
		t.Fatalf("expected ErrBudgetExhausted, got %v", err)
	}
}

func TestBudget_SuccessDoesNotConsumeBudget(t *testing.T) {
	p := NewBudgetPolicy(3, time.Minute)
	_ = p.Wrap(context.Background(), func(_ context.Context) error { return nil })
	if got := p.Remaining(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestBudget_FailureConsumes(t *testing.T) {
	p := NewBudgetPolicy(3, time.Minute)
	_ = p.Wrap(context.Background(), func(_ context.Context) error { return errors.New("boom") })
	if got := p.Remaining(); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestBudget_ResetsAfterWindow(t *testing.T) {
	p := NewBudgetPolicy(2, 50*time.Millisecond)
	p.Record()
	p.Record()
	time.Sleep(60 * time.Millisecond)
	if got := p.Remaining(); got != 2 {
		t.Fatalf("expected 2 after reset, got %d", got)
	}
}
