package runner

import (
	"context"
	"sort"
	"sync"
)

// Priority levels for calls.
const (
	PriorityLow    = 0
	PriorityNormal = 5
	PriorityHigh   = 10
)

type priorityCall struct {
	call     Call
	priority int
}

// PriorityQueue orders calls by priority before execution.
type PriorityQueue struct {
	mu    sync.Mutex
	items []priorityCall
}

// NewPriorityQueue creates an empty PriorityQueue.
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

// Enqueue adds a call with the given priority.
func (pq *PriorityQueue) Enqueue(c Call, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items = append(pq.items, priorityCall{call: c, priority: priority})
}

// Drain returns all calls sorted by descending priority.
func (pq *PriorityQueue) Drain() []Call {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	sort.SliceStable(pq.items, func(i, j int) bool {
		return pq.items[i].priority > pq.items[j].priority
	})
	out := make([]Call, len(pq.items))
	for i, item := range pq.items {
		out[i] = item.call
	}
	pq.items = nil
	return out
}

// Len returns the number of queued calls.
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.items)
}

// PriorityPolicy wraps a call executor, draining a PriorityQueue before running.
type PriorityPolicy struct {
	queue *PriorityQueue
}

// NewPriorityPolicy creates a PriorityPolicy backed by the given queue.
func NewPriorityPolicy(q *PriorityQueue) *PriorityPolicy {
	return &PriorityPolicy{queue: q}
}

// RunAll drains the queue and executes each call via fn in priority order.
func (p *PriorityPolicy) RunAll(ctx context.Context, fn func(context.Context, Call) (string, error)) ([]string, error) {
	calls := p.queue.Drain()
	results := make([]string, 0, len(calls))
	for _, c := range calls {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		res, err := fn(ctx, c)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}
