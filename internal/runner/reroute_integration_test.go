package runner

import (
	"context"
	"sync"
	"testing"
)

func TestReroutePolicy_ConcurrentSafety(t *testing.T) {
	p := NewReroutePolicy(nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			p.AddRule("src", "dst")
		}(i)
		go func(i int) {
			defer wg.Done()
			p.Resolve("src")
		}(i)
		go func(i int) {
			defer wg.Done()
			p.RemoveRule("src")
		}(i)
	}
	wg.Wait()
}

func TestReroutePolicy_WrapChain_AllRulesApplied(t *testing.T) {
	p1 := NewReroutePolicy([]RerouteRule{{From: "a:1", To: "b:1"}})
	p2 := NewReroutePolicy([]RerouteRule{{From: "b:1", To: "c:1"}})

	var finalAddr string
	base := func(_ context.Context, c Call) (Result, error) {
		finalAddr = c.Address
		return Result{}, nil
	}

	chained := p1.Wrap(p2.Wrap(base))
	_, err := chained(context.Background(), Call{Address: "a:1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finalAddr != "c:1" {
		t.Fatalf("expected c:1 after chained reroute, got %q", finalAddr)
	}
}
