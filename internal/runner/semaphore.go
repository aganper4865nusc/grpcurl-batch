package runner

import "context"

// Semaphore limits concurrent access to a resource.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a Semaphore with the given capacity.
// If capacity <= 0, it defaults to 1.
func NewSemaphore(capacity int) *Semaphore {
	if capacity <= 0 {
		capacity = 1
	}
	return &Semaphore{ch: make(chan struct{}, capacity)}
}

// Acquire blocks until a slot is available or ctx is cancelled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a slot.
func (s *Semaphore) Release() {
	<-s.ch
}

// Available returns the number of free slots.
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// Capacity returns the total capacity.
func (s *Semaphore) Capacity() int {
	return cap(s.ch)
}
