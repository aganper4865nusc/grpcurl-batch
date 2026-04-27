package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// RerouteRule maps a source address to a replacement address.
type RerouteRule struct {
	From string
	To   string
}

// ReroutePolicy rewrites the Address field of a Call based on a set of rules.
// Rules are matched in order; the first match wins.
type ReroutePolicy struct {
	mu    sync.RWMutex
	rules []RerouteRule
}

// NewReroutePolicy creates a ReroutePolicy with the given initial rules.
func NewReroutePolicy(rules []RerouteRule) *ReroutePolicy {
	copy := make([]RerouteRule, len(rules))
	for i, r := range rules {
		copy[i] = r
	}
	return &ReroutePolicy{rules: copy}
}

// AddRule appends a new reroute rule at runtime.
func (p *ReroutePolicy) AddRule(from, to string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rules = append(p.rules, RerouteRule{From: from, To: to})
}

// RemoveRule removes all rules where From matches addr.
func (p *ReroutePolicy) RemoveRule(from string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	filtered := p.rules[:0]
	for _, r := range p.rules {
		if r.From != from {
			filtered = append(filtered, r)
		}
	}
	p.rules = filtered
}

// Resolve returns the destination address for src, or src unchanged if no rule matches.
func (p *ReroutePolicy) Resolve(src string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, r := range p.rules {
		if r.From == src {
			return r.To
		}
	}
	return src
}

// Wrap applies reroute rules before delegating to next.
func (p *ReroutePolicy) Wrap(next CallFunc) CallFunc {
	return func(ctx context.Context, c Call) (Result, error) {
		dest := p.Resolve(c.Address)
		if dest == "" {
			return Result{}, errors.New("reroute: resolved address is empty")
		}
		c.Address = dest
		return next(ctx, c)
	}
}

// RerouteStatus returns a human-readable summary of active rules.
func RerouteStatus(p *ReroutePolicy) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("reroute: %d rule(s) active", len(p.rules))
}
