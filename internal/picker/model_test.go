package picker

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tqer39/ccw-cli/internal/gh"
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
		{worktree.StatusPrunable, "🧹"},
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
		{ActionResume, "run"},
		{ActionDelete, "delete"},
		{ActionNew, "new"},
		{ActionBulkDelete, "bulk-delete"},
	}
	for _, tc := range cases {
		if got := tc.a.String(); got != tc.want {
			t.Errorf("Action(%d).String() = %q, want %q", tc.a, got, tc.want)
		}
	}
}

func TestUpdate_DeleteAll_NoDirty(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a", Status: worktree.StatusPushed},
		{Branch: "b", Path: "/b", Status: worktree.StatusPushed},
	})
	m.bulkTargets = []int{0, 1}
	m.state = stateBulkConfirm
	got, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	mm := got.(Model)
	if mm.Action() != ActionBulkDelete {
		t.Errorf("want ActionBulkDelete, got %v", mm.Action())
	}
	if mm.bulkForce {
		t.Error("force should be false without dirty")
	}
	b := mm.Bulk()
	if len(b.Paths) != 2 {
		t.Errorf("want 2 paths, got %d", len(b.Paths))
	}
}

func TestUpdate_DeleteAll_WithDirty_Force(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a", Status: worktree.StatusPushed},
		{Branch: "b", Path: "/b", Status: worktree.StatusDirty},
	})
	m.bulkTargets = []int{0, 1}
	m.state = stateBulkConfirm
	got, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	mm := got.(Model)
	if !mm.bulkForce {
		t.Error("force should be true when dirty is included")
	}
}

func TestUpdate_DeleteAll_SkipDirty(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a", Status: worktree.StatusPushed},
		{Branch: "b", Path: "/b", Status: worktree.StatusDirty},
	})
	m.bulkTargets = []int{0, 1}
	m.state = stateBulkConfirm
	got, _ := m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	mm := got.(Model)
	if mm.Action() != ActionBulkDelete {
		t.Errorf("want ActionBulkDelete, got %v", mm.Action())
	}
	b := mm.Bulk()
	if len(b.Paths) != 1 || b.Paths[0] != "/a" {
		t.Errorf("want only /a, got %v", b.Paths)
	}
}

func TestUpdate_BulkFilter_TogglesAndConfirms(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a", Status: worktree.StatusPushed},
		{Branch: "b", Path: "/b", Status: worktree.StatusDirty},
	})
	m.state = stateBulkFilter
	m.bulkFilter = map[worktree.Status]bool{}
	// toggle dirty
	next, _ := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	m = next.(Model)
	if !m.bulkFilter[worktree.StatusDirty] {
		t.Error("dirty should be toggled on")
	}
	// enter -> bulk confirm with 1 target
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(Model)
	if m.state != stateBulkConfirm {
		t.Errorf("state = %d, want stateBulkConfirm", m.state)
	}
	if len(m.bulkTargets) != 1 || m.bulkTargets[0] != 1 {
		t.Errorf("targets = %v, want [1]", m.bulkTargets)
	}
}

func TestUpdateList_EnterOnNewQuits(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	// index 0: worktree, 1: delete-all, 2: clean-pushed, 3: custom-select, 4: new, 5: quit
	for i := 0; i < 4; i++ {
		next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m = next.(Model)
	}
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(Model)
	if m.Action() != ActionNew {
		t.Errorf("Action = %s, want new", m.Action())
	}
	if cmd == nil {
		t.Error("Enter on [new] should emit tea.Quit cmd")
	}
}

func TestUpdate_PRFetched(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a/.claude/worktrees/x", Status: worktree.StatusPushed},
	})
	got, _ := m.Update(prFetchedMsg{prs: map[string]gh.PRInfo{
		"a": {Number: 1, Title: "t", State: "OPEN"},
	}})
	mm := got.(Model)
	if mm.prUnavailable {
		t.Error("prUnavailable should be false on success")
	}
	if len(mm.prs) != 1 {
		t.Errorf("want 1 pr, got %d", len(mm.prs))
	}
}

func TestUpdate_PRFetchErr_SetsUnavailable(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a/.claude/worktrees/x", Status: worktree.StatusPushed},
	})
	got, _ := m.Update(prFetchErrMsg{err: errors.New("rate limit")})
	mm := got.(Model)
	if !mm.prUnavailable {
		t.Error("prUnavailable should be true after error")
	}
}

func TestUpdateList_QKeyCancels(t *testing.T) {
	m := New(nil)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
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
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return next.(Model)
}

func TestMenu_ResumeEmitsQuit(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = selectFirstWorktree(m)
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
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
	next, _ := m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
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
	next, _ := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
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
	out := m.View().Content
	for _, want := range []string{"feat-y", "local-only", "[r]", "[d]", "[b]"} {
		if !strings.Contains(out, want) {
			t.Errorf("menuView missing %q:\n%s", want, out)
		}
	}
}

func goToDeleteConfirm(m Model) Model {
	m = selectFirstWorktree(m)
	next, _ := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	return next.(Model)
}

func TestDeleteConfirm_YesOnCleanConfirmsWithoutForce(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = goToDeleteConfirm(m)
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
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
	next, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
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
	next, _ := m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
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
	out := m.View().Content
	if !strings.Contains(out, "--force") {
		t.Errorf("deleteConfirmView missing --force marker for dirty:\n%s", out)
	}
}

func TestDeleteConfirmView_HidesForceForClean(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
	})
	m = goToDeleteConfirm(m)
	out := m.View().Content
	if strings.Contains(out, "--force") {
		t.Errorf("deleteConfirmView should not show --force for clean worktree:\n%s", out)
	}
}

func TestDeleteConfirm_Prunable_SetsIsPrunable(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/missing", Branch: "missing", Status: worktree.StatusPrunable},
	})
	m = goToDeleteConfirm(m)
	next, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	mm := next.(Model)
	if mm.Action() != ActionDelete {
		t.Errorf("Action = %s, want delete", mm.Action())
	}
	if !mm.Selection().IsPrunable {
		t.Error("Selection.IsPrunable should be true for prunable row")
	}
	if mm.Selection().ForceDelete {
		t.Error("ForceDelete should be false for prunable (no remove --force)")
	}
}

func TestDeleteConfirmView_PrunableSinglePromptsPrune(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/missing", Branch: "missing", Status: worktree.StatusPrunable},
	})
	m = goToDeleteConfirm(m)
	out := m.View().Content
	if !strings.Contains(out, "git worktree prune") {
		t.Errorf("prunable confirm view must mention git worktree prune:\n%s", out)
	}
	// only one prunable in the list -> short prompt, no enumeration
	if strings.Contains(out, "following") || strings.Contains(out, "以下") {
		t.Errorf("single-prunable view should not enumerate, got:\n%s", out)
	}
}

func TestDeleteConfirmView_PrunableMultipleEnumerates(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/p1", Branch: "p1", Status: worktree.StatusPrunable},
		{Path: "/a/.claude/worktrees/p2", Branch: "p2", Status: worktree.StatusPrunable},
	})
	m = goToDeleteConfirm(m)
	out := m.View().Content
	for _, want := range []string{"git worktree prune", "/a/.claude/worktrees/p1", "/a/.claude/worktrees/p2"} {
		if !strings.Contains(out, want) {
			t.Errorf("multi-prunable confirm view missing %q:\n%s", want, out)
		}
	}
}

func TestUpdate_DeleteAll_WithPrunable_SetsRunPrune(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a", Status: worktree.StatusPushed},
		{Branch: "p", Path: "/p", Status: worktree.StatusPrunable},
	})
	m.bulkTargets = []int{0, 1}
	m.state = stateBulkConfirm
	got, _ := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	mm := got.(Model)
	if mm.Action() != ActionBulkDelete {
		t.Errorf("Action = %s, want bulk-delete", mm.Action())
	}
	b := mm.Bulk()
	if !b.RunPrune {
		t.Error("RunPrune should be true when a prunable target is selected")
	}
	if len(b.Paths) != 1 || b.Paths[0] != "/a" {
		t.Errorf("Paths = %v, want only /a (prunable excluded)", b.Paths)
	}
}

func TestBulkConfirmView_ShowsPruneNote(t *testing.T) {
	m := New([]worktree.Info{
		{Branch: "a", Path: "/a", Status: worktree.StatusPushed},
		{Branch: "p", Path: "/p", Status: worktree.StatusPrunable},
	})
	m.bulkTargets = []int{0, 1}
	m.state = stateBulkConfirm
	out := m.View().Content
	if !strings.Contains(out, "git worktree prune") {
		t.Errorf("bulkConfirmView with prunable target must mention git worktree prune:\n%s", out)
	}
}

func TestView_ListRendersItems(t *testing.T) {
	m := New([]worktree.Info{
		{Path: "/a/.claude/worktrees/feat-x", Branch: "feat-x", Status: worktree.StatusPushed},
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	out := m.View().Content
	if !strings.Contains(out, "feat-x") {
		t.Errorf("View() missing branch name:\n%s", out)
	}
	if !strings.Contains(out, "[new]") || !strings.Contains(out, "[quit]") {
		t.Errorf("View() missing synthetic rows:\n%s", out)
	}
}
