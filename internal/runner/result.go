package runner

import "time"

// CallResult holds the outcome of a single gRPC call attempt.
type CallResult struct {
	CallName  string
	Address   string
	Method    string
	Success   bool
	Attempts  int
	Output    string
	Error     string
	Duration  time.Duration
}

// Summary aggregates results from all calls in a batch run.
type Summary struct {
	Total     int
	Succeeded int
	Failed    int
	Results   []CallResult
	TotalTime time.Duration
}

// Summarize builds a Summary from a slice of CallResults.
func Summarize(results []CallResult, elapsed time.Duration) Summary {
	s := Summary{
		Total:     len(results),
		Results:   results,
		TotalTime: elapsed,
	}
	for _, r := range results {
		if r.Success {
			s.Succeeded++
		} else {
			s.Failed++
		}
	}
	return s
}

// SuccessRate returns the percentage of successful calls.
func (s Summary) SuccessRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Succeeded) / float64(s.Total) * 100
}
