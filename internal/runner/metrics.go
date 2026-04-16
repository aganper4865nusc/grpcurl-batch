package runner

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks runtime statistics for a batch run.
type Metrics struct {
	mu           sync.RWMutex
	TotalCalls   int64
	Succeeded    int64
	Failed       int64
	Retried      int64
	TotalLatency time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
}

// NewMetrics creates an initialized Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordSuccess records a successful call with its latency.
func (m *Metrics) RecordSuccess(latency time.Duration) {
	atomic.AddInt64(&m.TotalCalls, 1)
	atomic.AddInt64(&m.Succeeded, 1)
	m.recordLatency(latency)
}

// RecordFailure records a failed call with its latency.
func (m *Metrics) RecordFailure(latency time.Duration) {
	atomic.AddInt64(&m.TotalCalls, 1)
	atomic.AddInt64(&m.Failed, 1)
	m.recordLatency(latency)
}

// RecordRetry increments the retry counter.
func (m *Metrics) RecordRetry() {
	atomic.AddInt64(&m.Retried, 1)
}

func (m *Metrics) recordLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalLatency += latency
	if m.MinLatency == 0 || latency < m.MinLatency {
		m.MinLatency = latency
	}
	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}
}

// AvgLatency returns the average latency across all calls.
func (m *Metrics) AvgLatency() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total := atomic.LoadInt64(&m.TotalCalls)
	if total == 0 {
		return 0
	}
	return m.TotalLatency / time.Duration(total)
}

// Snapshot returns a copy of the current metrics.
func (m *Metrics) Snapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Metrics{
		TotalCalls:   atomic.LoadInt64(&m.TotalCalls),
		Succeeded:    atomic.LoadInt64(&m.Succeeded),
		Failed:       atomic.LoadInt64(&m.Failed),
		Retried:      atomic.LoadInt64(&m.Retried),
		TotalLatency: m.TotalLatency,
		MinLatency:   m.MinLatency,
		MaxLatency:   m.MaxLatency,
	}
}
