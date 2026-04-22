package gitx

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRequireRepo_Success(t *testing.T) {
	dir := initRepo(t)
	if err := RequireRepo(dir); err != nil {
		t.Fatalf("RequireRepo: %v", err)
	}
}

func TestRequireRepo_FailsOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	if err := RequireRepo(dir); err == nil {
		t.Fatal("RequireRepo outside repo: want error")
	}
}

func TestResolveMainRepo_FromRepoRoot(t *testing.T) {
	dir := initRepo(t)
	got, err := ResolveMainRepo(dir)
	if err != nil {
		t.Fatalf("ResolveMainRepo: %v", err)
	}
	if !samePath(t, got, dir) {
		t.Errorf("ResolveMainRepo = %q, want equivalent of %q", got, dir)
	}
}

func TestResolveMainRepo_FromWorktree(t *testing.T) {
	main := initRepo(t)
	mustRun(t, main, "git", "commit", "--allow-empty", "-m", "init")
	wt := filepath.Join(main, ".claude", "worktrees", "sample")
	mustRun(t, main, "git", "worktree", "add", "-b", "sample-branch", wt)

	got, err := ResolveMainRepo(wt)
	if err != nil {
		t.Fatalf("ResolveMainRepo from worktree: %v", err)
	}
	if !samePath(t, got, main) {
		t.Errorf("ResolveMainRepo from %q = %q, want %q", wt, got, main)
	}
}

func samePath(t *testing.T, a, b string) bool {
	t.Helper()
	ea, err := filepath.EvalSymlinks(a)
	if err != nil {
		ea = a
	}
	eb, err := filepath.EvalSymlinks(b)
	if err != nil {
		eb = b
	}
	return strings.EqualFold(ea, eb) || ea == eb
}
