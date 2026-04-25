package gitx

import (
	"fmt"
	"strings"
)

// DefaultBranch returns the canonical default branch name for the repo at mainRepo.
// Resolution order:
//  1. refs/remotes/origin/HEAD (e.g. "refs/remotes/origin/main") — strip prefix
//  2. local branch "main"
//  3. local branch "master"
//
// Returns an error when none of the above exist.
func DefaultBranch(mainRepo string) (string, error) {
	if out, err := OutputSilent(mainRepo, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		s := strings.TrimSpace(out)
		if idx := strings.LastIndex(s, "/"); idx >= 0 && idx < len(s)-1 {
			return s[idx+1:], nil
		}
	}
	for _, name := range []string{"main", "master"} {
		if _, err := OutputSilent(mainRepo, "rev-parse", "--verify", "--quiet", "refs/heads/"+name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("no default branch found (origin/HEAD, main, master all unset)")
}

// ShortHash returns the trimmed output of `git rev-parse --short=<length> <ref>`.
func ShortHash(mainRepo, ref string, length int) (string, error) {
	out, err := Output(mainRepo, "rev-parse", fmt.Sprintf("--short=%d", length), ref)
	if err != nil {
		return "", fmt.Errorf("short hash %s: %w", ref, err)
	}
	return strings.TrimSpace(out), nil
}
