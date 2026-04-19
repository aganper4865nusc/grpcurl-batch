package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestQuotaPolicy_ExactlyMaxCallsSucceed(t *testing.T) {
	const max = 20
	q := NewQuotaPolicy(max)
	var succeeded, failed int64
	var wg sync.WaitGroup
	for i := 0; i < max*2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := q.Wrap(context.Background(), func(_ context.Context) error {
				atomic.AddInt64(&succeeded, 1)
				return nil
			})
			if err == ErrQuotaExceeded {
				atomic.AddInt64(&failed, 1)
			}
		}()
	}
	wg.Wait()
	if succeeded != max {
		t.Errorf("expected exactly %d successes, got %d", max, succeeded)
	}
	if succeeded+failed != max*2 {
		t.Errorf("expected %d total outcomes, got %d", max*2, succeeded+failed)
	}
}

func TestQuotaPolicy_ResetAllowsNewBatch(t *testing.T) {
	q := NewQuotaPolicy(5)
	for i := 0; i < 5; i++ {
		_ = q.Allow()
	}
	if q.Allow() != ErrQuotaExceeded {
		t.Fatal("expected quota exceeded before reset")
	}
	q.Reset()
	var wg sync.WaitGroup
	var ok int64
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if q.Allow() == nil {
				atomic.AddInt64(&ok, 1)
			}
		}()
	}
	wg.Wait()
	if ok != 5 {
		t.Errorf("expected 5 allowed after reset, got %d", ok)
	}
}
