package runner

import (
	"sync"
	"time"
)

// WindowPolicy tracks call outcomes within a sliding time window.
type WindowPolicy struct {
	mu       sync.Mutex
	duration time.Duration
	entries  []windowEntry
}

type windowEntry struct {
	at      time.Time
	success bool
}

// WindowStats holds aggregated stats for a window.
type WindowStats struct {
	Total    int
	Success  int
	Failures int
}

// NewWindowPolicy creates a WindowPolicy with the given duration.
func NewWindowPolicy(d time.Duration) *WindowPolicy {
	if d <= 0 {
		d = 10 * time.Second
	}
	return &WindowPolicy{duration: d}
}

// Record adds an outcome to the window.
func (w *WindowPolicy) Record(success bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	w.entries = append(w.entries, windowEntry{at: time.Now(), success: success})
}

// Stats returns aggregated stats for the current window.
func (w *WindowPolicy) Stats() WindowStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	var s WindowStats
	for _, e := range w.entries {
		s.Total++
		if e.success {
			s.Success++
		} else {
			s.Failures++
		}
	}
	return s
}

// Reset clears all entries.
func (w *WindowPolicy) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = nil
}

func (w *WindowPolicy) evict() {
	cutoff := time.Now().Add(-w.duration)
	i := 0
	for i < len(w.entries) && w.entries[i].at.Before(cutoff) {
		i++
	}
	w.entries = w.entries[i:]
}
