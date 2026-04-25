package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeProjectPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/Users/foo/repo/.claude/worktrees/bar", "-Users-foo-repo--claude-worktrees-bar"},
		{"/a.b/c", "-a-b-c"},
		{"/", "-"},
	}
	for _, tc := range cases {
		if got := EncodeProjectPath(tc.in); got != tc.want {
			t.Errorf("EncodeProjectPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHasSession_True(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	wt := "/Users/foo/repo/.claude/worktrees/bar"
	dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "abc.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !HasSession(wt) {
		t.Error("HasSession() = false, want true")
	}
}

func TestHasSession_FalseWhenNoJsonl(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	wt := "/Users/foo/repo/.claude/worktrees/bar"
	dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if HasSession(wt) {
		t.Error("HasSession() = true, want false (no .jsonl)")
	}
}

func TestHasSession_FalseWhenDirMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if HasSession("/nonexistent/path") {
		t.Error("HasSession() = true, want false (dir missing)")
	}
}

func TestHasSession_FalseWhenHomeUnset(t *testing.T) {
	t.Setenv("HOME", "")
	if HasSession("/Users/foo/repo/.claude/worktrees/bar") {
		t.Error("HasSession() = true, want false (HOME empty)")
	}
}
