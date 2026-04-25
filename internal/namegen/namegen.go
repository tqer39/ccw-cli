// Package namegen generates deterministic worktree / Claude Code session names
// of the form "ccw-<owner>-<repo>-<shorthash6>".
package namegen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tqer39/ccw-cli/internal/gitx"
)

// nonSlugRE matches anything outside [a-z0-9-]. Used by normalize.
var nonSlugRE = regexp.MustCompile(`[^a-z0-9-]+`)

// dashRunRE matches runs of two or more dashes. Used by normalize.
var dashRunRE = regexp.MustCompile(`-{2,}`)

// normalize returns a slug-safe lowercase form of s: ASCII-only, [a-z0-9-]+,
// with consecutive dashes collapsed and leading/trailing dashes trimmed.
// Callers (e.g. ParseOriginURL) strip `.git` before calling.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = nonSlugRE.ReplaceAllString(s, "-")
	s = dashRunRE.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// origin / branch / shorthash hooks are package-level vars so tests can
// substitute fakes without spinning up a real repo.
var (
	originURLFn     = gitx.OriginURL
	defaultBranchFn = gitx.DefaultBranch
	shortHashFn     = gitx.ShortHash
)

// maxCollisionSuffix bounds numeric suffixes attempted before giving up.
const maxCollisionSuffix = 99

// buildName composes "ccw-<owner>-<repo>-<shorthash>" with normalization,
// suffixing "-2", "-3", ... when the candidate is in `taken`.
func buildName(owner, repo, shorthash string, taken map[string]bool) (string, error) {
	o := normalize(owner)
	r := normalize(repo)
	h := normalize(shorthash)
	if o == "" {
		return "", fmt.Errorf("buildName: owner is empty after normalization (input %q)", owner)
	}
	if r == "" {
		return "", fmt.Errorf("buildName: repo is empty after normalization (input %q)", repo)
	}
	if h == "" {
		return "", fmt.Errorf("buildName: shorthash is empty after normalization (input %q)", shorthash)
	}
	base := fmt.Sprintf("ccw-%s-%s-%s", o, r, h)
	if !taken[base] {
		return base, nil
	}
	for i := 2; i <= maxCollisionSuffix; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !taken[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("buildName: %d collisions for %q, giving up", maxCollisionSuffix, base)
}

// Generate returns a deterministic worktree name of the form
// "ccw-<owner>-<repo>-<shorthash6>" for the repository at mainRepo.
// When `origin` is unset, owner becomes "local" and repo is the basename
// of mainRepo. Numeric "-N" suffixes are appended on collision (cap: 99).
func Generate(mainRepo string) (string, error) {
	owner, repo, err := resolveOwnerRepo(mainRepo)
	if err != nil {
		return "", err
	}
	branch, err := defaultBranchFn(mainRepo)
	if err != nil {
		return "", fmt.Errorf("default branch: %w", err)
	}
	shorthash, err := shortHashFn(mainRepo, branch, 6)
	if err != nil {
		return "", fmt.Errorf("short hash: %w", err)
	}
	taken, err := takenNames(mainRepo)
	if err != nil {
		return "", err
	}
	return buildName(owner, repo, shorthash, taken)
}

func resolveOwnerRepo(mainRepo string) (string, string, error) {
	url, err := originURLFn(mainRepo)
	if err != nil {
		return "", "", fmt.Errorf("origin url: %w", err)
	}
	if url == "" {
		return "local", filepath.Base(mainRepo), nil
	}
	owner, repo, err := gitx.ParseOriginURL(url)
	if err != nil {
		return "", "", fmt.Errorf("parse origin url: %w", err)
	}
	return owner, repo, nil
}

// takenNames returns the set of worktree directory names already present
// under <mainRepo>/.claude/worktrees/. Missing dir is treated as empty set.
func takenNames(mainRepo string) (map[string]bool, error) {
	dir := filepath.Join(mainRepo, ".claude", "worktrees")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, fmt.Errorf("read worktrees dir: %w", err)
	}
	out := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			out[e.Name()] = true
		}
	}
	return out, nil
}
