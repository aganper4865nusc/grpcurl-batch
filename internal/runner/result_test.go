package runner

import (
	"testing"
	"time"
)

func makeResult(success bool) CallResult {
	return CallResult{Success: success, Attempts: 1, Duration: time.Millisecond}
}

func TestSummarize_Counts(t *testing.T) {
	results := []CallResult{
		makeResult(true),
		makeResult(true),
		makeResult(false),
	}
	s := Summarize(results, 50*time.Millisecond)
	if s.Total != 3 {
		t.Errorf("expected Total=3, got %d", s.Total)
	}
	if s.Succeeded != 2 {
		t.Errorf("expected Succeeded=2, got %d", s.Succeeded)
	}
	if s.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", s.Failed)
	}
}

func TestSummarize_TotalTime(t *testing.T) {
	elapsed := 123 * time.Millisecond
	s := Summarize(nil, elapsed)
	if s.TotalTime != elapsed {
		t.Errorf("expected TotalTime=%v, got %v", elapsed, s.TotalTime)
	}
}

func TestSuccessRate_AllSuccess(t *testing.T) {
	s := Summarize([]CallResult{makeResult(true), makeResult(true)}, 0)
	if s.SuccessRate() != 100.0 {
		t.Errorf("expected 100%%, got %.2f", s.SuccessRate())
	}
}

func TestSuccessRate_NoResults(t *testing.T) {
	s := Summarize(nil, 0)
	if s.SuccessRate() != 0 {
		t.Errorf("expected 0%%, got %.2f", s.SuccessRate())
	}
}

func TestSuccessRate_Mixed(t *testing.T) {
	s := Summarize([]CallResult{makeResult(true), makeResult(false)}, 0)
	if s.SuccessRate() != 50.0 {
		t.Errorf("expected 50%%, got %.2f", s.SuccessRate())
	}
}
