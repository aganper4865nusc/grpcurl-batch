package runner

// WithLabeler returns a middleware that applies the given Labeler to each call
// and stores the resulting LabelSet in the call's Metadata map before
// forwarding to the next handler. Existing metadata keys are preserved.
func WithLabeler(l *Labeler) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx interface{ Done() <-chan struct{} }, call Call) (Result, error) {
			if l != nil {
				ls := l.Apply(call)
				if call.Metadata == nil {
					call.Metadata = make(map[string]string, len(ls))
				}
				for k, v := range ls {
					if _, exists := call.Metadata[k]; !exists {
						call.Metadata[k] = v
					}
				}
			}
			return next(ctx, call)
		}
	}
}

// LabelerStatus returns a snapshot of the labels that would be applied to a
// call without executing it. Useful for debugging pipeline configuration.
func LabelerStatus(l *Labeler, call Call) LabelSet {
	if l == nil {
		return LabelSet{}
	}
	return l.Apply(call)
}
