package runner

import (
	"strings"
	"testing"
)

func baseLabelerCall(method string) Call {
	return Call{
		Method:  method,
		Address: "localhost:50051",
	}
}

func TestLabeler_StaticLabelApplied(t *testing.T) {
	l := NewLabeler(LabelSet{"env": "test"})
	ls := l.Apply(baseLabelerCall("pkg.Svc/Ping"))
	if ls["env"] != "test" {
		t.Fatalf("expected env=test, got %q", ls["env"])
	}
}

func TestLabeler_NilStatic_NoNilPanic(t *testing.T) {
	l := NewLabeler(nil)
	ls := l.Apply(baseLabelerCall("pkg.Svc/Ping"))
	if ls == nil {
		t.Fatal("expected non-nil LabelSet")
	}
}

func TestLabeler_DerivedLabelAdded(t *testing.T) {
	l := NewLabeler(nil)
	l.AddDerived("region", func(_ Call) string { return "us-east" })
	ls := l.Apply(baseLabelerCall("pkg.Svc/Ping"))
	if ls["region"] != "us-east" {
		t.Fatalf("expected region=us-east, got %q", ls["region"])
	}
}

func TestLabeler_DerivedEmptyValue_Skipped(t *testing.T) {
	l := NewLabeler(nil)
	l.AddDerived("skip", func(_ Call) string { return "" })
	ls := l.Apply(baseLabelerCall("pkg.Svc/Ping"))
	if _, ok := ls["skip"]; ok {
		t.Fatal("expected empty-value derived label to be skipped")
	}
}

func TestLabeler_DerivedOverridesStatic(t *testing.T) {
	l := NewLabeler(LabelSet{"env": "prod"})
	l.AddDerived("env", func(_ Call) string { return "staging" })
	ls := l.Apply(baseLabelerCall("pkg.Svc/Ping"))
	if ls["env"] != "staging" {
		t.Fatalf("expected derived env=staging to override static, got %q", ls["env"])
	}
}

func TestServiceLabel_ExtractsService(t *testing.T) {
	l := NewLabeler(nil)
	l.AddDerived(ServiceLabel().Key, ServiceLabel().Fn)
	ls := l.Apply(baseLabelerCall("mypackage.MyService/DoThing"))
	if ls["service"] != "MyService" {
		t.Fatalf("expected service=MyService, got %q", ls["service"])
	}
}

func TestMethodLabel_ExtractsMethod(t *testing.T) {
	l := NewLabeler(nil)
	l.AddDerived(MethodLabel().Key, MethodLabel().Fn)
	ls := l.Apply(baseLabelerCall("mypackage.MyService/DoThing"))
	if ls["method"] != "DoThing" {
		t.Fatalf("expected method=DoThing, got %q", ls["method"])
	}
}

func TestFormatLabels_Empty(t *testing.T) {
	if s := FormatLabels(LabelSet{}); s != "" {
		t.Fatalf("expected empty string, got %q", s)
	}
}

func TestFormatLabels_ContainsKeyValue(t *testing.T) {
	s := FormatLabels(LabelSet{"env": "test"})
	if !strings.Contains(s, "env=test") {
		t.Fatalf("expected env=test in %q", s)
	}
}
