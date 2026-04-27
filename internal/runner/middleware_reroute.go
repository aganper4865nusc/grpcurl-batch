package runner

// WithReroute returns a middleware that rewrites Call.Address using the
// provided ReroutePolicy before passing the call to the next handler.
//
// Example:
//
//	p := NewReroutePolicy([]RerouteRule{{From: "legacy:443", To: "v2:443"}})
//	mw := WithReroute(p)
func WithReroute(p *ReroutePolicy) Middleware {
	return func(next CallFunc) CallFunc {
		return p.Wrap(next)
	}
}

// DefaultReroute creates a Middleware from a slice of RerouteRule values
// without requiring callers to construct a ReroutePolicy directly.
func DefaultReroute(rules []RerouteRule) Middleware {
	return WithReroute(NewReroutePolicy(rules))
}
