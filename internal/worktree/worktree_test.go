package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initMainRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q", "-b", "main")
	run(t, dir, "git", "config", "user.email", "test@example.com")
	run(t, dir, "git", "config", "user.name", "test")
	run(t, dir, "git", "config", "commit.gpgsign", "false")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func addWorktree(t *testing.T, main, name string) string {
	t.Helper()
	wt := filepath.Join(main, ".claude", "worktrees", name)
	run(t, main, "git", "worktree", "add", "-b", name+"-branch", wt)
	return wt
}

func TestStatus_String(t *testing.T) {
	cases := []struct {
		s    Status
		want string
	}{
		{StatusPushed, "pushed"},
		{StatusLocalOnly, "local-only"},
		{StatusDirty, "dirty"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestClassify_LocalOnlyWhenNoUpstream(t *testing.T) {
	main := initMainRepo(t)
	wt := addWorktree(t, main, "a")

	got, err := Classify(wt)
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if got != StatusLocalOnly {
		t.Errorf("Classify no upstream = %s, want local-only", got)
	}
}

func TestClassify_DirtyWhenUntracked(t *testing.T) {
	main := initMainRepo(t)
	wt := addWorktree(t, main, "b")
	path := filepath.Join(wt, "dirty.txt")
	if err := writeFile(path, "x"); err != nil {
		t.Fatal(err)
	}

	got, err := Classify(wt)
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if got != StatusDirty {
		t.Errorf("Classify dirty = %s, want dirty", got)
	}
}

func TestList_FiltersCcwManagedOnly(t *testing.T) {
	main := initMainRepo(t)
	addWorktree(t, main, "c")
	notCcw := filepath.Join(main, "..", "other")
	run(t, main, "git", "worktree", "add", "-b", "other-branch", notCcw)

	got, err := List(main)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("List len = %d, want 1, entries=%+v", len(got), got)
	}
	if !strings.Contains(got[0].Path, "/.claude/worktrees/") {
		t.Errorf("List returned non-ccw path: %q", got[0].Path)
	}
	if got[0].Branch != "c-branch" {
		t.Errorf("List branch = %q, want c-branch", got[0].Branch)
	}
}

func TestList_FillsCounts(t *testing.T) {
	main := initMainRepo(t)
	addWorktree(t, main, "counts")

	got, err := List(main)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("want at least one worktree")
	}
	w := got[0]
	// No upstream and a clean tree, so ahead/behind/dirty should all be 0.
	if w.AheadCount != 0 || w.BehindCount != 0 {
		t.Errorf("want 0/0 without upstream, got %d/%d", w.AheadCount, w.BehindCount)
	}
	if w.DirtyCount != 0 {
		t.Errorf("clean worktree should have DirtyCount=0, got %d", w.DirtyCount)
	}
}

func TestRemove_Integration(t *testing.T) {
	main := initMainRepo(t)
	wt := addWorktree(t, main, "d")

	if err := Remove(main, wt, false); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	list, err := List(main)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range list {
		if e.Path == wt {
			t.Errorf("worktree %q still present after Remove", wt)
		}
	}
}

func writeFile(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

func TestList_PopulatesHasSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	main := initMainRepo(t)
	wt := addWorktree(t, main, "alpha")

	// Resolve symlinks so the path matches what gitx.ListRaw returns
	// (on macOS, /var/folders/... resolves to /private/var/folders/...).
	wtResolved, err := filepath.EvalSymlinks(wt)
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wtResolved))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "x.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	infos, err := List(main)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, in := range infos {
		if in.Path == wtResolved {
			if !in.HasSession {
				t.Errorf("Info.HasSession = false, want true for %s", wtResolved)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("worktree %s not in List() output", wtResolved)
	}
}
