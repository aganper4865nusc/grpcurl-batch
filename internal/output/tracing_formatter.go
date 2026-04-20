package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/example/grpcurl-batch/internal/runner"
)

// TracingFormatter writes span data in text or JSON format.
type TracingFormatter struct {
	w      io.Writer
	format string
}

func NewTracingFormatter(w io.Writer, format string) *TracingFormatter {
	if w == nil {
		w = os.Stdout
	}
	return &TracingFormatter{w: w, format: format}
}

func (tf *TracingFormatter) Write(spans []runner.Span) error {
	switch tf.format {
	case "json":
		return tf.writeJSON(spans)
	default:
		return tf.writeText(spans)
	}
}

func (tf *TracingFormatter) writeText(spans []runner.Span) error {
	tw := tabwriter.NewWriter(tf.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CALL\tMETHOD\tDURATION\tSUCCESS\tERROR")
	for _, s := range spans {
		errStr := ""
		if s.Err != nil {
			errStr = s.Err.Error()
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%s\n",
			s.CallID, s.Method, s.Duration(), s.Success, errStr)
	}
	return tw.Flush()
}

func (tf *TracingFormatter) writeJSON(spans []runner.Span) error {
	type row struct {
		CallID   string `json:"call_id"`
		Method   string `json:"method"`
		DurationMs float64 `json:"duration_ms"`
		Success  bool   `json:"success"`
		Error    string `json:"error,omitempty"`
	}
	rows := make([]row, len(spans))
	for i, s := range spans {
		errStr := ""
		if s.Err != nil {
			errStr = s.Err.Error()
		}
		rows[i] = row{
			CallID:     s.CallID,
			Method:     s.Method,
			DurationMs: float64(s.Duration().Milliseconds()),
			Success:    s.Success,
			Error:      errStr,
		}
	}
	enc := json.NewEncoder(tf.w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
