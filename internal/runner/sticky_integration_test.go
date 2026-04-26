package runner

import (
	"context"
	"sync"
	"testing"
)

func TestStickyPolicy_ConcurrentSafety(t *testing.T) {
	pool := []string{"h1:50051", "h2:50051", "h3:50051"}
	s := NewStickyPolicy(pool)
	next := func(_ context.Context, c Call) (Response, error) {
		return Response{Body: c.Address}, nil
	}

	var wg sync.WaitGroup
	keys := []string{"k1", "k2", "k3", "k4", "k5"}
	for _, k := range keys {
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(key string) {
				defer wg.Done()
				s.Wrap(key, next)(context.Background(), Call{})
			}(k)
		}
	}
	wg.Wait()

	// Each key must be pinned to exactly one address
	for _, k := range keys {
		addr := s.Lookup(k)
		if addr == "" {
			t.Errorf("key %q has no sticky address after concurrent calls", k)
		}
	}
}

func TestStickyPolicy_EvictAndReroute(t *testing.T) {
	s := NewStickyPolicy([]string{"host-a:50051", "host-b:50051"})
	next := func(_ context.Context, c Call) (Response, error) {
		return Response{Body: c.Address}, nil
	}

	r1, _ := s.Wrap("sess", next)(context.Background(), Call{})
	first := r1.Body

	s.Evict("sess")

	// After eviction a new address is assigned (may be same or different)
	r2, _ := s.Wrap("sess", next)(context.Background(), Call{})
	if r2.Body == "" {
		t.Fatal("expected non-empty address after re-route")
	}
	_ = first // first address recorded for reference
}
