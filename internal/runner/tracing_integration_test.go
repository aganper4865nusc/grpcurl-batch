package runner

import (
	"context"
	"sync"
	"testing"
)

func TestTracing_ConcurrentRecord_NoDataRace(t *testing.T) {
	tc := NewTracingCollector()
	mw := WithTracing(tc)
	base := func(_ context.Context, _ Call) (string, error) { return "ok", nil }
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mw(base)(context.Background(), Call{Service: "s", Method: "m"})
		}()
	}
	wg.Wait()
	if len(tc.Spans()) != 50 {
		t.Errorf("expected 50 spans, got %d", len(tc.Spans()))
	}
}

func TestTracing_SpansPreserveOrder(t *testing.T) {
	tc := NewTracingCollector()
	mw := WithTracing(tc)
	base := func(_ context.Context, _ Call) (string, error) { return "ok", nil }
	for i := 0; i < 5; i++ {
		mw(base)(context.Background(), Call{Service: "svc", Method: "m"})
	}
	spans := tc.Spans()
	for i := 1; i < len(spans); i++ {
		if spans[i].Start.Before(spans[i-1].Start) {
			t.Errorf("spans out of order at index %d", i)
		}
	}
}
