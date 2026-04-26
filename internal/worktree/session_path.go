package worktree

import (
	"os"
	"path/filepath"
	"strings"
)

// SessionLogPath returns the absolute path to the first *.jsonl session log
// for absPath under ~/.claude/projects/<encoded>/, or "" if none.
// Mirrors HasSession's lookup so callers can use both consistently.
func SessionLogPath(absPath string) string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(absPath))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}
