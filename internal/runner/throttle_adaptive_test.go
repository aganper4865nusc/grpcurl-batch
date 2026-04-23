package runner

import (
	"context"
	"testing"
	"time"
)

func newTestWindow() *WindowPolicy {
	return NewWindowPolicy(10 * time.Second)
}

func TestAdaptiveThrottle_InitialCurrentEqualsMax(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(8, 0.5, 0.1, window)
	defer at.Stop()
	if got := at.Current(); got != 8 {
		t.Fatalf("expected initial current=8, got %d", got)
	}
}

func TestAdaptiveThrottle_ZeroMax_DefaultsToOne(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(0, 0.5, 0.1, window)
	defer at.Stop()
	if got := at.Current(); got != 1 {
		t.Fatalf("expected current=1, got %d", got)
	}
}

func TestAdaptiveThrottle_HighErrorRate_HalvesLimit(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(8, 0.5, 0.1, window)
	defer at.Stop()
	// Simulate 60% error rate
	for i := 0; i < 4; i++ {
		window.Record(false)
	}
	for i := 0; i < 6; i++ {
		window.Record(true)
	}
	at.adjust()
	if got := at.Current(); got != 4 {
		t.Fatalf("expected current=4 after halving, got %d", got)
	}
}

func TestAdaptiveThrottle_LowErrorRate_GrowsLimit(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(8, 0.5, 0.1, window)
	defer at.Stop()
	at.mu.Lock()
	at.current = 4
	at.mu.Unlock()
	// Simulate 5% error rate
	for i := 0; i < 95; i++ {
		window.Record(false)
	}
	for i := 0; i < 5; i++ {
		window.Record(true)
	}
	at.adjust()
	if got := at.Current(); got != 5 {
		t.Fatalf("expected current=5 after growth, got %d", got)
	}
}

func TestAdaptiveThrottle_DoesNotExceedMax(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(4, 0.5, 0.1, window)
	defer at.Stop()
	// Already at max, low error rate should not exceed max
	for i := 0; i < 100; i++ {
		window.Record(false)
	}
	at.adjust()
	if got := at.Current(); got > 4 {
		t.Fatalf("current %d exceeds max 4", got)
	}
}

func TestAdaptiveThrottle_Wrap_ExecutesCall(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(4, 0.5, 0.1, window)
	defer at.Stop()
	called := false
	err := at.Wrap(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected fn to be called")
	}
}

func TestAdaptiveThrottle_Wrap_CancelledContext(t *testing.T) {
	window := newTestWindow()
	at := NewAdaptiveThrottle(1, 0.5, 0.1, window)
	defer at.Stop()
	// Saturate the single slot
	blockCh := make(chan struct{})
	go func() {
		_ = at.Wrap(context.Background(), func(_ context.Context) error {
			<-blockCh
			return nil
		})
	}()
	time.Sleep(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := at.Wrap(ctx, func(_ context.Context) error { return nil })
	close(blockCh)
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
}
