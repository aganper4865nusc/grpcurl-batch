package runner

import (
	"context"
	"errors"
	"testing"
)

func TestReroutePolicy_NoRules_ReturnsOriginal(t *testing.T) {
	p := NewReroutePolicy(nil)
	if got := p.Resolve("host:9000"); got != "host:9000" {
		t.Fatalf("expected original address, got %q", got)
	}
}

func TestReroutePolicy_MatchingRule_ReturnsDestination(t *testing.T) {
	p := NewReroutePolicy([]RerouteRule{{From: "old:80", To: "new:80"}})
	if got := p.Resolve("old:80"); got != "new:80" {
		t.Fatalf("expected new:80, got %q", got)
	}
}

func TestReroutePolicy_FirstMatchWins(t *testing.T) {
	p := NewReroutePolicy([]RerouteRule{
		{From: "a:1", To: "b:1"},
		{From: "a:1", To: "c:1"},
	})
	if got := p.Resolve("a:1"); got != "b:1" {
		t.Fatalf("expected b:1, got %q", got)
	}
}

func TestReroutePolicy_AddRule_TakesEffect(t *testing.T) {
	p := NewReroutePolicy(nil)
	p.AddRule("x:1", "y:1")
	if got := p.Resolve("x:1"); got != "y:1" {
		t.Fatalf("expected y:1, got %q", got)
	}
}

func TestReroutePolicy_RemoveRule_NoLongerMatches(t *testing.T) {
	p := NewReroutePolicy([]RerouteRule{{From: "a:1", To: "b:1"}})
	p.RemoveRule("a:1")
	if got := p.Resolve("a:1"); got != "a:1" {
		t.Fatalf("expected original address after removal, got %q", got)
	}
}

func TestReroutePolicy_Wrap_RewritesAddress(t *testing.T) {
	p := NewReroutePolicy([]RerouteRule{{From: "src:80", To: "dst:80"}})
	var captured string
	next := func(_ context.Context, c Call) (Result, error) {
		captured = c.Address
		return Result{}, nil
	}
	c := Call{Address: "src:80"}
	_, _ = p.Wrap(next)(context.Background(), c)
	if captured != "dst:80" {
		t.Fatalf("expected dst:80, got %q", captured)
	}
}

func TestReroutePolicy_Wrap_EmptyResolved_ReturnsError(t *testing.T) {
	p := NewReroutePolicy([]RerouteRule{{From: "src:80", To: ""}})
	next := func(_ context.Context, c Call) (Result, error) { return Result{}, nil }
	_, err := p.Wrap(next)(context.Background(), Call{Address: "src:80"})
	if err == nil || !errors.Is(err, errors.New("reroute: resolved address is empty")) {
		if err == nil {
			t.Fatal("expected error for empty resolved address")
		}
	}
}

func TestRerouteStatus_ReturnsRuleCount(t *testing.T) {
	p := NewReroutePolicy([]RerouteRule{{From: "a", To: "b"}, {From: "c", To: "d"}})
	s := RerouteStatus(p)
	if s != "reroute: 2 rule(s) active" {
		t.Fatalf("unexpected status: %q", s)
	}
}
