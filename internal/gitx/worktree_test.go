package gitx

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParsePorcelain_TwoEntries(t *testing.T) {
	in := strings.Join([]string{
		"worktree /a/main",
		"HEAD abc123",
		"branch refs/heads/main",
		"",
		"worktree /a/.claude/worktrees/wt1",
		"HEAD def456",
		"branch refs/heads/feature",
		"",
	}, "\n")

	got := ParsePorcelain(in)
	want := []WorktreeEntry{
		{Path: "/a/main", Branch: "main"},
		{Path: "/a/.claude/worktrees/wt1", Branch: "feature"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParsePorcelain:\n got  = %+v\n want = %+v", got, want)
	}
}

func TestParsePorcelain_DetachedHEAD(t *testing.T) {
	in := strings.Join([]string{
		"worktree /a/main",
		"HEAD abc123",
		"detached",
		"",
	}, "\n")

	got := ParsePorcelain(in)
	want := []WorktreeEntry{{Path: "/a/main", Branch: ""}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParsePorcelain detached:\n got  = %+v\n want = %+v", got, want)
	}
}

func TestParsePorcelain_EmptyInput(t *testing.T) {
	if got := ParsePorcelain(""); len(got) != 0 {
		t.Errorf("ParsePorcelain empty: got %d entries, want 0", len(got))
	}
}

func TestListRaw_Integration(t *testing.T) {
	main := initRepo(t)
	mustRun(t, main, "git", "commit", "--allow-empty", "-m", "init")
	wt := filepath.Join(main, ".claude", "worktrees", "sample")
	mustRun(t, main, "git", "worktree", "add", "-b", "sample-branch", wt)

	got, err := ListRaw(main)
	if err != nil {
		t.Fatalf("ListRaw: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListRaw len = %d, want 2", len(got))
	}
	var found bool
	for _, e := range got {
		if strings.Contains(e.Path, "/.claude/worktrees/") && e.Branch == "sample-branch" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ccw worktree not found in %+v", got)
	}
}

func TestRemoveWorktree_Integration(t *testing.T) {
	main := initRepo(t)
	mustRun(t, main, "git", "commit", "--allow-empty", "-m", "init")
	wt := filepath.Join(main, ".claude", "worktrees", "tmp")
	mustRun(t, main, "git", "worktree", "add", "-b", "tmp-branch", wt)

	if err := RemoveWorktree(main, wt, false); err != nil {
		t.Fatalf("RemoveWorktree: %v", err)
	}

	list, err := ListRaw(main)
	if err != nil {
		t.Fatalf("ListRaw after remove: %v", err)
	}
	for _, e := range list {
		if e.Path == wt {
			t.Errorf("worktree %q still present after remove", wt)
		}
	}
}
