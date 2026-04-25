package runner

import (
	"context"
	"sync"
	"time"
)

// EventKind categorises a runner lifecycle event.
type EventKind string

const (
	EventCallStarted   EventKind = "call.started"
	EventCallSucceeded EventKind = "call.succeeded"
	EventCallFailed    EventKind = "call.failed"
	EventCallRetried   EventKind = "call.retried"
	EventCallSkipped   EventKind = "call.skipped"
)

// Event holds a single structured log entry.
type Event struct {
	Kind      EventKind         `json:"kind"`
	Timestamp time.Time         `json:"timestamp"`
	CallID    string            `json:"call_id"`
	Method    string            `json:"method"`
	Attempt   int               `json:"attempt,omitempty"`
	Err       string            `json:"error,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// EventLog is a bounded, thread-safe in-memory log of runner events.
type EventLog struct {
	mu     sync.Mutex
	events []Event
	maxLen int
}

// NewEventLog creates an EventLog that retains at most maxLen events.
// If maxLen <= 0 it defaults to 500.
func NewEventLog(maxLen int) *EventLog {
	if maxLen <= 0 {
		maxLen = 500
	}
	return &EventLog{maxLen: maxLen}
}

// Record appends an event, evicting the oldest entry when the log is full.
func (el *EventLog) Record(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	el.mu.Lock()
	defer el.mu.Unlock()
	if len(el.events) >= el.maxLen {
		el.events = el.events[1:]
	}
	el.events = append(el.events, e)
}

// All returns a snapshot of all recorded events.
func (el *EventLog) All() []Event {
	el.mu.Lock()
	defer el.mu.Unlock()
	out := make([]Event, len(el.events))
	copy(out, el.events)
	return out
}

// Filter returns events matching the given kind.
func (el *EventLog) Filter(kind EventKind) []Event {
	el.mu.Lock()
	defer el.mu.Unlock()
	var out []Event
	for _, e := range el.events {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// Reset clears all events.
func (el *EventLog) Reset() {
	el.mu.Lock()
	defer el.mu.Unlock()
	el.events = el.events[:0]
}

// WithEventLog stores an EventLog in the context.
func WithEventLog(ctx context.Context, el *EventLog) context.Context {
	return context.WithValue(ctx, eventLogKey{}, el)
}

// EventLogFromContext retrieves the EventLog from ctx, or nil.
func EventLogFromContext(ctx context.Context) *EventLog {
	v, _ := ctx.Value(eventLogKey{}).(*EventLog)
	return v
}

type eventLogKey struct{}
