package runner

import (
	"sync"
	"time"
)

// CacheEntry holds a cached result with an expiry timestamp.
type CacheEntry struct {
	Output    string
	CachedAt  time.Time
	ExpiresAt time.Time
}

// IsExpired returns true if the cache entry is past its TTL.
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ResponseCache is a thread-safe in-memory cache for gRPC call results.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewResponseCache creates a new ResponseCache with the given TTL.
// A zero TTL disables caching.
func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached entry by key. Returns nil if not found or expired.
func (c *ResponseCache) Get(key string) *CacheEntry {
	if c.ttl == 0 {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok || entry.IsExpired() {
		return nil
	}
	return entry
}

// Set stores a result in the cache under the given key.
func (c *ResponseCache) Set(key, output string) {
	if c.ttl == 0 {
		return
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = &CacheEntry{
		Output:    output,
		CachedAt:  now,
		ExpiresAt: now.Add(c.ttl),
	}
}

// Invalidate removes a single entry from the cache.
func (c *ResponseCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Flush removes all entries from the cache.
func (c *ResponseCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Size returns the number of entries currently in the cache (including expired).
func (c *ResponseCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
