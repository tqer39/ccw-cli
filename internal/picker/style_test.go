package picker

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestPRBadge_NoColorLowercase(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cases := map[string]string{
		"OPEN":   "[open]",
		"DRAFT":  "[draft]",
		"MERGED": "[merged]",
		"CLOSED": "[closed]",
	}
	for in, want := range cases {
		got := PRBadge(in)
		if got != want {
			t.Errorf("PRBadge(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPRBadge_ColoredContainsLabel(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// Force a color profile; go test has no TTY so lipgloss would otherwise
	// strip ANSI codes and the test couldn't distinguish colored output.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	for _, state := range []string{"OPEN", "DRAFT", "MERGED", "CLOSED"} {
		got := PRBadge(state)
		if !strings.Contains(got, "["+state+"]") {
			t.Errorf("PRBadge(%q) should contain [%s], got %q", state, state, got)
		}
		if !strings.Contains(got, "\x1b[") {
			t.Errorf("PRBadge(%q) expected ANSI escape when NO_COLOR unset, got %q", state, got)
		}
	}
}

func TestPRBadge_UnknownState(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := PRBadge("WEIRD")
	if got != "[weird]" {
		t.Errorf("PRBadge(WEIRD) = %q, want [weird]", got)
	}
}
