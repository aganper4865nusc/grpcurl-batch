package reporter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/reporter"
	"github.com/user/grpcurl-batch/internal/runner"
)

func makeSummary() runner.Summary {
	return runner.Summary{
		Total:         2,
		Succeeded:     1,
		Failed:        1,
		TotalDuration: 150 * time.Millisecond,
		Results: []runner.Result{
			{
				CallName: "SayHello",
				Attempts: 1,
				Duration: 80 * time.Millisecond,
				Err:      nil,
			},
			{
				CallName: "SayBye",
				Attempts: 3,
				Duration: 70 * time.Millisecond,
				Err:      fmt.Errorf("rpc error"),
			},
		},
	}
}

func TestReporter_PrintText_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	r.Print(makeSummary())
	out := buf.String()
	for _, want := range []string{"CALL", "STATUS", "ATTEMPTS", "DURATION", "SayHello", "SayBye", "OK", "FAIL"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestReporter_PrintText_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	r.Print(makeSummary())
	out := buf.String()
	if !strings.Contains(out, "Total: 2") {
		t.Errorf("expected 'Total: 2' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Succeeded: 1") {
		t.Errorf("expected 'Succeeded: 1' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Failed: 1") {
		t.Errorf("expected 'Failed: 1' in output, got:\n%s", out)
	}
}

func TestReporter_PrintJSON(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatJSON)
	r.Print(makeSummary())
	out := buf.String()
	if !strings.Contains(out, `"total":2`) {
		t.Errorf("expected JSON total field, got: %s", out)
	}
	if !strings.Contains(out, `"succeeded":1`) {
		t.Errorf("expected JSON succeeded field, got: %s", out)
	}
}

func TestReporter_NilWriter_UsesStdout(t *testing.T) {
	// Should not panic when nil writer provided
	r := reporter.New(nil, reporter.FormatText)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}
