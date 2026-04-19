package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestShedder_AllowsCallsBelowMax(t *testing.T) {
	s := NewShedderPolicy(5)
	err := s.Do(context.Background(), func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestShedder_DropsCallsAboveMax(t *testing.T) {
	s := NewShedderPolicy(1)
	blocked := make(chan struct{})
	go func() {
		_ = s.Do(context.Background(), func(_ context.Context) error {
			<-blocked
			return nil
		})
	}()
	time.Sleep(10 * time.Millisecond)
	err := s.Do(context.Background(), func(_ context.Context) error { return nil })
	close(blocked)
	if err != ErrLoadShed {
		t.Fatalf("expected ErrLoadShed, got %v", err)
	}
}

func TestShedder_ZeroMax_DefaultsTo100(t *testing.T) {
	s := NewShedderPolicy(0)
	if s.max != 100 {
		t.Fatalf("expected max=100, got %d", s.max)
	}
}

func TestShedder_InflightDecrementsAfterDone(t *testing.T) {
	s := NewShedderPolicy(10)
	_ = s.Do(context.Background(), func(_ context.Context) error { return nil })
	if s.Inflight() != 0 {
		t.Fatalf("expected 0 inflight, got %d", s.Inflight())
	}
}

func TestShedder_CancelledContext_ReturnsCtxErr(t *testing.T) {
	s := NewShedderPolicy(10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.Do(ctx, func(_ context.Context) error { return nil })
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestShedder_ConcurrentSafety(t *testing.T) {
	s := NewShedderPolicy(50)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Do(context.Background(), func(_ context.Context) error {
				time.Sleep(time.Millisecond)
				return nil
			})
		}()
	}
	wg.Wait()
	if s.Inflight() != 0 {
		t.Fatalf("expected 0 inflight after all done, got %d", s.Inflight())
	}
}
