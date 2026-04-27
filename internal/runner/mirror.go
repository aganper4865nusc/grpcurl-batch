package runner

import (
	"context"
	"fmt"
	"sync"
)

// MirrorPolicy duplicates every call to one or more mirror addresses,
// discards mirror results, and returns only the primary result.
type MirrorPolicy struct {
	mu      sync.RWMutex
	mirrors []string
	exec    func(ctx context.Context, call Call) (string, error)
}

// NewMirrorPolicy creates a MirrorPolicy that fans out calls to the
// provided mirror addresses. exec is used to dispatch mirror calls.
func NewMirrorPolicy(mirrors []string, exec func(ctx context.Context, call Call) (string, error)) *MirrorPolicy {
	clone := make([]string, len(mirrors))
	copy(clone, mirrors)
	return &MirrorPolicy{mirrors: clone, exec: exec}
}

// AddMirror appends an address to the mirror pool.
func (m *MirrorPolicy) AddMirror(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mirrors = append(m.mirrors, addr)
}

// RemoveMirror removes the first occurrence of addr from the mirror pool.
func (m *MirrorPolicy) RemoveMirror(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, a := range m.mirrors {
		if a == addr {
			m.mirrors = append(m.mirrors[:i], m.mirrors[i+1:]...)
			return
		}
	}
}

// Mirrors returns a snapshot of the current mirror addresses.
func (m *MirrorPolicy) Mirrors() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.mirrors))
	copy(out, m.mirrors)
	return out
}

// Wrap executes call against the primary address, fires mirror copies
// asynchronously, and returns the primary outcome.
func (m *MirrorPolicy) Wrap(next CallFunc) CallFunc {
	return func(ctx context.Context, call Call) (string, error) {
		m.mu.RLock()
		mirrors := make([]string, len(m.mirrors))
		copy(mirrors, m.mirrors)
		m.mu.RUnlock()

		for _, addr := range mirrors {
			mc := call
			mc.Address = addr
			go func(c Call) { //nolint:errcheck
				_, _ = m.exec(ctx, c)
			}(mc)
		}

		res, err := next(ctx, call)
		if err != nil {
			return "", fmt.Errorf("mirror primary: %w", err)
		}
		return res, nil
	}
}
