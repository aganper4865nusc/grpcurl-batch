// Package output provides formatting and progress-tracking utilities
// for grpcurl-batch CLI output.
//
// Formatter writes gRPC call results as human-readable text tables or
// machine-readable JSON. ProgressTracker prints incremental progress
// to the terminal as calls complete, and is safe for concurrent use.
package output
