package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/grpcurl-batch/internal/runner"
)

// Format represents the output format type.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Formatter writes call results in a specified format.
type Formatter struct {
	format Format
	out    io.Writer
}

// New creates a new Formatter with the given format and writer.
func New(format Format, out io.Writer) *Formatter {
	return &Formatter{format: format, out: out}
}

// WriteResults writes a slice of results using the configured format.
func (f *Formatter) WriteResults(results []runner.Result) error {
	switch f.format {
	case FormatJSON:
		return f.writeJSON(results)
	default:
		return f.writeText(results)
	}
}

func (f *Formatter) writeText(results []runner.Result) error {
	w := tabwriter.NewWriter(f.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CALL\tSTATUS\tATTEMPTS\tDURATION\tERROR")
	for _, r := range results {
		status := "OK"
		if !r.Success {
			status = "FAIL"
		}
		errMsg := ""
		if r.Err != nil {
			errMsg = r.Err.Error()
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			r.CallName,
			status,
			r.Attempts,
			r.Duration.Round(time.Millisecond),
			errMsg,
		)
	}
	return w.Flush()
}

func (f *Formatter) writeJSON(results []runner.Result) error {
	enc := json.NewEncoder(f.out)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
