package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBulkhead_AllowsUpToMax(t *testing.T) {
	b := NewBulkheadPolicy(3)
	var wg sync.WaitGroup
	blocked := make(chan struct{})
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Do(context.Background(), func(ctx context.Context) error {
				<-blocked
				return nil
			})
		}()
	}
	time.Sleep(20 * time.Millisecond)
	if b.Active() != 3 {
		t.Fatalf("expected 3 active, got %d", b.Active())
	}
	close(blocked)
	wg.Wait()
}

func TestBulkhead_RejectsWhenFull(t *testing.T) {
	b := NewBulkheadPolicy(1)
	blocked := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			<-blocked
			return nil
		})
	}()
	time.Sleep(20 * time.Millisecond)
	err := b.Do(context.Background(), func(ctx context.Context) error { return nil })
	if err != ErrBulkheadFull {
		t.Fatalf("expected ErrBulkheadFull, got %v", err)
	}
	close(blocked)
	wg.Wait()
}

func TestBulkhead_ZeroMax_DefaultsToOne(t *testing.T) {
	b := NewBulkheadPolicy(0)
	if b.Max() != 1 {
		t.Fatalf("expected max=1, got %d", b.Max())
	}
}

func TestBulkhead_ReleasesAfterDone(t *testing.T) {
	b := NewBulkheadPolicy(1)
	_ = b.Do(context.Background(), func(ctx context.Context) error { return nil })
	if b.Active() != 0 {
		t.Fatalf("expected 0 active after completion, got %d", b.Active())
	}
}

func TestBulkhead_ReleasesAfterError(t *testing.T) {
	b := NewBulkheadPolicy(1)
	_ = b.Do(context.Background(), func(ctx context.Context) error {
		return context.DeadlineExceeded
	})
	if b.Active() != 0 {
		t.Fatalf("expected 0 active after error, got %d", b.Active())
	}
	// Slot should be free; a subsequent call must not be rejected.
	err := b.Do(context.Background(), func(ctx context.Context) error { return nil })
	if err != nil {
		t.Fatalf("expected nil after slot released, got %v", err)
	}
}

func TestBulkhead_ConcurrentSafety(t *testing.T) {
	b := NewBulkheadPolicy(10)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Do(context.Background(), func(ctx context.Context) error {
				time.Sleep(2 * time.Millisecond)
				return nil
			})
		}()
	}
	wg.Wait()
	if b.Active() != 0 {
		t.Fatalf("expected 0 active after all done, got %d", b.Active())
	}
}
