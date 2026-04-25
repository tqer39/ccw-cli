package worktree

import (
	"os"
	"path/filepath"
	"strings"
)

// EncodeProjectPath converts an absolute worktree path to the directory name
// Claude Code uses under ~/.claude/projects/. Both '/' and '.' map to '-'.
// This rule is observed from claude's behavior; it is not part of any
// public contract and may change.
func EncodeProjectPath(absPath string) string {
	return strings.NewReplacer("/", "-", ".", "-").Replace(absPath)
}

// HasSession reports whether ~/.claude/projects/<encoded(absPath)>/ contains
// at least one *.jsonl file. Returns false on any error (HOME unset, dir
// missing, read failure) so callers can use it as a UI hint without
// branching on errors.
func HasSession(absPath string) bool {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return false
	}
	dir := filepath.Join(home, ".claude", "projects", EncodeProjectPath(absPath))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			return true
		}
	}
	return false
}
