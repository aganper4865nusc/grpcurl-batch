package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/nickcorin/grpcurl-batch/internal/manifest"
)

// WithTracing wraps a CallFunc with distributed tracing via the TracingCollector.
// Each call records a span with its service, method, duration, and outcome.
// The span is stored in context so downstream middleware can read it.
func WithTracing(tc *TracingCollector) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call manifest.Call) (string, error) {
			if tc == nil {
				return next(ctx, call)
			}

			// Propagate or create a trace context.
			ctx = ensureTracingContext(ctx, call)

			start := time.Now()
			result, err := next(ctx, call)
			dur := time.Since(start)

			spanErr := ""
			if err != nil {
				spanErr = err.Error()
			}

			tc.Record(Span{
				TraceID:  traceIDFromContext(ctx),
				SpanID:   newSpanID(),
				Service:  call.Service,
				Method:   call.Method,
				Start:    start,
				Duration: dur,
				Error:    spanErr,
			})

			return result, err
		}
	}
}

// TracingStatus returns a human-readable summary of the spans recorded so far.
func TracingStatus(tc *TracingCollector) string {
	if tc == nil {
		return "tracing: disabled"
	}
	spans := tc.Spans()
	var errs int
	for _, s := range spans {
		if s.Error != "" {
			errs++
		}
	}
	return fmt.Sprintf("tracing: %d spans recorded, %d errors", len(spans), errs)
}

// ensureTracingContext attaches a fresh tracing context if none exists.
func ensureTracingContext(ctx context.Context, call manifest.Call) context.Context {
	if _, ok := ctx.Value(tracingContextKey{}).(tracingMeta); ok {
		return ctx
	}
	meta := tracingMeta{
		TraceID: fmt.Sprintf("%s.%s.%d", call.Service, call.Method, time.Now().UnixNano()),
	}
	return context.WithValue(ctx, tracingContextKey{}, meta)
}

// traceIDFromContext retrieves the trace ID embedded by ensureTracingContext.
func traceIDFromContext(ctx context.Context) string {
	if m, ok := ctx.Value(tracingContextKey{}).(tracingMeta); ok {
		return m.TraceID
	}
	return "unknown"
}

// newSpanID generates a lightweight span identifier based on the current time.
func newSpanID() string {
	return fmt.Sprintf("span-%d", time.Now().UnixNano())
}

// tracingMeta holds per-call tracing metadata stored in context.
type tracingMeta struct {
	TraceID string
}

// tracingContextKey is the unexported key used to store tracingMeta in a context.
type tracingContextKey struct{}
