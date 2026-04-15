package output_test

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/grpcurl-batch/internal/output"
)

func TestProgressTracker_Increment_PrintsProgress(t *testing.T) {
	var buf bytes.Buffer
	pt := output.NewProgressTracker(&buf, 3, false)
	pt.Increment("SayHello", true)
	out := buf.String()
	if !strings.Contains(out, "[1/3]") {
		t.Errorf("expected progress indicator [1/3], got %q", out)
	}
	if !strings.Contains(out, "SayHello") {
		t.Errorf("expected call name in output, got %q", out)
	}
}

func TestProgressTracker_Increment_FailureMarker(t *testing.T) {
	var buf bytes.Buffer
	pt := output.NewProgressTracker(&buf, 1, false)
	pt.Increment("GetUser", false)
	out := buf.String()
	if !strings.Contains(out, "✗") {
		t.Errorf("expected failure marker in output, got %q", out)
	}
}

func TestProgressTracker_Silent_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	pt := output.NewProgressTracker(&buf, 2, true)
	pt.Increment("SayHello", true)
	pt.Increment("GetUser", false)
	if buf.Len() != 0 {
		t.Errorf("expected no output in silent mode, got %q", buf.String())
	}
	if pt.Done() != 2 {
		t.Errorf("expected Done()=2, got %d", pt.Done())
	}
}

func TestProgressTracker_ConcurrentSafety(t *testing.T) {
	var buf bytes.Buffer
	total := 50
	pt := output.NewProgressTracker(&buf, total, false)
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pt.Increment(fmt.Sprintf("call-%d", i), i%2 == 0)
		}(i)
	}
	wg.Wait()
	if pt.Done() != total {
		t.Errorf("expected Done()=%d, got %d", total, pt.Done())
	}
}
