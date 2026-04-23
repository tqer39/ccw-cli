package gitx

import (
	"fmt"
	"path/filepath"
	"strings"
)

// RequireRepo returns nil if cwd is inside a git working tree.
func RequireRepo(cwd string) error {
	out, err := OutputSilent(cwd, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("not a git repository: %s", cwd)
	}
	if strings.TrimSpace(out) != "true" {
		return fmt.Errorf("not a git working tree: %s", cwd)
	}
	return nil
}

// ResolveMainRepo returns the absolute, symlink-resolved path to the main
// repository root for cwd. Works when cwd is a worktree.
func ResolveMainRepo(cwd string) (string, error) {
	common, err := Output(cwd, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("resolve main repo: %w", err)
	}
	abs := common
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(cwd, abs)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve main repo symlinks: %w", err)
	}
	return filepath.Dir(resolved), nil
}

// SetOriginHead runs `git remote set-head origin -a`. Typical callers ignore
// the error because the command is best-effort.
func SetOriginHead(cwd string) error {
	_, err := OutputSilent(cwd, "remote", "set-head", "origin", "-a")
	return err
}
