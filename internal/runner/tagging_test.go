package runner

import (
	"testing"
)

func baseTagCall(tags ...string) Call {
	return Call{
		Service: "svc",
		Method:  "Method",
		Address: "localhost:50051",
		Tags:    tags,
	}
}

func TestTagEnricher_StaticTagAdded(t *testing.T) {
	te := NewTagEnricher(map[string]string{"env": "prod"})
	c := te.Enrich(baseTagCall())
	if !hasAnyTag(c, "env=prod") {
		t.Errorf("expected env=prod tag, got %v", c.Tags)
	}
}

func TestTagEnricher_StaticTagDoesNotOverrideExisting(t *testing.T) {
	te := NewTagEnricher(map[string]string{"env": "prod"})
	c := te.Enrich(baseTagCall("env=staging"))
	if !hasAnyTag(c, "env=staging") {
		t.Errorf("expected env=staging to be preserved, got %v", c.Tags)
	}
	for _, tag := range c.Tags {
		if tag == "env=prod" {
			t.Errorf("static tag should not override existing tag")
		}
	}
}

func TestTagEnricher_DerivedTagAdded(t *testing.T) {
	te := NewTagEnricher(nil)
	te.AddDerived("method", func(c Call) string { return c.Method })
	c := te.Enrich(baseTagCall())
	if !hasAnyTag(c, "method=Method") {
		t.Errorf("expected method=Method tag, got %v", c.Tags)
	}
}

func TestTagEnricher_DerivedTagEmptyValueSkipped(t *testing.T) {
	te := NewTagEnricher(nil)
	te.AddDerived("empty", func(c Call) string { return "" })
	c := te.Enrich(baseTagCall())
	for _, tag := range c.Tags {
		k, _, ok := splitTag(tag)
		if ok && k == "empty" {
			t.Errorf("empty derived tag should not be added")
		}
	}
}

func TestTagEnricher_AddStatic_AppliedToSubsequentEnrich(t *testing.T) {
	te := NewTagEnricher(nil)
	te.AddStatic("region", "us-east-1")
	c := te.Enrich(baseTagCall())
	if !hasAnyTag(c, "region=us-east-1") {
		t.Errorf("expected region=us-east-1, got %v", c.Tags)
	}
}

func TestTagEnricher_MultipleStaticAndDerived(t *testing.T) {
	te := NewTagEnricher(map[string]string{"env": "dev", "team": "platform"})
	te.AddDerived("service", func(c Call) string { return c.Service })
	c := te.Enrich(baseTagCall())
	if len(c.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(c.Tags), c.Tags)
	}
}

func TestSplitTag_Valid(t *testing.T) {
	k, v, ok := splitTag("foo=bar")
	if !ok || k != "foo" || v != "bar" {
		t.Errorf("unexpected result: %q %q %v", k, v, ok)
	}
}

func TestSplitTag_Invalid(t *testing.T) {
	_, _, ok := splitTag("notakey")
	if ok {
		t.Errorf("expected false for tag without '='")
	}
}
