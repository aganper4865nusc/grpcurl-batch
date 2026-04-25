package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/grpcurl-batch/internal/manifest"
)

func TestProfiler_ConcurrentRecord_NoDataRace(t *testing.T) {
	p := NewProfiler(200)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Record(ProfileRecord{Duration: time.Millisecond, Success: true})
		}()
	}
	wg.Wait()
	if len(p.Snapshot()) == 0 {
		t.Fatal("expected records")
	}
}

func TestProfiler_Middleware_RecordsCallDuration(t *testing.T) {
	p := NewProfiler(10)
	mw := WithProfiler(p)

	call := func(ctx context.Context, c manifest.Call) (string, error) {
		time.Sleep(10 * time.Millisecond)
		return "ok", nil
	}

	wrapped := mw(call)
	_, err := wrapped(context.Background(), manifest.Call{ID: "test-1", Service: "svc", Method: "Ping"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := p.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 record, got %d", len(snap))
	}
	if snap[0].Duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", snap[0].Duration)
	}
	if !snap[0].Success {
		t.Error("expected success=true")
	}
}
