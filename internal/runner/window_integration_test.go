package runner

import (
	"sync"
	"testing"
	"time"
)

func TestWindowPolicy_ConcurrentRecord_NoDataRace(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w.Record(i%2 == 0)
		}(i)
	}
	wg.Wait()
	s := w.Stats()
	if s.Total != 50 {
		t.Fatalf("expected 50 entries, got %d", s.Total)
	}
}

func TestWindowPolicy_SlidingEviction(t *testing.T) {
	w := NewWindowPolicy(60 * time.Millisecond)
	for i := 0; i < 5; i++ {
		w.Record(false)
	}
	time.Sleep(80 * time.Millisecond)
	for i := 0; i < 3; i++ {
		w.Record(true)
	}
	s := w.Stats()
	if s.Total != 3 {
		t.Fatalf("expected 3 entries after sliding eviction, got %d", s.Total)
	}
	if s.Failures != 0 {
		t.Fatalf("expected 0 failures after eviction, got %d", s.Failures)
	}
}
