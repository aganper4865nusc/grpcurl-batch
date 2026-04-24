package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRoundRobinBalancer_ConcurrentPick_NoDataRace(t *testing.T) {
	lb, err := NewRoundRobinBalancer([]string{"h1:1", "h2:2", "h3:3"})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	const goroutines = 50
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = lb.Pick()
			}
		}()
	}
	wg.Wait()
}

func TestWithLoadBalancer_DistributesAcrossPool(t *testing.T) {
	addrs := []string{"a:1", "b:2", "c:3"}
	lb, _ := NewRoundRobinBalancer(addrs)
	mw := WithLoadBalancer(lb)

	hits := make(map[string]*int64, len(addrs))
	for _, a := range addrs {
		v := new(int64)
		hits[a] = v
	}

	inner := func(_ context.Context, c Call) (Result, error) {
		if v, ok := hits[c.Address]; ok {
			atomic.AddInt64(v, 1)
		}
		return Result{Call: c}, nil
	}

	chain := mw(inner)
	const total = 300
	for i := 0; i < total; i++ {
		_, _ = chain(context.Background(), Call{Method: "svc/M"})
	}

	expected := int64(total / len(addrs))
	for addr, v := range hits {
		if atomic.LoadInt64(v) != expected {
			t.Errorf("addr %s: want %d hits, got %d", addr, expected, atomic.LoadInt64(v))
		}
	}
}
