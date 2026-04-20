package runner

import (
	"context"
	"sync"
	"time"
)

// AuditEvent captures a single call execution event for auditing purposes.
type AuditEvent struct {
	CallID    string
	Service   string
	Method    string
	Address   string
	Status    string // "success", "failure", "skipped"
	Error     string
	Attempts  int
	Latency   time.Duration
	Timestamp time.Time
}

// AuditLog collects AuditEvents in memory and supports flushing.
type AuditLog struct {
	mu     sync.Mutex
	events []AuditEvent
	max    int
}

// NewAuditLog creates an AuditLog with the given max capacity.
// If max is zero or negative, it defaults to 1000.
func NewAuditLog(max int) *AuditLog {
	if max <= 0 {
		max = 1000
	}
	return &AuditLog{max: max}
}

// Record appends an AuditEvent, evicting the oldest if at capacity.
func (a *AuditLog) Record(evt AuditEvent) {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.events) >= a.max {
		a.events = a.events[1:]
	}
	a.events = append(a.events, evt)
}

// Events returns a snapshot of all recorded events.
func (a *AuditLog) Events() []AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditEvent, len(a.events))
	copy(out, a.events)
	return out
}

// Flush returns all events and resets the log.
func (a *AuditLog) Flush() []AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := a.events
	a.events = nil
	return out
}

// Len returns the current number of recorded events.
func (a *AuditLog) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.events)
}

// WithAudit wraps a CallFunc to record an AuditEvent for every execution.
func WithAudit(log *AuditLog, next CallFunc) CallFunc {
	return func(ctx context.Context, call Call) (string, error) {
		start := time.Now()
		resp, err := next(ctx, call)
		latency := time.Since(start)

		evt := AuditEvent{
			CallID:    call.ID,
			Service:   call.Service,
			Method:    call.Method,
			Address:   call.Address,
			Latency:   latency,
			Timestamp: start,
			Attempts:  1,
		}
		if err != nil {
			evt.Status = "failure"
			evt.Error = err.Error()
		} else {
			evt.Status = "success"
			_ = resp
		}
		log.Record(evt)
		return resp, err
	}
}
