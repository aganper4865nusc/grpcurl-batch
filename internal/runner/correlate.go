package runner

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// correlationKeyType is an unexported type for context keys to avoid collisions.
type correlationKeyType struct{}

var correlationKey = correlationKeyType{}

// CorrelationIDFunc generates a correlation ID for a call.
type CorrelationIDFunc func(service, method string) string

// CorrelationStore tracks active correlation IDs and their associated metadata.
type CorrelationStore struct {
	mu      sync.RWMutex
	records map[string]correlationRecord
	genID   CorrelationIDFunc
}

type correlationRecord struct {
	ID        string
	Service   string
	Method    string
	CreatedAt time.Time
}

// DefaultCorrelationIDFunc generates a random hex correlation ID prefixed with
// the service short name.
func DefaultCorrelationIDFunc(service, method string) string {
	return fmt.Sprintf("%s-%08x", method, rand.Uint32())
}

// NewCorrelationStore creates a CorrelationStore with the given ID generator.
// If gen is nil, DefaultCorrelationIDFunc is used.
func NewCorrelationStore(gen CorrelationIDFunc) *CorrelationStore {
	if gen == nil {
		gen = DefaultCorrelationIDFunc
	}
	return &CorrelationStore{
		records: make(map[string]correlationRecord),
		genID:   gen,
	}
}

// Assign generates and stores a correlation ID for the given call, returning
// a context enriched with the ID.
func (c *CorrelationStore) Assign(ctx context.Context, service, method string) (context.Context, string) {
	id := c.genID(service, method)
	c.mu.Lock()
	c.records[id] = correlationRecord{
		ID:        id,
		Service:   service,
		Method:    method,
		CreatedAt: time.Now(),
	}
	c.mu.Unlock()
	return context.WithValue(ctx, correlationKey, id), id
}

// FromContext retrieves the correlation ID from a context, returning empty
// string if none is set.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(correlationKey).(string)
	return v
}

// Remove deletes the correlation record for the given ID.
func (c *CorrelationStore) Remove(id string) {
	c.mu.Lock()
	delete(c.records, id)
	c.mu.Unlock()
}

// Snapshot returns a copy of all current correlation records.
func (c *CorrelationStore) Snapshot() []correlationRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]correlationRecord, 0, len(c.records))
	for _, r := range c.records {
		out = append(out, r)
	}
	return out
}

// WithCorrelation is a middleware that assigns a correlation ID to each call
// and removes it after the call completes.
func WithCorrelation(store *CorrelationStore) func(CallFunc) CallFunc {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			ctx, id := store.Assign(ctx, call.Service, call.Method)
			defer store.Remove(id)
			return next(ctx, call)
		}
	}
}
