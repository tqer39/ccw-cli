package picker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestIcon(t *testing.T) {
	cases := []struct {
		s    worktree.Status
		want string
	}{
		{worktree.StatusPushed, "✅"},
		{worktree.StatusLocalOnly, "⚠"},
		{worktree.StatusDirty, "⛔"},
		{worktree.Status(99), "•"},
	}
	for _, tc := range cases {
		if got := Icon(tc.s); got != tc.want {
			t.Errorf("Icon(%s) = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestActionString(t *testing.T) {
	cases := []struct {
		a    Action
		want string
	}{
		{ActionCancel, "cancel"},
		{ActionResume, "resume"},
		{ActionDelete, "delete"},
		{ActionNew, "new"},
	}
	for _, tc := range cases {
		if got := tc.a.String(); got != tc.want {
			t.Errorf("Action(%d).String() = %q, want %q", tc.a, got, tc.want)
		}
	}
}

func TestUpdateList_EnterOnNewQuits(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.Action() != ActionNew {
		t.Errorf("Action = %s, want new", m.Action())
	}
	if cmd == nil {
		t.Error("Enter on [new] should emit tea.Quit cmd")
	}
}

func TestUpdateList_QKeyCancels(t *testing.T) {
	m := New(nil)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = next.(Model)
	if m.Action() != ActionCancel {
		t.Errorf("Action = %s, want cancel", m.Action())
	}
	if cmd == nil {
		t.Error("q should emit tea.Quit cmd")
	}
}

func TestUpdateList_EnterOnWorktreeEntersMenu(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.state != stateMenu {
		t.Errorf("state = %d, want stateMenu", m.state)
	}
	if m.selIdx != 0 {
		t.Errorf("selIdx = %d, want 0", m.selIdx)
	}
}

func TestView_ListRendersItems(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/feat-x", Branch: "feat-x", Status: worktree.StatusPushed},
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	out := m.View()
	if !strings.Contains(out, "feat-x") {
		t.Errorf("View() missing branch name:\n%s", out)
	}
	if !strings.Contains(out, "[new]") || !strings.Contains(out, "[quit]") {
		t.Errorf("View() missing synthetic rows:\n%s", out)
	}
}
