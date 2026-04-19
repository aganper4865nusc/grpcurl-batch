package runner

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestCheckpointStore_ConcurrentWrites_NoDataRace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cp.json")
	cs, err := NewCheckpointStore(path)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := string(rune('a' + n))
			_ = cs.Record(id, n%2 == 0)
			_ = cs.IsDone(id)
		}(i)
	}
	wg.Wait()
}

func TestCheckpointStore_ResumeSkipsDoneItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cp.json")
	cs, _ := NewCheckpointStore(path)

	allCalls := []string{"a", "b", "c", "d"}
	// Pre-complete a and c
	_ = cs.Record("a", true)
	_ = cs.Record("c", true)

	var pending []string
	for _, id := range allCalls {
		if !cs.IsDone(id) {
			pending = append(pending, id)
		}
	}

	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got %d: %v", len(pending), pending)
	}
	for _, p := range pending {
		if p == "a" || p == "c" {
			t.Errorf("completed call %q should not be pending", p)
		}
	}
}
