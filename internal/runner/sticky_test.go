package runner

import (
	"context"
	"testing"
)

func TestStickyPolicy_EmptyPool_ReturnsError(t *testing.T) {
	s := NewStickyPolicy(nil)
	wrapped := s.Wrap("key1", func(_ context.Context, c Call) (Response, error) {
		return Response{}, nil
	})
	_, err := wrapped(context.Background(), Call{Address: "original"})
	if err == nil {
		t.Fatal("expected error for empty pool")
	}
}

func TestStickyPolicy_SameKey_SameAddress(t *testing.T) {
	s := NewStickyPolicy([]string{"host-a:50051", "host-b:50051"})
	var seen []string
	next := func(_ context.Context, c Call) (Response, error) {
		seen = append(seen, c.Address)
		return Response{}, nil
	}
	for i := 0; i < 4; i++ {
		s.Wrap("session-1", next)(context.Background(), Call{})
	}
	for i := 1; i < len(seen); i++ {
		if seen[i] != seen[0] {
			t.Fatalf("expected sticky address %q, got %q", seen[0], seen[i])
		}
	}
}

func TestStickyPolicy_DifferentKeys_MayDiffer(t *testing.T) {
	s := NewStickyPolicy([]string{"host-a:50051", "host-b:50051"})
	var addrA, addrB string
	next := func(_ context.Context, c Call) (Response, error) {
		return Response{Body: c.Address}, nil
	}
	rA, _ := s.Wrap("key-a", next)(context.Background(), Call{})
	rB, _ := s.Wrap("key-b", next)(context.Background(), Call{})
	addrA, addrB = rA.Body, rB.Body
	if addrA == addrB {
		t.Log("both keys routed to same address (pool may cycle)")
	}
	// Ensure each key stays pinned
	rA2, _ := s.Wrap("key-a", next)(context.Background(), Call{})
	if rA2.Body != addrA {
		t.Fatalf("key-a changed address from %q to %q", addrA, rA2.Body)
	}
}

func TestStickyPolicy_Assign_PinsAddress(t *testing.T) {
	s := NewStickyPolicy([]string{"host-a:50051"})
	s.Assign("user-42", "host-z:50051")
	var got string
	next := func(_ context.Context, c Call) (Response, error) {
		got = c.Address
		return Response{}, nil
	}
	s.Wrap("user-42", next)(context.Background(), Call{})
	if got != "host-z:50051" {
		t.Fatalf("expected host-z:50051, got %q", got)
	}
}

func TestStickyPolicy_Evict_RemovesBinding(t *testing.T) {
	s := NewStickyPolicy([]string{"host-a:50051"})
	s.Assign("user-7", "host-z:50051")
	s.Evict("user-7")
	if addr := s.Lookup("user-7"); addr != "" {
		t.Fatalf("expected empty after evict, got %q", addr)
	}
}

func TestStickyPolicy_Lookup_MissingKey_ReturnsEmpty(t *testing.T) {
	s := NewStickyPolicy([]string{"host-a:50051"})
	if addr := s.Lookup("nonexistent"); addr != "" {
		t.Fatalf("expected empty string, got %q", addr)
	}
}
