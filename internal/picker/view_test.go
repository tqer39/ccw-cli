package picker

import (
	"strings"
	"testing"

	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestView_FooterShowsTips(t *testing.T) {
	m := New([]worktree.Info{{Path: "/x/.claude/worktrees/a", Branch: "a"}})
	m.ghAvailable = true
	m.tip = "test tip line"
	m.state = stateList
	out := m.View()
	if !strings.Contains(out, "💡 Tip: test tip line") {
		t.Errorf("View footer missing tip:\n%s", out)
	}
}

func TestView_FooterShowsGhHintWhenUnavailable(t *testing.T) {
	m := New([]worktree.Info{{Path: "/x/.claude/worktrees/a", Branch: "a"}})
	m.ghAvailable = false
	m.tip = "should-not-show"
	m.state = stateList
	out := m.View()
	if !strings.Contains(out, "Install gh") {
		t.Errorf("View should show gh hint when gh unavailable:\n%s", out)
	}
	if strings.Contains(out, "should-not-show") {
		t.Errorf("View should not show tip when gh unavailable:\n%s", out)
	}
}
