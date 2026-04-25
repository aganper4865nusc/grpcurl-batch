package runner

import (
	"context"
	"fmt"
)

// WithScatter returns a middleware that fans out every call to all addresses
// in the pool and returns the first successful result. If all addresses fail
// the combined error is returned.
func WithScatter(pool []string, exec func(context.Context, Call) (*Result, error)) Middleware {
	sp, err := NewScatterPolicy(pool)
	if err != nil {
		// Return a middleware that always errors if the pool is invalid.
		return func(next Handler) Handler {
			return func(ctx context.Context, call Call) (*Result, error) {
				return nil, fmt.Errorf("scatter middleware: %w", err)
			}
		}
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, call Call) (*Result, error) {
			results := sp.Scatter(ctx, call, exec)
			sr, ferr := FirstSuccess(results)
			if ferr != nil {
				return nil, ferr
			}
			return sr.Result, nil
		}
	}
}

// ScatterStatus summarises a scatter fan-out for observability.
type ScatterStatus struct {
	Total     int
	Succeeded int
	Failed    int
}

// SummariseScatter builds a ScatterStatus from a slice of ScatterResults.
func SummariseScatter(results []ScatterResult) ScatterStatus {
	s := ScatterStatus{Total: len(results)}
	for _, r := range results {
		if r.Err == nil {
			s.Succeeded++
		} else {
			s.Failed++
		}
	}
	return s
}
