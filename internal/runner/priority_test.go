package runner

import (
	"context"
	"testing"
)

func TestPriorityQueue_DrainSortedByPriority(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Enqueue(Call{Method: "low"}, PriorityLow)
	pq.Enqueue(Call{Method: "high"}, PriorityHigh)
	pq.Enqueue(Call{Method: "normal"}, PriorityNormal)

	calls := pq.Drain()
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
	if calls[0].Method != "high" {
		t.Errorf("expected first call to be high, got %s", calls[0].Method)
	}
	if calls[1].Method != "normal" {
		t.Errorf("expected second call to be normal, got %s", calls[1].Method)
	}
	if calls[2].Method != "low" {
		t.Errorf("expected third call to be low, got %s", calls[2].Method)
	}
}

func TestPriorityQueue_DrainEmptiesQueue(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Enqueue(Call{Method: "a"}, PriorityNormal)
	pq.Drain()
	if pq.Len() != 0 {
		t.Errorf("expected empty queue after drain")
	}
}

func TestPriorityQueue_Len(t *testing.T) {
	pq := NewPriorityQueue()
	if pq.Len() != 0 {
		t.Errorf("expected 0")
	}
	pq.Enqueue(Call{}, PriorityLow)
	pq.Enqueue(Call{}, PriorityHigh)
	if pq.Len() != 2 {
		t.Errorf("expected 2, got %d", pq.Len())
	}
}

func TestPriorityPolicy_RunAll_OrderPreserved(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Enqueue(Call{Method: "b"}, PriorityLow)
	pq.Enqueue(Call{Method: "a"}, PriorityHigh)

	policy := NewPriorityPolicy(pq)
	var order []string
	_, err := policy.RunAll(context.Background(), func(_ context.Context, c Call) (string, error) {
		order = append(order, c.Method)
		return c.Method, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestPriorityPolicy_RunAll_ContextCancelled(t *testing.T) {
	pq := NewPriorityQueue()
	pq.Enqueue(Call{Method: "x"}, PriorityNormal)
	pq.Enqueue(Call{Method: "y"}, PriorityNormal)

	policy := NewPriorityPolicy(pq)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := policy.RunAll(ctx, func(_ context.Context, c Call) (string, error) {
		return "", nil
	})
	if err == nil {
		t.Error("expected context cancellation error")
	}
}
