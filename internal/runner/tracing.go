package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Span represents a single traced call.
type Span struct {
	CallID  string
	Method  string
	Start   time.Time
	End     time.Time
	Success bool
	Err     error
	Tags    map[string]string
}

func (s Span) Duration() time.Duration {
	return s.End.Sub(s.Start)
}

// TracingCollector records spans for all calls.
type TracingCollector struct {
	mu    sync.Mutex
	spans []Span
}

func NewTracingCollector() *TracingCollector {
	return &TracingCollector{}
}

func (tc *TracingCollector) Record(span Span) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.spans = append(tc.spans, span)
}

func (tc *TracingCollector) Spans() []Span {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	out := make([]Span, len(tc.spans))
	copy(out, tc.spans)
	return out
}

func (tc *TracingCollector) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.spans = nil
}

type tracingKey struct{}

func ContextWithTracing(ctx context.Context, tc *TracingCollector) context.Context {
	return context.WithValue(ctx, tracingKey{}, tc)
}

func TracingFromContext(ctx context.Context) (*TracingCollector, bool) {
	tc, ok := ctx.Value(tracingKey{}).(*TracingCollector)
	return tc, ok
}

// WithTracing middleware wraps a call and records a span.
func WithTracing(tc *TracingCollector) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			start := time.Now()
			res, err := next(ctx, call)
			span := Span{
				CallID:  fmt.Sprintf("%s/%s", call.Service, call.Method),
				Method:  call.Method,
				Start:   start,
				End:     time.Now(),
				Success: err == nil,
				Err:     err,
				Tags:    map[string]string{"service": call.Service},
			}
			tc.Record(span)
			return res, err
		}
	}
}
