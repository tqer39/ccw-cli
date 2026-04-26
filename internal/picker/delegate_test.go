package picker

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestRenderRow_ResumeBadge(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:       "/repo/.claude/worktrees/foo",
			Branch:     "feature/auth",
			Status:     worktree.StatusPushed,
			HasSession: true,
		},
	}
	got := renderRow(li, 120, true, false)
	if !strings.Contains(got, "[RESUME]") {
		t.Errorf("missing RESUME badge:\n%s", got)
	}
	if !strings.Contains(got, "🌲 foo") {
		t.Errorf("missing tree icon + worktree name '🌲 foo':\n%s", got)
	}
	if !strings.Contains(got, "branch:  feature/auth") {
		t.Errorf("missing branch line:\n%s", got)
	}
	if strings.Contains(got, "path:") {
		t.Errorf("path: line should be removed:\n%s", got)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("got %d lines, want 3:\n%s", len(lines), got)
	}
}

func TestRowDelegateHeight(t *testing.T) {
	if got := (rowDelegate{}).Height(); got != 3 {
		t.Errorf("rowDelegate.Height() = %d, want 3", got)
	}
}

func TestRenderRow_NewBadge(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:       "/repo/.claude/worktrees/bar",
			Branch:     "bar",
			Status:     worktree.StatusLocalOnly,
			HasSession: false,
		},
	}
	got := renderRow(li, 120, true, true)
	if !strings.Contains(got, "[NEW]") {
		t.Errorf("missing NEW badge:\n%s", got)
	}
	if !strings.HasPrefix(got, "> ") {
		t.Errorf("selected row should start with '> ': %q", got[:2])
	}
}

func TestRenderRow_StatusBadgeAndIndicators(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:        "/repo/.claude/worktrees/dirty",
			Branch:      "wip",
			Status:      worktree.StatusDirty,
			AheadCount:  2,
			BehindCount: 1,
			DirtyCount:  5,
		},
	}
	got := renderRow(li, 120, true, false)
	if !strings.Contains(got, "[dirty]") {
		t.Errorf("missing [dirty]:\n%s", got)
	}
	if !strings.Contains(got, "↑2 ↓1 ✎5") {
		t.Errorf("missing indicators:\n%s", got)
	}
}

func TestRenderRow_PRLineWithPR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:   "/repo/.claude/worktrees/foo",
			Branch: "feat/login",
			Status: worktree.StatusPushed,
		},
		pr: &gh.PRInfo{Number: 42, State: "OPEN", Title: "add login page"},
	}
	got := renderRow(li, 120, false, false)
	if !strings.Contains(got, "pr:      ") {
		t.Errorf("missing pr: label:\n%s", got)
	}
	if !strings.Contains(got, "[open]") {
		t.Errorf("missing [open] PR badge:\n%s", got)
	}
	if !strings.Contains(got, "#42") {
		t.Errorf("missing #42:\n%s", got)
	}
}

func TestRenderRow_PRLineNoPR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:   "/repo/.claude/worktrees/lonely",
			Branch: "lonely",
			Status: worktree.StatusLocalOnly,
		},
		pr: nil,
	}
	got := renderRow(li, 120, false, false)
	if !strings.Contains(got, "(no PR)") {
		t.Errorf("missing (no PR) placeholder:\n%s", got)
	}
}

func TestRenderRow_PRUnavailableHidesPRContent(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:   "/repo/.claude/worktrees/n",
			Branch: "n",
			Status: worktree.StatusDirty,
		},
		pr: nil,
	}
	got := renderRow(li, 120, true, false)
	if strings.Contains(got, "(no PR)") {
		t.Errorf("PR placeholder should not appear when prUnavailable:\n%s", got)
	}
	if strings.Contains(got, "#") {
		t.Errorf("PR number should not appear when prUnavailable:\n%s", got)
	}
	// pr line still appears as label, but the cell is empty
	if !strings.Contains(got, "pr:") {
		t.Errorf("pr: label should still appear:\n%s", got)
	}
}

func TestRenderRow_HeaderUsesPipeSeparator(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:       "/repo/.claude/worktrees/foo",
			Branch:     "feat/login",
			Status:     worktree.StatusPushed,
			HasSession: true,
		},
	}
	got := renderRow(li, 200, true, false)
	header := strings.SplitN(got, "\n", 2)[0]
	for _, want := range []string{
		"[RESUME] | 🌲 foo",
		"🌲 foo | [pushed]",
	} {
		if !strings.Contains(header, want) {
			t.Errorf("header missing %q:\n%s", want, header)
		}
	}
	if strings.Contains(header, "·") {
		t.Errorf("header should not contain `·` (replaced by `|`):\n%s", header)
	}
}

func TestRenderRow_HeaderHasNoLargeRightPadding(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:       "/repo/.claude/worktrees/foo",
			Branch:     "feat/login",
			Status:     worktree.StatusPushed,
			HasSession: true,
		},
	}
	got := renderRow(li, 200, true, false)
	header := strings.SplitN(got, "\n", 2)[0]
	if w := lipgloss.Width(header); w > 80 {
		t.Errorf("header visible width %d > 80 at terminal width 200; left/right alignment leaked back in:\n%s", w, header)
	}
}

func TestRenderRow_RightMargin(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	li := listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Path:        "/repo/.claude/worktrees/foo",
			Branch:      "feature/right-margin",
			Status:      worktree.StatusLocalOnly,
			AheadCount:  0,
			BehindCount: 0,
		},
	}
	const width = 80
	const margin = 4
	got := renderRow(li, width, true, false)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	for i, line := range lines {
		if w := lipgloss.Width(line); w > width-margin {
			t.Errorf("line %d has visible width %d > %d (width %d - margin %d):\n%s",
				i, w, width-margin, width, margin, line)
		}
	}
}
