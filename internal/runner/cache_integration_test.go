package runner

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestResponseCache_TTLPrecision(t *testing.T) {
	ttl := 50 * time.Millisecond
	c := NewResponseCache(ttl)
	c.Set("precise", "data")

	// Should still be valid just before expiry
	time.Sleep(30 * time.Millisecond)
	if got := c.Get("precise"); got == nil {
		t.Error("expected cache hit before TTL expiry")
	}

	// Should be expired after TTL
	time.Sleep(30 * time.Millisecond)
	if got := c.Get("precise"); got != nil {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestResponseCache_HighConcurrency_NoDataRace(t *testing.T) {
	c := NewResponseCache(200 * time.Millisecond)
	var wg sync.WaitGroup
	workers := 50
	iterations := 100

	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("key-%d-%d", w, i%10)
				c.Set(key, fmt.Sprintf("value-%d", i))
				_ = c.Get(key)
			}
		}(w)
	}
	wg.Wait()
}

func TestResponseCache_FlushResetsSize(t *testing.T) {
	c := NewResponseCache(10 * time.Second)
	for i := 0; i < 100; i++ {
		c.Set(fmt.Sprintf("key-%d", i), "val")
	}
	if c.Size() != 100 {
		t.Fatalf("expected 100 entries, got %d", c.Size())
	}
	c.Flush()
	if c.Size() != 0 {
		t.Errorf("expected 0 entries after flush, got %d", c.Size())
	}
	// Verify cache is still usable after flush
	c.Set("after-flush", "data")
	if c.Get("after-flush") == nil {
		t.Error("expected cache to be usable after flush")
	}
}
