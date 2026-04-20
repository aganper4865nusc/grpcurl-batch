package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTracingCollector_RecordAndSpans(t *testing.T) {
	tc := NewTracingCollector()
	span := Span{CallID: "svc/Method", Method: "Method", Start: time.Now(), End: time.Now(), Success: true}
	tc.Record(span)
	spans := tc.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].CallID != "svc/Method" {
		t.Errorf("unexpected CallID: %s", spans[0].CallID)
	}
}

func TestTracingCollector_Reset(t *testing.T) {
	tc := NewTracingCollector()
	tc.Record(Span{CallID: "a"})
	tc.Reset()
	if len(tc.Spans()) != 0 {
		t.Fatal("expected empty spans after reset")
	}
}

func TestContextWithTracing_RoundTrip(t *testing.T) {
	tc := NewTracingCollector()
	ctx := ContextWithTracing(context.Background(), tc)
	got, ok := TracingFromContext(ctx)
	if !ok || got != tc {
		t.Fatal("expected to retrieve same collector from context")
	}
}

func TestTracingFromContext_Missing(t *testing.T) {
	_, ok := TracingFromContext(context.Background())
	if ok {
		t.Fatal("expected no collector in empty context")
	}
}

func TestWithTracing_RecordsSuccess(t *testing.T) {
	tc := NewTracingCollector()
	mw := WithTracing(tc)
	call := Call{Service: "svc", Method: "Say"}
	_, err := mw(func(_ context.Context, _ Call) (string, error) {
		return "ok", nil
	})(context.Background(), call)
	if err != nil {
		t.Fatal(err)
	}
	spans := tc.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Success {
		t.Error("expected success span")
	}
}

func TestWithTracing_RecordsFailure(t *testing.T) {
	tc := NewTracingCollector()
	mw := WithTracing(tc)
	call := Call{Service: "svc", Method: "Fail"}
	wantErr := errors.New("boom")
	_, _ = mw(func(_ context.Context, _ Call) (string, error) {
		return "", wantErr
	})(context.Background(), call)
	spans := tc.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}
	if spans[0].Success {
		t.Error("expected failure span")
	}
	if !errors.Is(spans[0].Err, wantErr) {
		t.Errorf("unexpected err: %v", spans[0].Err)
	}
}

func TestSpan_Duration(t *testing.T) {
	start := time.Now()
	end := start.Add(50 * time.Millisecond)
	s := Span{Start: start, End: end}
	if s.Duration() != 50*time.Millisecond {
		t.Errorf("unexpected duration: %v", s.Duration())
	}
}
