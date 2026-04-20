package runner

import (
	"testing"
	"time"
)

func TestWindowPolicy_EmptyStats(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	s := w.Stats()
	if s.Total != 0 {
		t.Fatalf("expected 0 total, got %d", s.Total)
	}
}

func TestWindowPolicy_RecordAndStats(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	w.Record(true)
	w.Record(true)
	w.Record(false)
	s := w.Stats()
	if s.Total != 3 || s.Success != 2 || s.Failures != 1 {
		t.Fatalf("unexpected stats: %+v", s)
	}
}

func TestWindowPolicy_EvictsExpired(t *testing.T) {
	w := NewWindowPolicy(50 * time.Millisecond)
	w.Record(true)
	w.Record(false)
	time.Sleep(80 * time.Millisecond)
	w.Record(true)
	s := w.Stats()
	if s.Total != 1 || s.Success != 1 {
		t.Fatalf("expected only 1 entry after eviction, got %+v", s)
	}
}

func TestWindowPolicy_Reset(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	w.Record(true)
	w.Record(false)
	w.Reset()
	s := w.Stats()
	if s.Total != 0 {
		t.Fatalf("expected 0 after reset, got %d", s.Total)
	}
}

func TestWindowPolicy_ZeroDuration_DefaultsTo10s(t *testing.T) {
	w := NewWindowPolicy(0)
	if w.duration != 10*time.Second {
		t.Fatalf("expected default 10s, got %v", w.duration)
	}
}
