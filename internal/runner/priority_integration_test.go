package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestPriorityQueue_ConcurrentEnqueue_NoDataRace(t *testing.T) {
	pq := NewPriorityQueue()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			pq.Enqueue(Call{Method: "m"}, p%3)
		}(i)
	}
	wg.Wait()
	if pq.Len() != 50 {
		t.Errorf("expected 50 items, got %d", pq.Len())
	}
}

func TestPriorityPolicy_RunAll_AllCallsExecuted(t *testing.T) {
	const total = 20
	pq := NewPriorityQueue()
	for i := 0; i < total; i++ {
		pq.Enqueue(Call{Method: "call"}, i%PriorityHigh)
	}

	policy := NewPriorityPolicy(pq)
	var count int64
	_, err := policy.RunAll(context.Background(), func(_ context.Context, _ Call) (string, error) {
		atomic.AddInt64(&count, 1)
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != total {
		t.Errorf("expected %d executions, got %d", total, count)
	}
}
