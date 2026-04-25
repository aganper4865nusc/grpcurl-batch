package runner

import (
	"testing"
	"time"
)

func TestProfiler_RecordAndSnapshot(t *testing.T) {
	p := NewProfiler(10)
	p.Record(ProfileRecord{CallID: "a", Duration: 5 * time.Millisecond, Success: true})
	p.Record(ProfileRecord{CallID: "b", Duration: 10 * time.Millisecond, Success: false})

	snap := p.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 records, got %d", len(snap))
	}
}

func TestProfiler_ZeroMax_DefaultsTo500(t *testing.T) {
	p := NewProfiler(0)
	if p.max != 500 {
		t.Fatalf("expected max 500, got %d", p.max)
	}
}

func TestProfiler_EvictsOldest(t *testing.T) {
	p := NewProfiler(2)
	p.Record(ProfileRecord{CallID: "first"})
	p.Record(ProfileRecord{CallID: "second"})
	p.Record(ProfileRecord{CallID: "third"})

	snap := p.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2, got %d", len(snap))
	}
	if snap[0].CallID != "second" {
		t.Errorf("expected 'second', got %q", snap[0].CallID)
	}
}

func TestProfiler_Reset_ClearsRecords(t *testing.T) {
	p := NewProfiler(10)
	p.Record(ProfileRecord{CallID: "x"})
	p.Reset()
	if len(p.Snapshot()) != 0 {
		t.Fatal("expected empty after reset")
	}
}

func TestProfiler_Slowest_ReturnsSorted(t *testing.T) {
	p := NewProfiler(10)
	p.Record(ProfileRecord{CallID: "fast", Duration: 1 * time.Millisecond})
	p.Record(ProfileRecord{CallID: "slow", Duration: 100 * time.Millisecond})
	p.Record(ProfileRecord{CallID: "mid", Duration: 50 * time.Millisecond})

	top := p.Slowest(2)
	if len(top) != 2 {
		t.Fatalf("expected 2, got %d", len(top))
	}
	if top[0].CallID != "slow" {
		t.Errorf("expected 'slow' first, got %q", top[0].CallID)
	}
	if top[1].CallID != "mid" {
		t.Errorf("expected 'mid' second, got %q", top[1].CallID)
	}
}

func TestProfiler_Slowest_NLargerThanRecords(t *testing.T) {
	p := NewProfiler(10)
	p.Record(ProfileRecord{CallID: "only", Duration: 5 * time.Millisecond})
	top := p.Slowest(100)
	if len(top) != 1 {
		t.Fatalf("expected 1, got %d", len(top))
	}
}
