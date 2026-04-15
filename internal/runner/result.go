package runner

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Summary aggregates run results.
type Summary struct {
	Total   int
	Success int
	Failed  int
	Results []Result
}

// Summarize builds a Summary from a slice of Results.
func Summarize(results []Result) Summary {
	s := Summary{Total: len(results), Results: results}
	for _, r := range results {
		if r.Err == nil {
			s.Success++
		} else {
			s.Failed++
		}
	}
	return s
}

// WriteTo writes a human-readable summary table to w.
func (s Summary) WriteTo(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tATTEMPTS\tDURATION\tSTATUS")
	fmt.Fprintln(tw, strings.Repeat("-", 60))

	for _, r := range s.Results {
		status := "OK"
		if r.Err != nil {
			status = fmt.Sprintf("ERR: %s", r.Err.Error())
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n",
			r.CallName,
			r.Attempts,
			r.Duration.Round(1000000),
			status,
		)
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(w, "\nTotal: %d | Success: %d | Failed: %d\n",
		s.Total, s.Success, s.Failed)
	return nil
}
