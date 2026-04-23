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
	if !strings.Contains(row, "#12") || !strings.Contains(row, "[merged]") {
		t.Errorf("missing PR badge/number:\n%s", row)
	}
}

func TestRenderRow_ContainsArrowAndPRBadge(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	row := renderRow(listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Branch: "feat/login",
			Path:   "/tmp/x",
			Status: worktree.StatusPushed,
		},
		pr: &gh.PRInfo{Number: 42, State: "OPEN", Title: "add login page"},
	}, 120, false, false)
	if !strings.Contains(row, "->") {
		t.Errorf("want arrow separator `->` in NO_COLOR mode, got:\n%s", row)
	}
	if !strings.Contains(row, "[open]") {
		t.Errorf("want PR state badge [open], got:\n%s", row)
	}
	if !strings.Contains(row, "#42") || !strings.Contains(row, "add login page") {
		t.Errorf("want PR number + title, got:\n%s", row)
	}
	if strings.Count(row, "open") != 1 {
		t.Errorf("state label should appear exactly once, got:\n%s", row)
	}
}

func TestRenderRow_ArrowOmittedWhenPRUnavailable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	row := renderRow(listItem{
		tag: tagWorktree,
		wt: &worktree.Info{
			Branch: "nebula",
			Path:   "/tmp/n",
			Status: worktree.StatusDirty,
		},
		pr: nil,
	}, 120, true, false)
	if strings.Contains(row, "->") {
		t.Errorf("arrow should be omitted when prUnavailable, got:\n%s", row)
	}
}

func TestRenderRow_ArrowWithNoPRPlaceholder(t *testing.T) {
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
	if !strings.Contains(row, "->") {
		t.Errorf("arrow should appear even when PR is absent, got:\n%s", row)
	}
	if !strings.Contains(row, "(no PR)") {
		t.Errorf("want (no PR) placeholder, got:\n%s", row)
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
