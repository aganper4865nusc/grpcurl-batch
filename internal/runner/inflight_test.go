package runner

import (
	"context"
	"sync"
	"testing"
)

func TestInflightTracker_ZeroMax_DefaultsTo100(t *testing.T) {
	tr := NewInflightTracker(0)
	snap := tr.Snapshot()
	if snap.Max != 100 {
		t.Fatalf("expected max=100, got %d", snap.Max)
	}
}

func TestInflightTracker_AcquireRelease(t *testing.T) {
	tr := NewInflightTracker(5)
	release, err := tr.Acquire(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Current() != 1 {
		t.Fatalf("expected current=1, got %d", tr.Current())
	}
	release()
	if tr.Current() != 0 {
		t.Fatalf("expected current=0 after release, got %d", tr.Current())
	}
}

func TestInflightTracker_RejectsWhenFull(t *testing.T) {
	tr := NewInflightTracker(2)
	r1, _ := tr.Acquire(context.Background())
	r2, _ := tr.Acquire(context.Background())
	defer r1()
	defer r2()

	_, err := tr.Acquire(context.Background())
	if err == nil {
		t.Fatal("expected error when limit reached, got nil")
	}
}

func TestInflightTracker_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tr := NewInflightTracker(10)
	_, err := tr.Acquire(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestWithInflightLimit_PassesThrough(t *testing.T) {
	tr := NewInflightTracker(5)
	mw := WithInflightLimit(tr)
	called := false
	next := func(ctx context.Context, c Call) (Result, error) {
		called = true
		return Result{Call: c}, nil
	}
	_, err := mw(next)(context.Background(), Call{Method: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
	if tr.Current() != 0 {
		t.Fatalf("expected current=0 after call, got %d", tr.Current())
	}
}

func TestInflightTracker_ConcurrentSafety(t *testing.T) {
	tr := NewInflightTracker(50)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := tr.Acquire(context.Background())
			if err == nil {
				defer release()
			}
		}()
	}
	wg.Wait()
	if tr.Current() != 0 {
		t.Fatalf("expected current=0 after all goroutines done, got %d", tr.Current())
	}
}
