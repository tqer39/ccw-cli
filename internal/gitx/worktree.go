package gitx

import (
	"fmt"
	"strings"
)

// WorktreeEntry represents one record from `git worktree list --porcelain`.
type WorktreeEntry struct {
	Path     string
	Branch   string // without "refs/heads/" prefix; empty for detached HEAD
	Prunable bool   // true when git tagged this entry with `prunable`
}

// ListRaw returns every worktree attached to mainRepo. Caller is responsible
// for filtering (e.g. ccw-managed paths only).
func ListRaw(mainRepo string) ([]WorktreeEntry, error) {
	out, err := Output(mainRepo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	return ParsePorcelain(out), nil
}

// ParsePorcelain parses `git worktree list --porcelain` output.
func ParsePorcelain(s string) []WorktreeEntry {
	var entries []WorktreeEntry
	var cur WorktreeEntry
	flush := func() {
		if cur.Path != "" {
			entries = append(entries, cur)
		}
		cur = WorktreeEntry{}
	}
	for _, line := range strings.Split(s, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			cur.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(
				strings.TrimPrefix(line, "branch "),
				"refs/heads/",
			)
		case line == "prunable" || strings.HasPrefix(line, "prunable "):
			cur.Prunable = true
		case line == "":
			flush()
		}
	}
	flush()
	return entries
}

// RemoveWorktree removes a worktree. If force is true, passes --force so that
// dirty worktrees can still be removed (matches the bash delete_worktree).
func RemoveWorktree(mainRepo, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	if err := Run(mainRepo, args...); err != nil {
		return fmt.Errorf("git worktree remove: %w", err)
	}
	return nil
}

// Prune calls `git -C mainRepo worktree prune`. This removes admin files for
// worktrees whose working directory has been deleted (i.e. those flagged as
// `prunable` in `git worktree list --porcelain`).
func Prune(mainRepo string) error {
	if err := Run(mainRepo, "worktree", "prune"); err != nil {
		return fmt.Errorf("git worktree prune: %w", err)
	}
	return nil
}
