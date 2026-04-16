package runner

import (
	"sync"
	"testing"
	"time"
)

func TestMetrics_HighConcurrency_NoDataRace(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup
	workers := 50
	ops := 20
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				if id%2 == 0 {
					m.RecordSuccess(time.Duration(j) * time.Millisecond)
				} else {
					m.RecordFailure(time.Duration(j) * time.Millisecond)
				}
				m.RecordRetry()
				_ = m.Snapshot()
				_ = m.AvgLatency()
			}
		}(i)
	}
	wg.Wait()
	expected := int64(workers * ops)
	if m.TotalCalls != expected {
		t.Errorf("expected %d total calls, got %d", expected, m.TotalCalls)
	}
	if m.Retried != expected {
		t.Errorf("expected %d retries, got %d", expected, m.Retried)
	}
}

func TestMetrics_SucceededPlusFailedEqualsTotalCalls(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%3 == 0 {
				m.RecordFailure(1 * time.Millisecond)
			} else {
				m.RecordSuccess(1 * time.Millisecond)
			}
		}(i)
	}
	wg.Wait()
	snap := m.Snapshot()
	if snap.Succeeded+snap.Failed != snap.TotalCalls {
		t.Errorf("succeeded(%d)+failed(%d) != total(%d)", snap.Succeeded, snap.Failed, snap.TotalCalls)
	}
}
