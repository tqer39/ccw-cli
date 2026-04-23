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

func selectFirstWorktree(m Model) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(Model)
}

func TestMenu_ResumeEmitsQuit(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = selectFirstWorktree(m)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = next.(Model)
	if m.Action() != ActionResume {
		t.Errorf("Action = %s, want resume", m.Action())
	}
	if cmd == nil {
		t.Error("r should emit tea.Quit")
	}
	if m.Selection().Branch != "x" || m.Selection().Path != "/a/.claude/worktrees/x" {
		t.Errorf("Selection = %+v", m.Selection())
	}
}

func TestMenu_BackReturnsToList(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = selectFirstWorktree(m)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = next.(Model)
	if m.state != stateList {
		t.Errorf("state = %d, want stateList after back", m.state)
	}
}

func TestMenu_DeleteEntersConfirm(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusDirty},
	})
	m = selectFirstWorktree(m)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = next.(Model)
	if m.state != stateDeleteConfirm {
		t.Errorf("state = %d, want stateDeleteConfirm", m.state)
	}
}

func TestMenuView_ContainsSelectionSummary(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/feat-y", Branch: "feat-y", Status: worktree.StatusLocalOnly},
	})
	m = selectFirstWorktree(m)
	out := m.View()
	for _, want := range []string{"feat-y", "local-only", "[r]", "[d]", "[b]"} {
		if !strings.Contains(out, want) {
			t.Errorf("menuView missing %q:\n%s", want, out)
		}
	}
}

func goToDeleteConfirm(m Model) Model {
	m = selectFirstWorktree(m)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	return next.(Model)
}

func TestDeleteConfirm_YesOnCleanConfirmsWithoutForce(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = goToDeleteConfirm(m)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = next.(Model)
	if m.Action() != ActionDelete {
		t.Errorf("Action = %s, want delete", m.Action())
	}
	if m.Selection().ForceDelete {
		t.Error("ForceDelete should be false for clean worktree")
	}
	if cmd == nil {
		t.Error("y should emit tea.Quit")
	}
}

func TestDeleteConfirm_YesOnDirtyForces(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusDirty},
	})
	m = goToDeleteConfirm(m)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = next.(Model)
	if !m.Selection().ForceDelete {
		t.Error("ForceDelete should be true for dirty worktree")
	}
}

func TestDeleteConfirm_NoReturnsToList(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = goToDeleteConfirm(m)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = next.(Model)
	if m.state != stateList {
		t.Errorf("state = %d, want stateList after n", m.state)
	}
}

func TestDeleteConfirmView_ShowsForceForDirty(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusDirty},
	})
	m = goToDeleteConfirm(m)
	out := m.View()
	if !strings.Contains(out, "--force") {
		t.Errorf("deleteConfirmView missing --force marker for dirty:\n%s", out)
	}
}

func TestDeleteConfirmView_HidesForceForClean(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = goToDeleteConfirm(m)
	out := m.View()
	if strings.Contains(out, "--force") {
		t.Errorf("deleteConfirmView should not show --force for clean worktree:\n%s", out)
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
