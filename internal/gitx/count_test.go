package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tqer39/ccw-cli/internal/gitx"
)

func TestAheadBehind_NoUpstream(t *testing.T) {
	dir := initRepoWithCommit(t)
	ahead, behind, err := gitx.AheadBehind(dir)
	if err != nil {
		t.Fatalf("AheadBehind: %v", err)
	}
	if ahead != 0 || behind != 0 {
		t.Errorf("want 0/0 without upstream, got %d/%d", ahead, behind)
	}
}

func TestAheadBehind_Ahead(t *testing.T) {
	dir := initRepoWithCommit(t)
	bare := filepath.Join(t.TempDir(), "bare.git")
	mustRun(t, "git", "init", "--bare", bare)
	mustRun(t, "git", "-C", dir, "remote", "add", "origin", bare)
	mustRun(t, "git", "-C", dir, "push", "-u", "origin", "HEAD:main")
	mustRun(t, "git", "-C", dir, "commit", "--allow-empty", "-m", "second")
	mustRun(t, "git", "-C", dir, "commit", "--allow-empty", "-m", "third")

	ahead, behind, err := gitx.AheadBehind(dir)
	if err != nil {
		t.Fatalf("AheadBehind: %v", err)
	}
	if ahead != 2 || behind != 0 {
		t.Errorf("want 2/0, got %d/%d", ahead, behind)
	}
}

func TestDirtyCount_Clean(t *testing.T) {
	dir := initRepoWithCommit(t)
	n, err := gitx.DirtyCount(dir)
	if err != nil {
		t.Fatalf("DirtyCount: %v", err)
	}
	if n != 0 {
		t.Errorf("want 0, got %d", n)
	}
}

func TestDirtyCount_Mixed(t *testing.T) {
	dir := initRepoWithCommit(t)
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}
	n, err := gitx.DirtyCount(dir)
	if err != nil {
		t.Fatalf("DirtyCount: %v", err)
	}
	if n != 2 {
		t.Errorf("want 2 (untracked a, b), got %d", n)
	}
}

func initRepoWithCommit(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	mustRun(t, "git", "init", "-q", "-b", "main", dir)
	mustRun(t, "git", "-C", dir, "config", "user.email", "t@example.com")
	mustRun(t, "git", "-C", dir, "config", "user.name", "t")
	mustRun(t, "git", "-C", dir, "config", "commit.gpgsign", "false")
	mustRun(t, "git", "-C", dir, "commit", "--allow-empty", "-m", "first")
	return dir
}

func mustRun(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}
