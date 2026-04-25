package runner

import (
	"sync"
	"time"
)

// ProfileRecord captures timing and metadata for a single call execution.
type ProfileRecord struct {
	CallID    string
	Service   string
	Method    string
	StartedAt time.Time
	Duration  time.Duration
	Retries   int
	Success   bool
}

// Profiler collects per-call profiling data for later analysis.
type Profiler struct {
	mu      sync.Mutex
	records []ProfileRecord
	max     int
}

// NewProfiler creates a Profiler with the given maximum record capacity.
// If max is zero or negative it defaults to 500.
func NewProfiler(max int) *Profiler {
	if max <= 0 {
		max = 500
	}
	return &Profiler{max: max}
}

// Record appends a ProfileRecord, evicting the oldest entry when the cap is reached.
func (p *Profiler) Record(r ProfileRecord) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.records) >= p.max {
		p.records = p.records[1:]
	}
	p.records = append(p.records, r)
}

// Snapshot returns a copy of all collected records.
func (p *Profiler) Snapshot() []ProfileRecord {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]ProfileRecord, len(p.records))
	copy(out, p.records)
	return out
}

// Reset clears all profiling data.
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.records = p.records[:0]
}

// Slowest returns the n slowest records by duration.
func (p *Profiler) Slowest(n int) []ProfileRecord {
	snap := p.Snapshot()
	// simple selection sort for small n
	for i := 0; i < len(snap) && i < n; i++ {
		max := i
		for j := i + 1; j < len(snap); j++ {
			if snap[j].Duration > snap[max].Duration {
				max = j
			}
		}
		snap[i], snap[max] = snap[max], snap[i]
	}
	if n > len(snap) {
		n = len(snap)
	}
	return snap[:n]
}
