package version

import (
	"strings"
	"testing"
)

func TestString_Default(t *testing.T) {
	got := String()
	want := "ccw dev (commit: none, built: unknown)"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestString_WithInjectedValues(t *testing.T) {
	origV, origC, origD := Version, Commit, Date
	t.Cleanup(func() {
		Version, Commit, Date = origV, origC, origD
	})
	Version = "v1.2.3"
	Commit = "abc1234"
	Date = "2026-04-23T00:00:00Z"

	got := String()
	for _, sub := range []string{"v1.2.3", "abc1234", "2026-04-23T00:00:00Z"} {
		if !strings.Contains(got, sub) {
			t.Errorf("String() = %q, want substring %q", got, sub)
		}
	}
}
