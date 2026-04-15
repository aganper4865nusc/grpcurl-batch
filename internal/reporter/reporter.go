package reporter

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/grpcurl-batch/internal/runner"
)

// Format defines the output format for the reporter.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Reporter writes execution results to an output stream.
type Reporter struct {
	w      io.Writer
	format Format
}

// New creates a new Reporter writing to the given writer.
func New(w io.Writer, format Format) *Reporter {
	if w == nil {
		w = os.Stdout
	}
	return &Reporter{w: w, format: format}
}

// Print writes a human-readable summary of results.
func (r *Reporter) Print(summary runner.Summary) {
	switch r.format {
	case FormatJSON:
		r.printJSON(summary)
	default:
		r.printText(summary)
	}
}

func (r *Reporter) printText(s runner.Summary) {
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CALL\tSTATUS\tATTEMPTS\tDURATION")
	fmt.Fprintln(tw, "----\t------\t--------\t--------")
	for _, res := range s.Results {
		status := "OK"
		if res.Err != nil {
			status = "FAIL"
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			res.CallName,
			status,
			res.Attempts,
			res.Duration.Round(time.Millisecond),
		)
	}
	tw.Flush()
	fmt.Fprintf(r.w, "\nTotal: %d | Succeeded: %d | Failed: %d | Duration: %s\n",
		s.Total, s.Succeeded, s.Failed, s.TotalDuration.Round(time.Millisecond))
}

func (r *Reporter) printJSON(s runner.Summary) {
	fmt.Fprintf(r.w, `{"total":%d,"succeeded":%d,"failed":%d,"duration_ms":%d}\n`,
		s.Total, s.Succeeded, s.Failed, s.TotalDuration.Milliseconds())
}
