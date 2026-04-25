package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/yourorg/grpcurl-batch/internal/runner"
)

// ProfilerFormatter renders profiling records as text or JSON.
type ProfilerFormatter struct {
	w io.Writer
}

// NewProfilerFormatter creates a formatter writing to w.
// If w is nil, os.Stdout is used.
func NewProfilerFormatter(w io.Writer) *ProfilerFormatter {
	if w == nil {
		w = os.Stdout
	}
	return &ProfilerFormatter{w: w}
}

// WriteText outputs a tab-aligned table of profile records.
func (f *ProfilerFormatter) WriteText(records []runner.ProfileRecord) {
	tw := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CALL_ID\tSERVICE\tMETHOD\tDURATION\tSUCCESS\tSTARTED")
	for _, r := range records {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%v\t%s\n",
			r.CallID,
			r.Service,
			r.Method,
			r.Duration.Round(time.Millisecond),
			r.Success,
			r.StartedAt.Format(time.RFC3339),
		)
	}
	_ = tw.Flush()
}

// WriteJSON serialises records as a JSON array.
func (f *ProfilerFormatter) WriteJSON(records []runner.ProfileRecord) error {
	type row struct {
		CallID    string  `json:"call_id"`
		Service   string  `json:"service"`
		Method    string  `json:"method"`
		DurationMs float64 `json:"duration_ms"`
		Success   bool    `json:"success"`
		StartedAt string  `json:"started_at"`
	}
	rows := make([]row, len(records))
	for i, r := range records {
		rows[i] = row{
			CallID:     r.CallID,
			Service:    r.Service,
			Method:     r.Method,
			DurationMs: float64(r.Duration.Milliseconds()),
			Success:    r.Success,
			StartedAt:  r.StartedAt.Format(time.RFC3339),
		}
	}
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
