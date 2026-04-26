package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWarmupPolicy_SuccessOnFirstProbe(t *testing.T) {
	w := NewWarmupPolicy(func(_ context.Context) error { return nil }, 3, 10*time.Millisecond)
	if err := w.Warm(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !w.IsWarmed() {
		t.Fatal("expected warmed=true")
	}
}

func TestWarmupPolicy_SuccessAfterRetry(t *testing.T) {
	var calls int32
	probe := func(_ context.Context) error {
		if atomic.AddInt32(&calls, 1) < 3 {
			return errors.New("not ready")
		}
		return nil
	}
	w := NewWarmupPolicy(probe, 5, 5*time.Millisecond)
	if err := w.Warm(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWarmupPolicy_ExhaustsAttempts(t *testing.T) {
	w := NewWarmupPolicy(func(_ context.Context) error {
		return errors.New("always failing")
	}, 3, 5*time.Millisecond)
	err := w.Warm(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if w.IsWarmed() {
		t.Fatal("expected warmed=false")
	}
}

func TestWarmupPolicy_SecondCallIsNoop(t *testing.T) {
	var calls int32
	w := NewWarmupPolicy(func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}, 3, 5*time.Millisecond)
	_ = w.Warm(context.Background())
	_ = w.Warm(context.Background())
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected probe called once, got %d", calls)
	}
}

func TestWarmupPolicy_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := NewWarmupPolicy(func(_ context.Context) error {
		return errors.New("not ready")
	}, 5, 5*time.Millisecond)
	err := w.Warm(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWarmupPolicy_Reset_AllowsRewarm(t *testing.T) {
	var calls int32
	w := NewWarmupPolicy(func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}, 3, 5*time.Millisecond)
	_ = w.Warm(context.Background())
	w.Reset()
	if w.IsWarmed() {
		t.Fatal("expected warmed=false after reset")
	}
	_ = w.Warm(context.Background())
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 probe calls, got %d", calls)
	}
}

func TestWarmupPolicy_DefaultsApplied(t *testing.T) {
	w := NewWarmupPolicy(func(_ context.Context) error { return nil }, 0, 0)
	if w.attempts != 3 {
		t.Fatalf("expected default attempts=3, got %d", w.attempts)
	}
	if w.delay != 500*time.Millisecond {
		t.Fatalf("expected default delay=500ms, got %v", w.delay)
	}
}
