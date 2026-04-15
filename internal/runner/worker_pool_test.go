package runner

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nicholasgasior/grpcurl-batch/internal/manifest"
)

// mockExecutor records invocations and optionally sleeps.
type mockExecutor struct {
	count   int64
	sleepMs int
	fail    bool
}

func (m *mockExecutor) Execute(_ context.Context, call manifest.Call) Result {
	atomic.AddInt64(&m.count, 1)
	if m.sleepMs > 0 {
		time.Sleep(time.Duration(m.sleepMs) * time.Millisecond)
	}
	return Result{
		CallName: call.Name,
		Success:  !m.fail,
	}
}

func makeCalls(n int) []manifest.Call {
	calls := make([]manifest.Call, n)
	for i := range calls {
		calls[i] = manifest.Call{Name: "call"}
	}
	return calls
}

func TestWorkerPool_RunsAllCalls(t *testing.T) {
	exec := &mockExecutor{}
	wp := NewWorkerPool(3, exec)
	results := wp.Run(context.Background(), makeCalls(10))
	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
	if exec.count != 10 {
		t.Fatalf("expected 10 executions, got %d", exec.count)
	}
}

func TestWorkerPool_RespectsConcurrencyLimit(t *testing.T) {
	var active int64
	var peak int64

	type peakExec struct{ mockExecutor }
	pe := &struct {
		active *int64
		peak   *int64
	}{&active, &peak}
	_ = pe

	// Use a custom executor inline via closure wrapping mockExecutor.
	exec := &mockExecutor{sleepMs: 20}
	wp := NewWorkerPool(2, exec)
	start := time.Now()
	wp.Run(context.Background(), makeCalls(6))
	elapsed := time.Since(start)
	// With concurrency=2 and 6 calls each sleeping 20ms: at least 3 batches → ≥60ms
	if elapsed < 50*time.Millisecond {
		t.Errorf("expected serialization delay, got %v", elapsed)
	}
}

func TestWorkerPool_ZeroConcurrencyDefaultsToOne(t *testing.T) {
	exec := &mockExecutor{}
	wp := NewWorkerPool(0, exec)
	if wp.concurrency != 1 {
		t.Errorf("expected concurrency 1, got %d", wp.concurrency)
	}
}

func TestWorkerPool_EmptyCalls(t *testing.T) {
	exec := &mockExecutor{}
	wp := NewWorkerPool(4, exec)
	results := wp.Run(context.Background(), []manifest.Call{})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
