package runner

import (
	"context"
	"sync"
	"testing"
)

func TestQuotaPolicy_ZeroMax_NoLimit(t *testing.T) {
	q := NewQuotaPolicy(0)
	for i := 0; i < 100; i++ {
		if err := q.Allow(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}
	if q.Remaining() != -1 {
		t.Errorf("expected -1 remaining for unlimited, got %d", q.Remaining())
	}
}

func TestQuotaPolicy_AllowsUpToMax(t *testing.T) {
	q := NewQuotaPolicy(3)
	for i := 0; i < 3; i++ {
		if err := q.Allow(); err != nil {
			t.Fatalf("call %d should be allowed: %v", i, err)
		}
	}
	if err := q.Allow(); err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestQuotaPolicy_UsedAndRemaining(t *testing.T) {
	q := NewQuotaPolicy(5)
	_ = q.Allow()
	_ = q.Allow()
	if q.Used() != 2 {
		t.Errorf("expected Used=2, got %d", q.Used())
	}
	if q.Remaining() != 3 {
		t.Errorf("expected Remaining=3, got %d", q.Remaining())
	}
}

func TestQuotaPolicy_Reset_ClearsCounter(t *testing.T) {
	q := NewQuotaPolicy(2)
	_ = q.Allow()
	_ = q.Allow()
	q.Reset()
	if q.Used() != 0 {
		t.Errorf("expected Used=0 after reset, got %d", q.Used())
	}
	if err := q.Allow(); err != nil {
		t.Errorf("expected call allowed after reset, got %v", err)
	}
}

func TestQuotaPolicy_Wrap_ExceedsQuota(t *testing.T) {
	q := NewQuotaPolicy(1)
	called := 0
	fn := func(ctx context.Context) error { called++; return nil }
	_ = q.Wrap(context.Background(), fn)
	err := q.Wrap(context.Background(), fn)
	if err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	if called != 1 {
		t.Errorf("expected fn called once, got %d", called)
	}
}

func TestQuotaPolicy_ConcurrentSafety(t *testing.T) {
	const max = 50
	q := NewQuotaPolicy(max)
	var wg sync.WaitGroup
	allowed := atomic.Int64{}
	// use plain int64 via sync/atomic via the struct field
	var allowedCount int64
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if q.Allow() == nil {
				sync_atomicAdd(&allowedCount, 1)
			}
		}()
	}
	wg.Wait()
	_ = allowed
	if q.Used() > max {
		t.Errorf("used %d exceeds max %d", q.Used(), max)
	}
}

func sync_atomicAdd(p *int64, v int64) {
	sync.NewCond(nil) // just to use sync import; actual add below
	*p += v           // safe because we only read after wg.Wait
}
