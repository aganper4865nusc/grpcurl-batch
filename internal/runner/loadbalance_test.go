package runner

import (
	"context"
	"testing"
)

func TestRoundRobinBalancer_EmptyPool_Error(t *testing.T) {
	_, err := NewRoundRobinBalancer(nil)
	if err == nil {
		t.Fatal("expected error for empty pool")
	}
}

func TestRoundRobinBalancer_SingleAddr_AlwaysSame(t *testing.T) {
	lb, err := NewRoundRobinBalancer([]string{"host-a:443"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 5; i++ {
		addr, err := lb.Pick()
		if err != nil {
			t.Fatalf("pick error: %v", err)
		}
		if addr != "host-a:443" {
			t.Errorf("expected host-a:443, got %s", addr)
		}
	}
}

func TestRoundRobinBalancer_MultiAddr_Cycles(t *testing.T) {
	addrs := []string{"a:1", "b:2", "c:3"}
	lb, _ := NewRoundRobinBalancer(addrs)

	for round := 0; round < 2; round++ {
		for i, want := range addrs {
			got, err := lb.Pick()
			if err != nil {
				t.Fatalf("round %d pick %d: %v", round, i, err)
			}
			if got != want {
				t.Errorf("round %d pick %d: want %s got %s", round, i, want, got)
			}
		}
	}
}

func TestWithLoadBalancer_RewritesAddress(t *testing.T) {
	lb, _ := NewRoundRobinBalancer([]string{"lb-host:443"})
	mw := WithLoadBalancer(lb)

	var captured string
	inner := func(_ context.Context, c Call) (Result, error) {
		captured = c.Address
		return Result{Call: c}, nil
	}

	call := Call{Address: "original:9000", Method: "pkg.Svc/Method"}
	_, err := mw(inner)(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != "lb-host:443" {
		t.Errorf("expected lb-host:443, got %s", captured)
	}
}

func TestWithLoadBalancer_OriginalAddressUnchangedOnError(t *testing.T) {
	// empty pool triggers Pick error
	lb := &RoundRobinBalancer{addrs: []string{}} // bypass constructor
	mw := WithLoadBalancer(lb)

	called := false
	inner := func(_ context.Context, c Call) (Result, error) {
		called = true
		return Result{Call: c}, nil
	}

	call := Call{Address: "original:9000", Method: "pkg.Svc/Method"}
	_, err := mw(inner)(context.Background(), call)
	if err == nil {
		t.Fatal("expected error from empty balancer")
	}
	if called {
		t.Error("inner should not be called when Pick fails")
	}
}
