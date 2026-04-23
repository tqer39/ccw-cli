package superpowers

import (
	"strings"
	"testing"
)

func TestPreamble_ContainsWorkflow(t *testing.T) {
	p := Preamble()
	for _, want := range []string{
		"brainstorming",
		"writing-plans",
		"executing-plans",
		"--worktree",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("Preamble() missing %q; got:\n%s", want, p)
		}
	}
}

func TestPreamble_EndsWithNewline(t *testing.T) {
	p := Preamble()
	if !strings.HasSuffix(p, "\n") {
		t.Errorf("Preamble() must end with newline; got %q", p[len(p)-5:])
	}
}
