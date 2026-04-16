package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSemaphore_AcquireRelease(t *testing.T) {
	s := NewSemaphore(2)
	ctx := context.Background()

	if err := s.Acquire(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Available() != 1 {
		t.Errorf("expected 1 available, got %d", s.Available())
	}
	s.Release()
	if s.Available() != 2 {
		t.Errorf("expected 2 available, got %d", s.Available())
	}
}

func TestSemaphore_ZeroCapacityDefaultsToOne(t *testing.T) {
	s := NewSemaphore(0)
	if s.Capacity() != 1 {
		t.Errorf("expected capacity 1, got %d", s.Capacity())
	}
}

func TestSemaphore_BlocksWhenFull(t *testing.T) {
	s := NewSemaphore(1)
	ctx := context.Background()

	_ = s.Acquire(ctx)

	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := s.Acquire(ctx2)
	if err == nil {
		t.Error("expected error when semaphore is full")
	}
}

func TestSemaphore_ContextCancelled(t *testing.T) {
	s := NewSemaphore(1)
	_ = s.Acquire(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Acquire(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSemaphore_ConcurrentSafety(t *testing.T) {
	s := NewSemaphore(5)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Acquire(context.Background())
			time.Sleep(5 * time.Millisecond)
			s.Release()
		}()
	}
	wg.Wait()
	if s.Available() != 5 {
		t.Errorf("expected full availability after all goroutines done, got %d", s.Available())
	}
}
