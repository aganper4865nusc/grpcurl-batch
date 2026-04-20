package runner

import (
	"context"
	"testing"
)

func TestAdmission_BelowThreshold_Allows(t *testing.T) {
	ap := NewAdmissionPolicy(0.8)
	ap.SetLoad(0.5)
	if err := ap.Admit(); err != nil {
		t.Fatalf("expected admission, got %v", err)
	}
}

func TestAdmission_AtThreshold_Rejects(t *testing.T) {
	ap := NewAdmissionPolicy(0.8)
	ap.SetLoad(0.8)
	if err := ap.Admit(); err != ErrAdmissionDenied {
		t.Fatalf("expected ErrAdmissionDenied, got %v", err)
	}
}

func TestAdmission_AboveThreshold_Rejects(t *testing.T) {
	ap := NewAdmissionPolicy(0.5)
	ap.SetLoad(0.9)
	if err := ap.Admit(); err != ErrAdmissionDenied {
		t.Fatalf("expected ErrAdmissionDenied, got %v", err)
	}
}

func TestAdmission_ZeroLoad_Allows(t *testing.T) {
	ap := NewAdmissionPolicy(0.5)
	// default load is 0
	if err := ap.Admit(); err != nil {
		t.Fatalf("expected admission at zero load, got %v", err)
	}
}

func TestAdmission_Wrap_AllowsCall(t *testing.T) {
	ap := NewAdmissionPolicy(0.9)
	ap.SetLoad(0.1)
	called := false
	err := ap.Wrap(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("expected call to proceed, err=%v called=%v", err, called)
	}
}

func TestAdmission_Wrap_RejectsCall(t *testing.T) {
	ap := NewAdmissionPolicy(0.5)
	ap.SetLoad(0.7)
	called := false
	err := ap.Wrap(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})
	if err != ErrAdmissionDenied {
		t.Fatalf("expected ErrAdmissionDenied, got %v", err)
	}
	if called {
		t.Fatal("fn should not have been called")
	}
}

func TestAdmission_Wrap_CancelledContext(t *testing.T) {
	ap := NewAdmissionPolicy(0.9)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ap.Wrap(ctx, func(_ context.Context) error { return nil })
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestAdmission_ClampThreshold(t *testing.T) {
	ap := NewAdmissionPolicy(2.5)
	if ap.threshold != 1.0 {
		t.Fatalf("expected threshold clamped to 1.0, got %v", ap.threshold)
	}
	ap2 := NewAdmissionPolicy(-1)
	if ap2.threshold != 0 {
		t.Fatalf("expected threshold clamped to 0, got %v", ap2.threshold)
	}
}
