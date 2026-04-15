package runner

import (
	"testing"

	"github.com/nickcoast/grpcurl-batch/internal/manifest"
)

func makeCall(service, method string, tags []string) manifest.Call {
	return manifest.Call{
		Service: service,
		Method:  method,
		Tags:    tags,
	}
}

func TestBuildFilter_NoOptions_AcceptsAll(t *testing.T) {
	calls := []manifest.Call{
		makeCall("svc.Foo", "Bar", nil),
		makeCall("svc.Baz", "Qux", []string{"smoke"}),
	}
	fn := BuildFilter(FilterOptions{})
	result := ApplyFilter(calls, fn)
	if len(result) != len(calls) {
		t.Fatalf("expected %d calls, got %d", len(calls), len(result))
	}
}

func TestBuildFilter_ByTag(t *testing.T) {
	calls := []manifest.Call{
		makeCall("svc.A", "M1", []string{"smoke", "regression"}),
		makeCall("svc.B", "M2", []string{"regression"}),
		makeCall("svc.C", "M3", []string{"load"}),
	}
	fn := BuildFilter(FilterOptions{Tags: []string{"smoke"}})
	result := ApplyFilter(calls, fn)
	if len(result) != 1 {
		t.Fatalf("expected 1 call, got %d", len(result))
	}
	if result[0].Method != "M1" {
		t.Errorf("unexpected method: %s", result[0].Method)
	}
}

func TestBuildFilter_ByServicePrefix(t *testing.T) {
	calls := []manifest.Call{
		makeCall("payment.Service", "Charge", nil),
		makeCall("auth.Service", "Login", nil),
		makeCall("payment.Service", "Refund", nil),
	}
	fn := BuildFilter(FilterOptions{ServicePrefix: "payment"})
	result := ApplyFilter(calls, fn)
	if len(result) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(result))
	}
}

func TestBuildFilter_ByMethodContains(t *testing.T) {
	calls := []manifest.Call{
		makeCall("svc.A", "GetUser", nil),
		makeCall("svc.B", "ListUsers", nil),
		makeCall("svc.C", "DeleteItem", nil),
	}
	fn := BuildFilter(FilterOptions{MethodContains: "User"})
	result := ApplyFilter(calls, fn)
	if len(result) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(result))
	}
}

func TestBuildFilter_CombinedOptions(t *testing.T) {
	calls := []manifest.Call{
		makeCall("payment.Svc", "GetPayment", []string{"smoke"}),
		makeCall("payment.Svc", "ListPayments", []string{"regression"}),
		makeCall("auth.Svc", "GetSession", []string{"smoke"}),
	}
	fn := BuildFilter(FilterOptions{
		Tags:          []string{"smoke"},
		ServicePrefix: "payment",
	})
	result := ApplyFilter(calls, fn)
	if len(result) != 1 {
		t.Fatalf("expected 1 call, got %d", len(result))
	}
	if result[0].Method != "GetPayment" {
		t.Errorf("unexpected method: %s", result[0].Method)
	}
}

func TestApplyFilter_NilFilter_ReturnsAll(t *testing.T) {
	calls := []manifest.Call{
		makeCall("svc.A", "M1", nil),
		makeCall("svc.B", "M2", nil),
	}
	result := ApplyFilter(calls, nil)
	if len(result) != len(calls) {
		t.Fatalf("expected %d calls, got %d", len(calls), len(result))
	}
}
