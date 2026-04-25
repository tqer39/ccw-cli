package picker

import (
	"strings"
	"testing"

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
	if !strings.Contains(got, "foo") {
		t.Errorf("missing worktree name foo:\n%s", got)
	}
	if !strings.Contains(got, "branch:  feature/auth") {
		t.Errorf("missing branch line:\n%s", got)
	}
	if !strings.Contains(got, "path:    /repo/.claude/worktrees/foo") {
		t.Errorf("missing path line:\n%s", got)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 4 {
		t.Errorf("got %d lines, want 4:\n%s", len(lines), got)
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
