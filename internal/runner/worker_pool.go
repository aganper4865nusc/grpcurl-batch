package runner

import (
	"context"
	"sync"

	"github.com/nicholasgasior/grpcurl-batch/internal/manifest"
)

// WorkerPool manages concurrent execution of gRPC calls.
type WorkerPool struct {
	concurrency int
	executor    Executor
}

// Executor is the interface for executing a single gRPC call.
type Executor interface {
	Execute(ctx context.Context, call manifest.Call) Result
}

// NewWorkerPool creates a WorkerPool with the given concurrency limit.
func NewWorkerPool(concurrency int, executor Executor) *WorkerPool {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &WorkerPool{
		concurrency: concurrency,
		executor:    executor,
	}
}

// Run dispatches all calls from the manifest concurrently and collects results.
func (wp *WorkerPool) Run(ctx context.Context, calls []manifest.Call) []Result {
	results := make([]Result, len(calls))
	sem := make(chan struct{}, wp.concurrency)
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, c manifest.Call) {
			defer wg.Done()
			defer func() { <-sem }()
			results[idx] = wp.executor.Execute(ctx, c)
		}(i, call)
	}

	wg.Wait()
	return results
}
