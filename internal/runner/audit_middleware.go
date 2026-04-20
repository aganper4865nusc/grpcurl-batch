package runner

import (
	"context"
	"time"
)

// WithAudit wraps a CallFunc with audit logging, recording every call
// attempt, its outcome, latency, and any error into the provided AuditLog.
func WithAudit(log *AuditLog) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			start := time.Now()
			result, err := next(ctx, call)
			latency := time.Since(start)

			event := AuditEvent{
				Timestamp: start,
				Service:   call.Service,
				Method:    call.Method,
				Address:   call.Address,
				LatencyMs: latency.Milliseconds(),
				Success:   err == nil,
			}
			if err != nil {
				event.Error = err.Error()
			}
			if result != "" {
				event.Response = result
			}

			log.Record(event)
			return result, err
		}
	}
}

// WithAuditFilter wraps a CallFunc with audit logging but only records events
// that match the provided predicate. Useful for recording only failures.
func WithAuditFilter(log *AuditLog, predicate func(AuditEvent) bool) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			start := time.Now()
			result, err := next(ctx, call)
			latency := time.Since(start)

			event := AuditEvent{
				Timestamp: start,
				Service:   call.Service,
				Method:    call.Method,
				Address:   call.Address,
				LatencyMs: latency.Milliseconds(),
				Success:   err == nil,
			}
			if err != nil {
				event.Error = err.Error()
			}

			if predicate(event) {
				log.Record(event)
			}
			return result, err
		}
	}
}

// AuditFailuresOnly returnsicate that matches only failed events.
func AuditFailuresOnly() func(AuditEvent) bool {
	return func(e AuditEvent) bool {
		return !e.Success
	}
}
