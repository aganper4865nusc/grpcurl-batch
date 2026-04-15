package runner

import (
	"context"
	"sync"
	"time"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// Runner executes batch gRPC calls with concurrency and retry controls.
type Runner struct {
	executor    Executor
	concurrency int
}

// New creates a Runner with the given executor and concurrency limit.
func New(exec Executor, concurrency int) *Runner {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &Runner{executor: exec, concurrency: concurrency}
}

// Run executes all calls in the manifest and returns a Summary.
func (r *Runner) Run(ctx context.Context, m *manifest.Manifest) Summary {
	start := time.Now()
	sem := make(chan struct{}, r.concurrency)
	resultsCh := make(chan CallResult, len(m.Calls))

	var wg sync.WaitGroup
	for _, call := range m.Calls {
		wg.Add(1)
		sem <- struct{}{}
		go func(c manifest.Call) {
			defer wg.Done()
			defer func() { <-sem }()
			resultsCh <- r.runWithRetry(ctx, c)
		}(call)
	}

	wg.Wait()
	close(resultsCh)

	var results []CallResult
	for res := range resultsCh {
		results = append(results, res)
	}
	return Summarize(results, time.Since(start))
}

// runWithRetry attempts a call up to maxRetries+1 times.
func (r *Runner) runWithRetry(ctx context.Context, call manifest.Call) CallResult {
	result := CallResult{
		CallName: call.Name,
		Address:  call.Address,
		Method:   call.Method,
	}

	maxAttempts := call.Retries + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		start := time.Now()
		out, err := r.executor.Execute(ctx, call)
		result.Attempts = attempt
		result.Duration += time.Since(start)

		if err == nil {
			result.Success = true
			result.Output = out
			return result
		}
		result.Error = err.Error()

		if attempt < maxAttempts && call.RetryDelay > 0 {
			select {
			case <-ctx.Done():
				return result
			case <-time.After(call.RetryDelay):
			}
		}
	}
	return result
}
