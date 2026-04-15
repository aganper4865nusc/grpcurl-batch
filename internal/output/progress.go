package output

import (
	"fmt"
	"io"
	"sync"
)

// ProgressTracker tracks and prints progress of batch calls.
type ProgressTracker struct {
	mu      sync.Mutex
	out     io.Writer
	total   int
	done    int
	silent  bool
}

// NewProgressTracker creates a ProgressTracker for total calls.
func NewProgressTracker(out io.Writer, total int, silent bool) *ProgressTracker {
	return &ProgressTracker{
		out:    out,
		total:  total,
		silent: silent,
	}
}

// Increment records one completed call and prints current progress.
func (p *ProgressTracker) Increment(callName string, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++
	if p.silent {
		return
	}
	status := "✓"
	if !success {
		status = "✗"
	}
	fmt.Fprintf(p.out, "[%d/%d] %s %s\n", p.done, p.total, status, callName)
}

// Done returns the number of completed calls.
func (p *ProgressTracker) Done() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.done
}
