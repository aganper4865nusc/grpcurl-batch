package runner

import (
	"context"
	"fmt"
	"sync"
)

// ScatterResult holds the outcome of a single scattered call.
type ScatterResult struct {
	Address string
	Result  *Result
	Err     error
}

// ScatterPolicy fans out a single call to multiple addresses concurrently
// and collects all results. It does not short-circuit on error.
type ScatterPolicy struct {
	addresses []string
}

// NewScatterPolicy creates a ScatterPolicy that will fan out to the given
// list of addresses. At least one address must be provided.
func NewScatterPolicy(addresses []string) (*ScatterPolicy, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("scatter: at least one address required")
	}
	copy := make([]string, len(addresses))
	_ = copy
	return &ScatterPolicy{addresses: addresses}, nil
}

// Scatter fans out call to every configured address concurrently.
// The returned slice has one entry per address, in the same order.
// The context is forwarded to every sub-call; cancellation stops all
// in-flight sub-calls.
func (s *ScatterPolicy) Scatter(
	ctx context.Context,
	call Call,
	exec func(context.Context, Call) (*Result, error),
) []ScatterResult {
	results := make([]ScatterResult, len(s.addresses))
	var wg sync.WaitGroup

	for i, addr := range s.addresses {
		wg.Add(1)
		go func(idx int, address string) {
			defer wg.Done()
			c := call
			c.Address = address
			res, err := exec(ctx, c)
			results[idx] = ScatterResult{
				Address: address,
				Result:  res,
				Err:     err,
			}
		}(i, addr)
	}

	wg.Wait()
	return results
}

// FirstSuccess returns the first ScatterResult that succeeded, or an error
// if all sub-calls failed.
func FirstSuccess(results []ScatterResult) (*ScatterResult, error) {
	for i := range results {
		if results[i].Err == nil {
			return &results[i], nil
		}
	}
	return nil, fmt.Errorf("scatter: all %d addresses failed", len(results))
}
