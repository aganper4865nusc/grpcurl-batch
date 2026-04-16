package runner

import (
	"testing"
	"time"
)

func TestResponseCache_GetMiss_ReturnsNil(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	if got := c.Get("missing"); got != nil {
		t.Errorf("expected nil for cache miss, got %v", got)
	}
}

func TestResponseCache_SetAndGet(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	c.Set("key1", `{"result":"ok"}`)
	entry := c.Get("key1")
	if entry == nil {
		t.Fatal("expected cache hit, got nil")
	}
	if entry.Output != `{"result":"ok"}` {
		t.Errorf("unexpected output: %s", entry.Output)
	}
}

func TestResponseCache_Expired_ReturnsNil(t *testing.T) {
	c := NewResponseCache(10 * time.Millisecond)
	c.Set("key2", "data")
	time.Sleep(20 * time.Millisecond)
	if got := c.Get("key2"); got != nil {
		t.Error("expected nil for expired entry, got a result")
	}
}

func TestResponseCache_ZeroTTL_DisablesCache(t *testing.T) {
	c := NewResponseCache(0)
	c.Set("key3", "data")
	if got := c.Get("key3"); got != nil {
		t.Error("expected nil when TTL is zero (cache disabled)")
	}
}

func TestResponseCache_Invalidate(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	c.Set("key4", "value")
	c.Invalidate("key4")
	if got := c.Get("key4"); got != nil {
		t.Error("expected nil after invalidation")
	}
}

func TestResponseCache_Flush(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	c.Set("a", "1")
	c.Set("b", "2")
	c.Flush()
	if c.Size() != 0 {
		t.Errorf("expected size 0 after flush, got %d", c.Size())
	}
}

func TestResponseCache_Size(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	if c.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", c.Size())
	}
	c.Set("x", "data")
	c.Set("y", "data")
	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}
}

func TestResponseCache_ConcurrentSafety(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func(i int) {
			key := string(rune('a' + i%26))
			c.Set(key, "value")
			_ = c.Get(key)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}
