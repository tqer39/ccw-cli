package picker

import (
	"strings"
	"testing"

	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestRenderRow_PushedNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	row := renderRow(listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Branch:      "kahan",
			Path:        "/tmp/x",
			Status:      worktree.StatusPushed,
			AheadCount:  0,
			BehindCount: 0,
		},
		pr: &gh.PRInfo{Number: 12, State: "MERGED", Title: "Add picker mod"},
	}, 120, false, false)
	if !strings.Contains(row, "[pushed]") {
		t.Errorf("want [pushed] badge, got:\n%s", row)
	}
	if !strings.Contains(row, "kahan") || !strings.Contains(row, "↑0 ↓0") {
		t.Errorf("missing branch/counts:\n%s", row)
	}
	if !strings.Contains(row, "#12") || !strings.Contains(row, "merged") {
		t.Errorf("missing PR info:\n%s", row)
	}
}

func TestRenderRow_DirtyPRUnavailable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	row := renderRow(listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Branch:     "nebula",
			Path:       "/tmp/n",
			Status:     worktree.StatusDirty,
			AheadCount: 1,
			DirtyCount: 5,
		},
		pr: nil,
	}, 120, true, false)
	if !strings.Contains(row, "[dirty]") {
		t.Errorf("want [dirty]:\n%s", row)
	}
	if !strings.Contains(row, "✎5") {
		t.Errorf("want ✎5:\n%s", row)
	}
	if strings.Contains(row, "#") || strings.Contains(row, "no PR") {
		t.Errorf("PR column should be omitted when prUnavailable:\n%s", row)
	}
}

func TestRenderRow_NoPRShowsPlaceholder(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	row := renderRow(listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Branch: "lonely",
			Path:   "/tmp/l",
			Status: worktree.StatusLocalOnly,
		},
		pr: nil,
	}, 120, false, false)
	if !strings.Contains(row, "(no PR)") {
		t.Errorf("want (no PR) marker when PR col enabled but empty:\n%s", row)
	}
}
