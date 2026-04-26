package runner

import (
	"context"
	"errors"
	"sync"
)

// StickyPolicy routes repeated calls with the same key to the same address,
// providing session affinity across a pool of backends.
type StickyPolicy struct {
	mu      sync.RWMutex
	routes  map[string]string
	pool    []string
	counter uint64
}

// NewStickyPolicy creates a StickyPolicy backed by the given address pool.
// If pool is empty, Wrap returns an error on every call.
func NewStickyPolicy(pool []string) *StickyPolicy {
	copy := make([]string, len(pool))
	copy(copy, pool)
	return &StickyPolicy{
		routes: make(map[string]string),
		pool:   copy,
	}
}

// Assign pins key to a specific address. Subsequent calls with the same key
// will be routed to that address.
func (s *StickyPolicy) Assign(key, address string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[key] = address
}

// Lookup returns the pinned address for key, or an empty string if none.
func (s *StickyPolicy) Lookup(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routes[key]
}

// Evict removes the sticky binding for key.
func (s *StickyPolicy) Evict(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.routes, key)
}

// pick selects an address from the pool using round-robin, then pins it to key.
func (s *StickyPolicy) pick(key string) (string, error) {
	if len(s.pool) == 0 {
		return "", errors.New("sticky: address pool is empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if addr, ok := s.routes[key]; ok {
		return addr, nil
	}
	addr := s.pool[s.counter%uint64(len(s.pool))]
	s.counter++
	s.routes[key] = addr
	return addr, nil
}

// Wrap returns a CallFunc that rewrites the call's address based on sticky
// routing for the given key, then delegates to next.
func (s *StickyPolicy) Wrap(key string, next CallFunc) CallFunc {
	return func(ctx context.Context, c Call) (Response, error) {
		addr, err := s.pick(key)
		if err != nil {
			return Response{}, err
		}
		c.Address = addr
		return next(ctx, c)
	}
}
