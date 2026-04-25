package runner

import (
	"fmt"
	"strings"
)

// LabelSet is a map of string key-value labels attached to a call.
type LabelSet map[string]string

// Labeler enriches calls with computed or static labels before execution.
type Labeler struct {
	static  LabelSet
	derived []DerivedLabel
}

// DerivedLabel computes a label value from a call at runtime.
type DerivedLabel struct {
	Key string
	Fn  func(call Call) string
}

// NewLabeler creates a Labeler with optional static labels.
func NewLabeler(static LabelSet) *Labeler {
	if static == nil {
		static = LabelSet{}
	}
	return &Labeler{static: static}
}

// AddDerived registers a derived label that is computed from each call.
func (l *Labeler) AddDerived(key string, fn func(call Call) string) {
	l.derived = append(l.derived, DerivedLabel{Key: key, Fn: fn})
}

// Apply returns a copy of labels merged onto the call's existing metadata.
// Static labels are applied first; derived labels may override them.
func (l *Labeler) Apply(call Call) LabelSet {
	out := make(LabelSet, len(l.static)+len(l.derived))
	for k, v := range l.static {
		out[k] = v
	}
	for _, d := range l.derived {
		if v := d.Fn(call); v != "" {
			out[d.Key] = v
		}
	}
	return out
}

// ServiceLabel returns a derived label that extracts the service name from
// the call's Method field (format "package.Service/Method").
func ServiceLabel() DerivedLabel {
	return DerivedLabel{
		Key: "service",
		Fn: func(call Call) string {
			parts := strings.SplitN(call.Method, "/", 2)
			if len(parts) == 0 {
				return ""
			}
			seg := strings.Split(parts[0], ".")
			return seg[len(seg)-1]
		},
	}
}

// MethodLabel returns a derived label that extracts the bare method name.
func MethodLabel() DerivedLabel {
	return DerivedLabel{
		Key: "method",
		Fn: func(call Call) string {
			parts := strings.SplitN(call.Method, "/", 2)
			if len(parts) == 2 {
				return parts[1]
			}
			return call.Method
		},
	}
}

// FormatLabels serialises a LabelSet to a comma-separated key=value string.
func FormatLabels(ls LabelSet) string {
	if len(ls) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ls))
	for k, v := range ls {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}
