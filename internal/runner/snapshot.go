package runner

import (
	"sync"
	"time"
)

// Snapshot captures a point-in-time view of runner metrics and state.
type Snapshot struct {
	CapturedAt     time.Time
	TotalCalls     int64
	Succeeded      int64
	Failed         int64
	Retries        int64
	AvgLatencyMs   float64
	CircuitOpen    bool
	CacheSize      int
	ActiveWorkers  int
}

// SnapshotCollector periodically captures runner state snapshots.
type SnapshotCollector struct {
	mu        sync.RWMutex
	snapshots []Snapshot
	max       int
}

// NewSnapshotCollector creates a collector that retains up to max snapshots.
func NewSnapshotCollector(max int) *SnapshotCollector {
	if max <= 0 {
		max = 100
	}
	return &SnapshotCollector{max: max}
}

// Record appends a new snapshot, evicting the oldest if at capacity.
func (sc *SnapshotCollector) Record(s Snapshot) {
	s.CapturedAt = time.Now()
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if len(sc.snapshots) >= sc.max {
		sc.snapshots = sc.snapshots[1:]
	}
	sc.snapshots = append(sc.snapshots, s)
}

// Latest returns the most recent snapshot, and false if none exist.
func (sc *SnapshotCollector) Latest() (Snapshot, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if len(sc.snapshots) == 0 {
		return Snapshot{}, false
	}
	return sc.snapshots[len(sc.snapshots)-1], true
}

// All returns a copy of all retained snapshots.
func (sc *SnapshotCollector) All() []Snapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	out := make([]Snapshot, len(sc.snapshots))
	copy(out, sc.snapshots)
	return out
}

// Reset clears all stored snapshots.
func (sc *SnapshotCollector) Reset() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.snapshots = sc.snapshots[:0]
}
