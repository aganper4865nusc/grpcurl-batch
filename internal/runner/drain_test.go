package runner

import (
	"context"
	"testing"
	"time"
)

func TestDrainPolicy_NoInflight_WaitReturnsImmediately(t *testing.T) {
	d := NewDrainPolicy(time.Second)
	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDrainPolicy_ReleasedBeforeWait(t *testing.T) {
	d := NewDrainPolicy(time.Second)
	d.Acquire()
	d.Release()
	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDrainPolicy_WaitsForRelease(t *testing.T) {
	d := NewDrainPolicy(2 * time.Second)
	d.Acquire()
	go func() {
		time.Sleep(50 * time.Millisecond)
		d.Release()
	}()
	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDrainPolicy_TimeoutExpires(t *testing.T) {
	d := NewDrainPolicy(50 * time.Millisecond)
	d.Acquire()
	ctx := context.Background()
	err := d.Wait(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestDrainPolicy_ContextCancelled(t *testing.T) {
	d := NewDrainPolicy(5 * time.Second)
	d.Acquire()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := d.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestDrainPolicy_InFlight_Count(t *testing.T) {
	d := NewDrainPolicy(time.Second)
	d.Acquire()
	d.Acquire()
	if d.InFlight() != 2 {
		t.Fatalf("expected 2, got %d", d.InFlight())
	}
	d.Release()
	if d.InFlight() != 1 {
		t.Fatalf("expected 1, got %d", d.InFlight())
	}
}
