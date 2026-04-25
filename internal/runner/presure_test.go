package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestBackpressure_BelowHighWater_Allows(t *testing.T) {
	p := NewBackpressurePolicy(5, 2)
	res, err := p.Wrap(context.Background(), Call{Method: "m"}, func(_ context.Context, c Call) (Result, error) {
		return Result{Call: c}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Call.Method != "m" {
		t.Fatalf("expected method m, got %s", res.Call.Method)
	}
}

func TestBackpressure_AtHighWater_Rejects(t *testing.T) {
	p := NewBackpressurePolicy(2, 1)

	gate := make(chan struct{})
	var wg sync.WaitGroup

	// Occupy both slots.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = p.Wrap(context.Background(), Call{}, func(_ context.Context, c Call) (Result, error) {
				<-gate
				return Result{}, nil
			})
		}()
	}

	// Give goroutines time to enter Wrap.
	for p.Inflight() < 2 {
		runtime_Gosched()
	}

	_, err := p.Wrap(context.Background(), Call{}, func(_ context.Context, c Call) (Result, error) {
		return Result{}, nil
	})
	if !errors.Is(err, ErrBackpressure) {
		t.Fatalf("expected ErrBackpressure, got %v", err)
	}

	close(gate)
	wg.Wait()
}

func TestBackpressure_DeactivatesAfterDrain(t *testing.T) {
	p := NewBackpressurePolicy(2, 1)

	gate := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			_, _ = p.Wrap(context.Background(), Call{}, func(_ context.Context, c Call) (Result, error) {
				<-gate
				return Result{}, nil
			})
		}()
	}
	for p.Inflight() < 2 {
		runtime_Gosched()
	}
	close(gate)
	wg.Wait()

	if p.IsOpen() {
		t.Fatal("expected backpressure to be inactive after drain")
	}
}

func TestBackpressure_CancelledContext_ReturnsErr(t *testing.T) {
	p := NewBackpressurePolicy(10, 5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.Wrap(ctx, Call{}, func(_ context.Context, c Call) (Result, error) {
		return Result{}, nil
	})
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestBackpressure_DefaultHighWater(t *testing.T) {
	p := NewBackpressurePolicy(0, 0)
	if p.highWater != 50 {
		t.Fatalf("expected default highWater 50, got %d", p.highWater)
	}
}

// runtime_Gosched is a thin wrapper so we don't import runtime in test files directly.
func runtime_Gosched() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { wg.Done() }()
	wg.Wait()
}
