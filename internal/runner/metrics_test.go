package runner

import (
	"sync"
	"testing"
	"time"
)

func TestMetrics_RecordSuccess(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess(10 * time.Millisecond)
	if m.TotalCalls != 1 {
		t.Errorf("expected TotalCalls=1, got %d", m.TotalCalls)
	}
	if m.Succeeded != 1 {
		t.Errorf("expected Succeeded=1, got %d", m.Succeeded)
	}
	if m.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", m.Failed)
	}
}

func TestMetrics_RecordFailure(t *testing.T) {
	m := NewMetrics()
	m.RecordFailure(5 * time.Millisecond)
	if m.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", m.Failed)
	}
}

func TestMetrics_RecordRetry(t *testing.T) {
	m := NewMetrics()
	m.RecordRetry()
	m.RecordRetry()
	if m.Retried != 2 {
		t.Errorf("expected Retried=2, got %d", m.Retried)
	}
}

func TestMetrics_AvgLatency(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess(10 * time.Millisecond)
	m.RecordSuccess(20 * time.Millisecond)
	avg := m.AvgLatency()
	if avg != 15*time.Millisecond {
		t.Errorf("expected avg=15ms, got %v", avg)
	}
}

func TestMetrics_AvgLatency_NoCalls(t *testing.T) {
	m := NewMetrics()
	if m.AvgLatency() != 0 {
		t.Error("expected zero avg with no calls")
	}
}

func TestMetrics_MinMaxLatency(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess(30 * time.Millisecond)
	m.RecordSuccess(5 * time.Millisecond)
	m.RecordFailure(15 * time.Millisecond)
	if m.MinLatency != 5*time.Millisecond {
		t.Errorf("expected min=5ms, got %v", m.MinLatency)
	}
	if m.MaxLatency != 30*time.Millisecond {
		t.Errorf("expected max=30ms, got %v", m.MaxLatency)
	}
}

func TestMetrics_Snapshot_Isolation(t *testing.T) {
	m := NewMetrics()
	m.RecordSuccess(10 * time.Millisecond)
	snap := m.Snapshot()
	m.RecordSuccess(20 * time.Millisecond)
	if snap.TotalCalls != 1 {
		t.Errorf("snapshot should be isolated, got TotalCalls=%d", snap.TotalCalls)
	}
}

func TestMetrics_ConcurrentSafety(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RecordSuccess(1 * time.Millisecond)
			m.RecordRetry()
		}()
	}
	wg.Wait()
	if m.TotalCalls != 100 {
		t.Errorf("expected 100 calls, got %d", m.TotalCalls)
	}
}
