package runner

import (
	"testing"
	"time"
)

func TestSnapshotCollector_RecordAndLatest(t *testing.T) {
	sc := NewSnapshotCollector(10)
	sc.Record(Snapshot{TotalCalls: 5, Succeeded: 4, Failed: 1})

	s, ok := sc.Latest()
	if !ok {
		t.Fatal("expected snapshot, got none")
	}
	if s.TotalCalls != 5 {
		t.Errorf("expected TotalCalls=5, got %d", s.TotalCalls)
	}
	if s.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestSnapshotCollector_Latest_Empty(t *testing.T) {
	sc := NewSnapshotCollector(10)
	_, ok := sc.Latest()
	if ok {
		t.Error("expected no snapshot on empty collector")
	}
}

func TestSnapshotCollector_EvictsOldest(t *testing.T) {
	sc := NewSnapshotCollector(3)
	for i := int64(1); i <= 4; i++ {
		sc.Record(Snapshot{TotalCalls: i})
	}
	all := sc.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(all))
	}
	if all[0].TotalCalls != 2 {
		t.Errorf("expected oldest evicted, first TotalCalls=2, got %d", all[0].TotalCalls)
	}
}

func TestSnapshotCollector_Reset(t *testing.T) {
	sc := NewSnapshotCollector(10)
	sc.Record(Snapshot{TotalCalls: 1})
	sc.Reset()
	if all := sc.All(); len(all) != 0 {
		t.Errorf("expected empty after reset, got %d", len(all))
	}
}

func TestSnapshotCollector_ZeroMax_DefaultsTo100(t *testing.T) {
	sc := NewSnapshotCollector(0)
	if sc.max != 100 {
		t.Errorf("expected max=100, got %d", sc.max)
	}
}

func TestSnapshotCollector_ConcurrentSafety(t *testing.T) {
	sc := NewSnapshotCollector(50)
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func(n int) {
			sc.Record(Snapshot{TotalCalls: int64(n)})
			sc.Latest()
			sc.All()
			done <- struct{}{}
		}(i)
	}
	timeout := time.After(2 * time.Second)
	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timeout waiting for goroutines")
		}
	}
}
