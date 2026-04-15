package runner

import (
	"context"
	"sync"
	"time"

	"github.com/your-org/grpcurl-batch/internal/manifest"
)

// Result holds the outcome of a single gRPC call.
type Result struct {
	CallName string
	Attempts int
	Duration time.Duration
	Output   string
	Err      error
}

// Executor defines how a single gRPC call is executed.
type Executor interface {
	Execute(ctx context.Context, call manifest.Call) (string, error)
}

// Runner orchestrates batch execution with concurrency and retry.
type Runner struct {
	executor    Executor
	concurrency int
}

// New creates a Runner with the given executor and concurrency limit.
func New(executor Executor, concurrency int) *Runner {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &Runner{executor: executor, concurrency: concurrency}
}

// Run executes all calls from the manifest concurrently and returns results.
func (r *Runner) Run(ctx context.Context, calls []manifest.Call) []Result {
	sem := make(chan struct{}, r.concurrency)
	results := make([]Result, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c manifest.Call) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = r.runWithRetry(ctx, c)
		}(i, call)
	}

	wg.Wait()
	return results
}

func (r *Runner) runWithRetry(ctx context.Context, call manifest.Call) Result {
	start := time.Now()
	maxAttempts := call.Retries + 1
	var lastErr error
	var output string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		output, lastErr = r.executor.Execute(ctx, call)
		if lastErr == nil {
			return Result{
				CallName: call.Name,
				Attempts: attempt,
				Duration: time.Since(start),
				Output:   output,
			}
		}
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return Result{CallName: call.Name, Attempts: attempt, Duration: time.Since(start), Err: ctx.Err()}
			case <-time.After(call.RetryDelay):
			}
		}
	}

	return Result{
		CallName: call.Name,
		Attempts: maxAttempts,
		Duration: time.Since(start),
		Err:      lastErr,
	}
}
