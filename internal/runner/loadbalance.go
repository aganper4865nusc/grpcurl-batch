package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// LoadBalanceStrategy selects a target address from a pool.
type LoadBalanceStrategy interface {
	Pick() (string, error)
	Report(addr string, success bool)
}

// RoundRobinBalancer cycles through addresses in order.
type RoundRobinBalancer struct {
	addrs   []string
	counter uint64
}

// NewRoundRobinBalancer creates a balancer over the given address pool.
// Returns an error if the pool is empty.
func NewRoundRobinBalancer(addrs []string) (*RoundRobinBalancer, error) {
	if len(addrs) == 0 {
		return nil, errors.New("loadbalance: address pool must not be empty")
	}
	pool := make([]string, len(addrs))
	copy(pool, addrs)
	return &RoundRobinBalancer{addrs: pool}, nil
}

// Pick returns the next address in round-robin order.
func (r *RoundRobinBalancer) Pick() (string, error) {
	n := atomic.AddUint64(&r.counter, 1) - 1
	return r.addrs[n%uint64(len(r.addrs))], nil
}

// Report is a no-op for round-robin; success/failure does not affect selection.
func (r *RoundRobinBalancer) Report(_ string, _ bool) {}

// WithLoadBalancer is a middleware that rewrites Call.Address using the balancer.
func WithLoadBalancer(lb LoadBalanceStrategy) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (Result, error) {
			addr, err := lb.Pick()
			if err != nil {
				return Result{Call: call, Err: err}, err
			}
			original := call.Address
			if addr != "" {
				call.Address = addr
			}
			res, callErr := next(ctx, call)
			lb.Report(original, callErr == nil)
			return res, callErr
		}
	}
}
