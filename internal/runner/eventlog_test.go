package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEventLog_RecordAndAll(t *testing.T) {
	el := NewEventLog(10)
	el.Record(Event{Kind: EventCallStarted, CallID: "c1", Method: "Foo"})
	el.Record(Event{Kind: EventCallSucceeded, CallID: "c1", Method: "Foo"})

	events := el.All()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Kind != EventCallStarted {
		t.Errorf("expected first event to be %s", EventCallStarted)
	}
}

func TestEventLog_EvictsOldest(t *testing.T) {
	el := NewEventLog(3)
	for i := 0; i < 5; i++ {
		el.Record(Event{Kind: EventCallStarted, CallID: "x"})
	}
	if got := len(el.All()); got != 3 {
		t.Fatalf("expected 3 events after eviction, got %d", got)
	}
}

func TestEventLog_Filter(t *testing.T) {
	el := NewEventLog(20)
	el.Record(Event{Kind: EventCallStarted})
	el.Record(Event{Kind: EventCallFailed})
	el.Record(Event{Kind: EventCallFailed})
	el.Record(Event{Kind: EventCallSucceeded})

	failed := el.Filter(EventCallFailed)
	if len(failed) != 2 {
		t.Fatalf("expected 2 failed events, got %d", len(failed))
	}
}

func TestEventLog_Reset(t *testing.T) {
	el := NewEventLog(10)
	el.Record(Event{Kind: EventCallStarted})
	el.Reset()
	if got := len(el.All()); got != 0 {
		t.Fatalf("expected 0 events after reset, got %d", got)
	}
}

func TestEventLog_TimestampAutoSet(t *testing.T) {
	before := time.Now()
	el := NewEventLog(10)
	el.Record(Event{Kind: EventCallStarted})
	after := time.Now()

	events := el.All()
	ts := events[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not within [%v, %v]", ts, before, after)
	}
}

func TestEventLog_ZeroMax_DefaultsTo500(t *testing.T) {
	el := NewEventLog(0)
	if el.maxLen != 500 {
		t.Errorf("expected maxLen 500, got %d", el.maxLen)
	}
}

func TestEventLog_ContextRoundTrip(t *testing.T) {
	el := NewEventLog(10)
	ctx := WithEventLog(context.Background(), el)
	got := EventLogFromContext(ctx)
	if got != el {
		t.Error("EventLog not retrieved from context")
	}
}

func TestEventLog_ContextMissing_ReturnsNil(t *testing.T) {
	got := EventLogFromContext(context.Background())
	if got != nil {
		t.Error("expected nil for context without EventLog")
	}
}

func TestEventLog_ConcurrentSafety(t *testing.T) {
	el := NewEventLog(200)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			el.Record(Event{Kind: EventCallStarted})
			_ = el.All()
			_ = el.Filter(EventCallStarted)
		}()
	}
	wg.Wait()
}
