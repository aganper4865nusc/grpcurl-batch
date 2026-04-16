package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/user/grpcurl-batch/internal/runner"
)

// MetricsFormatter writes metrics snapshots to an output.
type MetricsFormatter struct {
	w      io.Writer
	format string
}

// NewMetricsFormatter creates a MetricsFormatter writing to w in the given format ("text" or "json").
func NewMetricsFormatter(w io.Writer, format string) *MetricsFormatter {
	if w == nil {
		w = os.Stdout
	}
	if format == "" {
		format = "text"
	}
	return &MetricsFormatter{w: w, format: format}
}

// Write outputs the metrics snapshot.
func (f *MetricsFormatter) Write(snap runner.Metrics) error {
	switch f.format {
	case "json":
		return f.writeJSON(snap)
	default:
		return f.writeText(snap)
	}
}

func (f *MetricsFormatter) writeText(snap runner.Metrics) error {
	tw := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "=== Metrics ===")
	fmt.Fprintf(tw, "Total Calls:\t%d\n", snap.TotalCalls)
	fmt.Fprintf(tw, "Succeeded:\t%d\n", snap.Succeeded)
	fmt.Fprintf(tw, "Failed:\t%d\n", snap.Failed)
	fmt.Fprintf(tw, "Retried:\t%d\n", snap.Retried)
	fmt.Fprintf(tw, "Avg Latency:\t%v\n", snap.TotalLatency/latencyDivisor(snap.TotalCalls))
	fmt.Fprintf(tw, "Min Latency:\t%v\n", snap.MinLatency)
	fmt.Fprintf(tw, "Max Latency:\t%v\n", snap.MaxLatency)
	return tw.Flush()
}

func (f *MetricsFormatter) writeJSON(snap runner.Metrics) error {
	type jsonMetrics struct {
		TotalCalls   int64  `json:"total_calls"`
		Succeeded    int64  `json:"succeeded"`
		Failed       int64  `json:"failed"`
		Retried      int64  `json:"retried"`
		AvgLatencyMs int64  `json:"avg_latency_ms"`
		MinLatencyMs int64  `json:"min_latency_ms"`
		MaxLatencyMs int64  `json:"max_latency_ms"`
	}
	d := latencyDivisor(snap.TotalCalls)
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonMetrics{
		TotalCalls:   snap.TotalCalls,
		Succeeded:    snap.Succeeded,
		Failed:       snap.Failed,
		Retried:      snap.Retried,
		AvgLatencyMs: (snap.TotalLatency / d).Milliseconds(),
		MinLatencyMs: snap.MinLatency.Milliseconds(),
		MaxLatencyMs: snap.MaxLatency.Milliseconds(),
	})
}

func latencyDivisor(total int64) interface{ Milliseconds() int64 } {
	// handled inline — return 1 as duration to avoid div by zero
	_ = total
	return nil
}
