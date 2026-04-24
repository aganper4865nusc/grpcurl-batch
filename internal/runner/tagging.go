package runner

import (
	"fmt"
	"strings"
	"sync"
)

// TagEnricher attaches derived or static tags to calls before execution.
type TagEnricher struct {
	mu      sync.RWMutex
	static  map[string]string
	derived []DerivedTag
}

// DerivedTag computes a tag value from a Call at enrichment time.
type DerivedTag struct {
	Key   string
	Value func(c Call) string
}

// NewTagEnricher creates a TagEnricher with optional static key=value pairs.
func NewTagEnricher(static map[string]string) *TagEnricher {
	if static == nil {
		static = make(map[string]string)
	}
	return &TagEnricher{static: static}
}

// AddStatic registers a static tag applied to every call.
func (te *TagEnricher) AddStatic(key, value string) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.static[key] = value
}

// AddDerived registers a tag whose value is computed from the call.
func (te *TagEnricher) AddDerived(key string, fn func(c Call) string) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.derived = append(te.derived, DerivedTag{Key: key, Value: fn})
}

// Enrich returns a copy of c with all static and derived tags merged in.
// Existing tags on the call are preserved; enricher tags do NOT override them.
func (te *TagEnricher) Enrich(c Call) Call {
	te.mu.RLock()
	defer te.mu.RUnlock()

	existing := make(map[string]struct{}, len(c.Tags))
	for _, t := range c.Tags {
		if k, _, ok := splitTag(t); ok {
			existing[k] = struct{}{}
		}
	}

	for k, v := range te.static {
		if _, found := existing[k]; !found {
			c.Tags = append(c.Tags, fmt.Sprintf("%s=%s", k, v))
			existing[k] = struct{}{}
		}
	}

	for _, d := range te.derived {
		if _, found := existing[d.Key]; !found {
			v := d.Value(c)
			if v != "" {
				c.Tags = append(c.Tags, fmt.Sprintf("%s=%s", d.Key, v))
			}
		}
	}

	return c
}

// splitTag parses "key=value" into (key, value, true).
func splitTag(tag string) (string, string, bool) {
	parts := strings.SplitN(tag, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return "", "", false
}

// WithTagEnrichment is a pipeline stage that enriches call tags.
func WithTagEnrichment(te *TagEnricher) Stage {
	return func(ctx interface{ Done() <-chan struct{} }, c Call) (Call, error) {
		return te.Enrich(c), nil
	}
}
