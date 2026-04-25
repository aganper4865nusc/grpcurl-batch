package runner

import (
	"context"
	"fmt"
)

// WithEventLog middleware records call lifecycle events into the EventLog
// stored in the context (if present).  It is a no-op when no EventLog is
// attached to ctx.
func WithEventLogMiddleware(next CallFunc) CallFunc {
	return func(ctx context.Context, c Call) (string, error) {
		el := EventLogFromContext(ctx)

		if el != nil {
			el.Record(Event{
				Kind:   EventCallStarted,
				CallID: c.ID,
				Method: c.Method,
			})
		}

		resp, err := next(ctx, c)

		if el == nil {
			return resp, err
		}

		if err != nil {
			el.Record(Event{
				Kind:   EventCallFailed,
				CallID: c.ID,
				Method: c.Method,
				Err:    err.Error(),
			})
		} else {
			el.Record(Event{
				Kind:   EventCallSucceeded,
				CallID: c.ID,
				Method: c.Method,
			})
		}
		return resp, err
	}
}

// EventLogSummary returns a short human-readable summary of event counts.
func EventLogSummary(el *EventLog) string {
	if el == nil {
		return "no event log"
	}
	started := len(el.Filter(EventCallStarted))
	succeeded := len(el.Filter(EventCallSucceeded))
	failed := len(el.Filter(EventCallFailed))
	retried := len(el.Filter(EventCallRetried))
	skipped := len(el.Filter(EventCallSkipped))
	return fmt.Sprintf(
		"started=%d succeeded=%d failed=%d retried=%d skipped=%d",
		started, succeeded, failed, retried, skipped,
	)
}
