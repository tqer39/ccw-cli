package gitx

import "testing"

func TestDefaultBranch_FromOriginHEAD(t *testing.T) {
	upstream := initRepo(t)
	mustRun(t, upstream, "git", "commit", "--allow-empty", "-m", "init")
	dir := initRepo(t)
	mustRun(t, dir, "git", "remote", "add", "origin", upstream)
	mustRun(t, dir, "git", "fetch", "origin")
	mustRun(t, dir, "git", "remote", "set-head", "origin", "-a")
	got, err := DefaultBranch(dir)
	if err != nil {
		t.Fatalf("DefaultBranch: %v", err)
	}
	if got != "main" {
		t.Errorf("DefaultBranch = %q, want %q", got, "main")
	}
}

func TestDefaultBranch_FallbackMain(t *testing.T) {
	dir := initRepo(t)
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	got, err := DefaultBranch(dir)
	if err != nil {
		t.Fatalf("DefaultBranch fallback main: %v", err)
	}
	if got != "main" {
		t.Errorf("DefaultBranch = %q, want %q", got, "main")
	}
}

func TestDefaultBranch_FallbackMaster(t *testing.T) {
	dir := initRepo(t)
	// Materialize main with a commit, switch to master, then delete main.
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init-main")
	mustRun(t, dir, "git", "checkout", "-q", "-b", "master")
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	mustRun(t, dir, "git", "branch", "-q", "-D", "main")
	got, err := DefaultBranch(dir)
	if err != nil {
		t.Fatalf("DefaultBranch fallback master: %v", err)
	}
	if got != "master" {
		t.Errorf("DefaultBranch = %q, want %q", got, "master")
	}
}

func TestDefaultBranch_NoBranches(t *testing.T) {
	dir := initRepo(t)
	// Materialize main with a commit, switch to feature, then delete main.
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init-main")
	mustRun(t, dir, "git", "checkout", "-q", "-b", "feature")
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	mustRun(t, dir, "git", "branch", "-q", "-D", "main")
	_, err := DefaultBranch(dir)
	if err == nil {
		t.Fatal("DefaultBranch with no main/master/origin: want error, got nil")
	}
}
