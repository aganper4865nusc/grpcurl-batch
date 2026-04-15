package runner

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected circuit to be closed, got: %v", err)
	}
	if cb.State() != "closed" {
		t.Fatalf("expected state closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Minute)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != "open" {
		t.Fatalf("expected state open, got %s", cb.State())
	}
	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_DoesNotOpenBeforeThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Minute)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != "closed" {
		t.Fatalf("expected closed after 2 failures, got %s", cb.State())
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(2, time.Minute)
	cb.RecordFailure()
	cb.RecordSuccess()
	if cb.State() != "closed" {
		t.Fatalf("expected closed after success, got %s", cb.State())
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected allow after reset, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	cb.RecordFailure()
	if cb.State() != "open" {
		t.Fatalf("expected open, got %s", cb.State())
	}
	time.Sleep(60 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected allow in half-open, got %v", err)
	}
	if cb.State() != "half-open" {
		t.Fatalf("expected half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpen_SuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	_ = cb.Allow() // transitions to half-open
	cb.RecordSuccess()
	if cb.State() != "closed" {
		t.Fatalf("expected closed after probe success, got %s", cb.State())
	}
}

func TestCircuitBreaker_ZeroThreshold_DefaultsToOne(t *testing.T) {
	cb := NewCircuitBreaker(0, time.Minute)
	cb.RecordFailure()
	if cb.State() != "open" {
		t.Fatalf("expected open with zero threshold defaulting to 1, got %s", cb.State())
	}
}
