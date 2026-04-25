package runner

import (
	"context"
	"time"

	"github.com/yourorg/grpcurl-batch/internal/manifest"
)

// WithProfiler wraps a CallFunc so that every execution is recorded in p.
// retries is the number of retry attempts already consumed before this call.
func WithProfiler(p *Profiler) func(CallFunc) CallFunc {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, c manifest.Call) (string, error) {
			start := time.Now()
			result, err := next(ctx, c)
			rec := ProfileRecord{
				CallID:    c.ID,
				Service:   c.Service,
				Method:    c.Method,
				StartedAt: start,
				Duration:  time.Since(start),
				Success:   err == nil,
			}
			_ = result
			p.Record(rec)
			return result, err
		}
	}
}

// ProfilerStatus returns a snapshot of the profiler's current records.
func ProfilerStatus(p *Profiler) []ProfileRecord {
	return p.Snapshot()
}
