package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionLogPath_FoundReturnsFirst(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	wt := "/Users/foo/repo/.claude/worktrees/bar"
	dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(wt))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "abc123.jsonl")
	if err := os.WriteFile(logPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := SessionLogPath(wt); got != logPath {
		t.Errorf("SessionLogPath = %q, want %q", got, logPath)
	}
}

func TestSessionLogPath_NotFoundReturnsEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if got := SessionLogPath("/nonexistent"); got != "" {
		t.Errorf("SessionLogPath = %q, want empty", got)
	}
}

func TestSessionLogPath_HomeUnsetReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", "")
	if got := SessionLogPath("/x"); got != "" {
		t.Errorf("SessionLogPath = %q, want empty", got)
	}
}
