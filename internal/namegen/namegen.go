// Package namegen generates timestamp-based worktree / Claude Code session names
// of the form "ccw-<owner>-<repo>-<yymmdd>-<hhmmss>" using local time.
package namegen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

// origin / worktree-list / clock hooks are package-level vars so tests can
// substitute fakes without spinning up a real repo or a real clock.
var (
	originURLFn    = gitx.OriginURL
	worktreeListFn = gitx.ListRaw
	nowFn          = time.Now
)

// maxCollisionSuffix bounds numeric suffixes attempted before giving up.
const maxCollisionSuffix = 99

// timestampLayout is Go's reference time formatted as yymmdd-hhmmss.
const timestampLayout = "060102-150405"

// buildName composes "ccw-<owner>-<repo>-<tail>" with normalization,
// suffixing "-2", "-3", ... when the candidate is in `taken`.
func buildName(owner, repo, tail string, taken map[string]bool) (string, error) {
	o := normalize(owner)
	r := normalize(repo)
	t := normalize(tail)
	if o == "" {
		return "", fmt.Errorf("buildName: owner is empty after normalization (input %q)", owner)
	}
	if r == "" {
		return "", fmt.Errorf("buildName: repo is empty after normalization (input %q)", repo)
	}
	if t == "" {
		return "", fmt.Errorf("buildName: tail is empty after normalization (input %q)", tail)
	}
	base := fmt.Sprintf("ccw-%s-%s-%s", o, r, t)
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

// Generate returns a worktree name of the form
// "ccw-<owner>-<repo>-<yymmdd>-<hhmmss>" for the repository at mainRepo.
// The timestamp is formatted from time.Now() in local time.
// When `origin` is unset, owner becomes "local" and repo is the basename
// of mainRepo. Numeric "-N" suffixes are appended on collision (cap: 99).
func Generate(mainRepo string) (string, error) {
	owner, repo, err := resolveOwnerRepo(mainRepo)
	if err != nil {
		return "", err
	}
	ts := nowFn().Format(timestampLayout)
	taken, err := takenNames(mainRepo)
	if err != nil {
		return "", err
	}
	return buildName(owner, repo, ts, taken)
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

// takenNames returns names already in use, by union of:
//   - directory entries under <mainRepo>/.claude/worktrees/
//   - basenames of paths returned by `git worktree list --porcelain`
//
// A registered git worktree without a matching .claude/worktrees entry (e.g.
// added outside ccw, or after manual cleanup) still collides with a fresh name.
func takenNames(mainRepo string) (map[string]bool, error) {
	out := map[string]bool{}
	dir := filepath.Join(mainRepo, ".claude", "worktrees")
	entries, err := os.ReadDir(dir)
	switch {
	case err == nil:
		for _, e := range entries {
			if e.IsDir() {
				out[e.Name()] = true
			}
		}
	case !os.IsNotExist(err):
		return nil, fmt.Errorf("read worktrees dir: %w", err)
	}
	wts, err := worktreeListFn(mainRepo)
	if err != nil {
		return nil, fmt.Errorf("list git worktrees: %w", err)
	}
	for _, wt := range wts {
		out[filepath.Base(wt.Path)] = true
	}
	return out, nil
}
