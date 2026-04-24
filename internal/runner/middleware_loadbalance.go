package runner

// DefaultLoadBalancer builds a round-robin balancer from the provided address
// pool and wraps it as a Middleware. It panics if the pool is empty, matching
// the convention used by other Default* helpers in this package.
func DefaultLoadBalancer(addrs []string) Middleware {
	lb, err := NewRoundRobinBalancer(addrs)
	if err != nil {
		panic("runner: DefaultLoadBalancer: " + err.Error())
	}
	return WithLoadBalancer(lb)
}

// LoadBalancerStatus returns the current pick counter and pool size for
// observability purposes. It is safe to call concurrently.
func LoadBalancerStatus(lb *RoundRobinBalancer) (pickCount uint64, poolSize int) {
	return lb.counter, len(lb.addrs)
}
