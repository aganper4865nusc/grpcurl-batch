package runner

import (
	"fmt"
	"sync"
	"testing"
)

func TestDedupeFilter_NewCall_NotDuplicate(t *testing.T) {
	d := NewDedupeFilter(nil)
	call := Call{Service: "svc", Method: "Foo", Address: "localhost:50051"}
	if d.IsDuplicate(call) {
		t.Error("expected first call to not be a duplicate")
	}
}

func TestDedupeFilter_SameCall_IsDuplicate(t *testing.T) {
	d := NewDedupeFilter(nil)
	call := Call{Service: "svc", Method: "Foo", Address: "localhost:50051"}
	d.IsDuplicate(call)
	if !d.IsDuplicate(call) {
		t.Error("expected second call to be a duplicate")
	}
}

func TestDedupeFilter_DifferentCalls_NotDuplicate(t *testing.T) {
	d := NewDedupeFilter(nil)
	a := Call{Service: "svc", Method: "Foo", Address: "localhost:50051"}
	b := Call{Service: "svc", Method: "Bar", Address: "localhost:50051"}
	d.IsDuplicate(a)
	if d.IsDuplicate(b) {
		t.Error("expected different call to not be a duplicate")
	}
}

func TestDedupeFilter_Reset_ClearsSeen(t *testing.T) {
	d := NewDedupeFilter(nil)
	call := Call{Service: "svc", Method: "Foo", Address: "localhost:50051"}
	d.IsDuplicate(call)
	d.Reset()
	if d.IsDuplicate(call) {
		t.Error("expected call to not be duplicate after reset")
	}
}

func TestDedupeFilter_SeenCount(t *testing.T) {
	d := NewDedupeFilter(nil)
	for i := 0; i < 5; i++ {
		d.IsDuplicate(Call{Service: "svc", Method: fmt.Sprintf("M%d", i), Address: "addr"})
	}
	if d.SeenCount() != 5 {
		t.Errorf("expected 5 seen, got %d", d.SeenCount())
	}
}

func TestDedupeFilter_CustomKeyFn(t *testing.T) {
	keyFn := func(c Call) string { return c.Method }
	d := NewDedupeFilter(keyFn)
	a := Call{Service: "svc1", Method: "Foo", Address: "addr1"}
	b := Call{Service: "svc2", Method: "Foo", Address: "addr2"}
	d.IsDuplicate(a)
	if !d.IsDuplicate(b) {
		t.Error("expected same method to be duplicate with custom key")
	}
}

func TestDedupeFilter_ConcurrentSafety(t *testing.T) {
	d := NewDedupeFilter(nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			call := Call{Service: "svc", Method: fmt.Sprintf("M%d", i%10), Address: "addr"}
			d.IsDuplicate(call)
		}(i)
	}
	wg.Wait()
}
