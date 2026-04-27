package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAdaptiveBreaker_InitiallyClosed(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{})
	if ab.State() != "closed" {
		t.Fatalf("expected closed, got %s", ab.State())
	}
	if err := ab.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAdaptiveBreaker_OpensAfterThreshold(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{
		MinRequests:        3,
		ErrorRateThreshold: 0.5,
		WindowSize:         10 * time.Second,
	})
	ab.Record(false)
	ab.Record(false)
	ab.Record(false)
	if ab.State() != "open" {
		t.Fatalf("expected open, got %s", ab.State())
	}
	if err := ab.Allow(); !errors.Is(err, ErrAdaptiveBreakerOpen) {
		t.Fatalf("expected ErrAdaptiveBreakerOpen, got %v", err)
	}
}

func TestAdaptiveBreaker_DoesNotOpenBelowMinRequests(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{
		MinRequests:        5,
		ErrorRateThreshold: 0.5,
		WindowSize:         10 * time.Second,
	})
	ab.Record(false)
	ab.Record(false)
	if ab.State() != "closed" {
		t.Fatalf("expected closed, got %s", ab.State())
	}
}

func TestAdaptiveBreaker_ResetsOnSuccessInHalfOpen(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{
		MinRequests:        2,
		ErrorRateThreshold: 0.5,
		HalfOpenAfter:      1 * time.Millisecond,
		WindowSize:         10 * time.Second,
	})
	ab.Record(false)
	ab.Record(false)
	if ab.State() != "open" {
		t.Fatal("expected open")
	}
	time.Sleep(5 * time.Millisecond)
	// Allow transitions to half-open
	if err := ab.Allow(); err != nil {
		t.Fatalf("expected probe allowed, got %v", err)
	}
	ab.Record(true)
	if ab.State() != "closed" {
		t.Fatalf("expected closed after success in half-open, got %s", ab.State())
	}
}

func TestWithAdaptiveBreaker_AllowsCallWhenClosed(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{MinRequests: 10})
	mw := WithAdaptiveBreaker(ab)
	call := Call{Service: "svc", Method: "m", Address: "localhost:50051"}
	ok := func(_ context.Context, c Call) (Result, error) {
		return Result{Call: c}, nil
	}
	res, err := mw(ok)(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Call.Service != "svc" {
		t.Fatalf("unexpected call service: %s", res.Call.Service)
	}
}

func TestWithAdaptiveBreaker_BlocksCallWhenOpen(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{
		MinRequests:        2,
		ErrorRateThreshold: 0.5,
		HalfOpenAfter:      1 * time.Hour,
		WindowSize:         10 * time.Second,
	})
	ab.Record(false)
	ab.Record(false)
	mw := WithAdaptiveBreaker(ab)
	call := Call{Service: "svc", Method: "m", Address: "localhost:50051"}
	noop := func(_ context.Context, c Call) (Result, error) { return Result{Call: c}, nil }
	_, err := mw(noop)(context.Background(), call)
	if !errors.Is(err, ErrAdaptiveBreakerOpen) {
		t.Fatalf("expected ErrAdaptiveBreakerOpen, got %v", err)
	}
}

func TestAdaptiveBreakerSnapshot_ContainsState(t *testing.T) {
	ab := NewAdaptiveBreaker(AdaptiveBreakerConfig{})
	s := AdaptiveBreakerSnapshot(ab)
	if len(s) == 0 {
		t.Fatal("expected non-empty snapshot")
	}
}
