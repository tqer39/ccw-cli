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
	if !strings.Contains(out.Content, "💡 Tip: test tip line") {
		t.Errorf("View footer missing tip:\n%s", out.Content)
	}
}

func TestView_FooterShowsGhHintWhenUnavailable(t *testing.T) {
	m := New([]worktree.Info{{Path: "/x/.claude/worktrees/a", Branch: "a"}})
	m.ghAvailable = false
	m.tip = "should-not-show"
	m.state = stateList
	out := m.View()
	if !strings.Contains(out.Content, "Install gh") {
		t.Errorf("View should show gh hint when gh unavailable:\n%s", out.Content)
	}
	if strings.Contains(out.Content, "should-not-show") {
		t.Errorf("View should not show tip when gh unavailable:\n%s", out.Content)
	}
}
