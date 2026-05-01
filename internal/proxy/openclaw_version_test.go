package proxy

import "testing"

func TestParseOpenClawVersion(t *testing.T) {
	cases := []struct {
		in    string
		year  int
		month int
		patch int
		raw   string
	}{
		{"2026.4.29", 2026, 4, 29, "2026.4.29"},
		{"2026.10.0", 2026, 10, 0, "2026.10.0"},
		{" 2027.1.5 ", 2027, 1, 5, "2027.1.5"},
		{"not.a.version", 0, 0, 0, "not.a.version"}, // raw kept, numeric zeroed
		{"2026.4", 0, 0, 0, "2026.4"},               // wrong arity, raw only
		{"", 0, 0, 0, ""},
	}
	for _, c := range cases {
		got := parseOpenClawVersion(c.in)
		if got.Year != c.year || got.Month != c.month || got.Patch != c.patch || got.Raw != c.raw {
			t.Errorf("parse(%q) = {%s %d %d %d}, want {%s %d %d %d}",
				c.in, got.Raw, got.Year, got.Month, got.Patch,
				c.raw, c.year, c.month, c.patch)
		}
	}
}

func TestOpenClawVersionGTE(t *testing.T) {
	v429 := parseOpenClawVersion("2026.4.29")
	if !v429.gte(2026, 4, 29) {
		t.Errorf("4.29 should be >= 4.29")
	}
	if !v429.gte(2026, 4, 0) {
		t.Errorf("4.29 should be >= 4.0")
	}
	if !v429.gte(2026, 3, 99) {
		t.Errorf("4.29 should be >= 3.99 (lower month)")
	}
	if v429.gte(2026, 5, 0) {
		t.Errorf("4.29 should NOT be >= 5.0")
	}
	if v429.gte(2027, 1, 0) {
		t.Errorf("4.29 should NOT be >= 2027.1.0")
	}
	// detection-failed sentinel: empty Raw answers true so writer defaults
	// to the newest schema branch instead of locking to old behaviour.
	empty := openClawVersion{}
	if !empty.gte(9999, 99, 99) {
		t.Errorf("empty version should be >= anything (newest-schema default)")
	}
}

func TestOpenClawSchemaTag(t *testing.T) {
	// Every reachable version is currently schemaTag()=v1. This test will
	// start failing the moment somebody adds a new branch in
	// openclaw_version.go without revisiting the writer dispatch in
	// agent_apply.go — that's the point.
	for _, raw := range []string{"2026.4.26", "2026.4.29", "2026.5.0", "2027.1.0", ""} {
		v := parseOpenClawVersion(raw)
		if got := v.schemaTag(); got != "v1" {
			t.Errorf("schemaTag(%q) = %q, want v1 — if this is intentional, update applyOpenClawConfig to handle the new tag", raw, got)
		}
	}
}
