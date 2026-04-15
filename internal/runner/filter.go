package runner

import (
	"strings"

	"github.com/nickcoast/grpcurl-batch/internal/manifest"
)

// FilterFunc determines whether a call should be included in execution.
type FilterFunc func(call manifest.Call) bool

// FilterOptions holds criteria for filtering calls from a manifest.
type FilterOptions struct {
	// Tags filters calls that have at least one matching tag.
	Tags []string
	// ServicePrefix filters calls whose service name starts with the given prefix.
	ServicePrefix string
	// MethodContains filters calls whose method name contains the given substring.
	MethodContains string
}

// BuildFilter constructs a FilterFunc from the provided FilterOptions.
// If no options are set, the returned filter accepts all calls.
func BuildFilter(opts FilterOptions) FilterFunc {
	return func(call manifest.Call) bool {
		if len(opts.Tags) > 0 && !hasAnyTag(call.Tags, opts.Tags) {
			return false
		}
		if opts.ServicePrefix != "" && !strings.HasPrefix(call.Service, opts.ServicePrefix) {
			return false
		}
		if opts.MethodContains != "" && !strings.Contains(call.Method, opts.MethodContains) {
			return false
		}
		return true
	}
}

// ApplyFilter returns the subset of calls accepted by the given FilterFunc.
func ApplyFilter(calls []manifest.Call, fn FilterFunc) []manifest.Call {
	if fn == nil {
		return calls
	}
	out := make([]manifest.Call, 0, len(calls))
	for _, c := range calls {
		if fn(c) {
			out = append(out, c)
		}
	}
	return out
}

// hasAnyTag returns true if the call's tag list contains at least one of the wanted tags.
func hasAnyTag(callTags, wanted []string) bool {
	set := make(map[string]struct{}, len(callTags))
	for _, t := range callTags {
		set[t] = struct{}{}
	}
	for _, w := range wanted {
		if _, ok := set[w]; ok {
			return true
		}
	}
	return false
}
