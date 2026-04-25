// Package worktree exposes high-level, ccw-specific worktree queries
// (filtering + status classification) on top of internal/gitx.
package worktree

import (
	"fmt"
	"strings"

	"github.com/tqer39/ccw-cli/internal/gitx"
)

// Status classifies a worktree's publish state.
type Status int

const (
	// StatusPushed means clean, upstream exists, ahead == 0.
	StatusPushed Status = iota
	// StatusLocalOnly means clean, but either no upstream or ahead > 0.
	StatusLocalOnly
	// StatusDirty means the working tree has untracked or modified files.
	StatusDirty
)

// String returns the short lowercase label used in picker UI.
func (s Status) String() string {
	switch s {
	case StatusPushed:
		return "pushed"
	case StatusLocalOnly:
		return "local-only"
	case StatusDirty:
		return "dirty"
	default:
		return "unknown"
	}
}

// Info is a ccw-managed worktree entry with its classified status and
// quantitative indicators (ahead/behind commits, dirty file count).
// AheadCount/BehindCount are meaningful for StatusPushed and StatusLocalOnly.
// DirtyCount is meaningful only when Status == StatusDirty.
// HasSession indicates whether a Claude Code session exists for this worktree.
type Info struct {
	Path        string
	Branch      string
	Status      Status
	AheadCount  int
	BehindCount int
	DirtyCount  int
	HasSession  bool
}

const ccwPathMarker = "/.claude/worktrees/"

// List returns ccw-managed worktrees under mainRepo, each classified.
func List(mainRepo string) ([]Info, error) {
	entries, err := gitx.ListRaw(mainRepo)
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}
	var result []Info
	for _, e := range entries {
		if !strings.Contains(e.Path, ccwPathMarker) {
			continue
		}
		st, err := Classify(e.Path)
		if err != nil {
			return nil, err
		}
		info := Info{Path: e.Path, Branch: e.Branch, Status: st}
		ahead, behind, _ := gitx.AheadBehind(e.Path)
		info.AheadCount = ahead
		info.BehindCount = behind
		if st == StatusDirty {
			n, _ := gitx.DirtyCount(e.Path)
			info.DirtyCount = n
		}
		info.HasSession = HasSession(e.Path)
		result = append(result, info)
	}
	return result, nil
}

// Classify inspects a worktree and returns pushed / local-only / dirty.
// Mirrors bash worktree_flags().
func Classify(wt string) (Status, error) {
	dirty, err := gitx.Output(wt, "status", "--porcelain")
	if err != nil {
		return 0, fmt.Errorf("classify status: %w", err)
	}
	if strings.TrimSpace(dirty) != "" {
		return StatusDirty, nil
	}
	upstream, err := gitx.OutputSilent(wt, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil || strings.TrimSpace(upstream) == "" {
		return StatusLocalOnly, nil
	}
	ahead, err := gitx.OutputSilent(wt, "rev-list", "--count", "@{u}..HEAD")
	if err != nil {
		return StatusLocalOnly, nil
	}
	if strings.TrimSpace(ahead) == "0" {
		return StatusPushed, nil
	}
	return StatusLocalOnly, nil
}

// Remove calls `git worktree remove [--force] path`.
func Remove(mainRepo, path string, force bool) error {
	if err := gitx.RemoveWorktree(mainRepo, path, force); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}
	return nil
}
