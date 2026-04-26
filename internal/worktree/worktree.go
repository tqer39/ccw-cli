// Package worktree exposes high-level, ccw-specific worktree queries
// (filtering + status classification) on top of internal/gitx.
package worktree

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	// StatusPrunable means the working directory is gone but git still
	// keeps admin files for it. Cleared by `git worktree prune`.
	StatusPrunable
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
	case StatusPrunable:
		return "prunable"
	default:
		return "unknown"
	}
}

// FilterAll is the wildcard token accepted by ParseStatusFilter / --status.
const FilterAll = "all"

// FilterAllowed returns the labels accepted by --status, in display order.
// "all" is the wildcard; the rest match Status.String() for the user-facing
// statuses (prunable is internal and not selectable).
func FilterAllowed() []string {
	return []string{
		FilterAll,
		StatusPushed.String(),
		StatusLocalOnly.String(),
		StatusDirty.String(),
	}
}

// ParseStatusFilter maps a --status value to a one-element filter set.
// Returns (nil, true) for "all" or "" (no filter), (nil, false) for unknown.
func ParseStatusFilter(s string) (map[Status]bool, bool) {
	switch s {
	case "", FilterAll:
		return nil, true
	case StatusPushed.String():
		return map[Status]bool{StatusPushed: true}, true
	case StatusLocalOnly.String():
		return map[Status]bool{StatusLocalOnly: true}, true
	case StatusDirty.String():
		return map[Status]bool{StatusDirty: true}, true
	}
	return nil, false
}

// CommitInfo summarizes the HEAD commit of a worktree.
type CommitInfo struct {
	SHA     string
	Subject string
	Time    time.Time
}

// Info is a ccw-managed worktree entry with its classified status and
// quantitative indicators (ahead/behind commits, dirty file count).
// AheadCount/BehindCount are meaningful for StatusPushed and StatusLocalOnly.
// DirtyCount is meaningful only when Status == StatusDirty.
// HasSession indicates whether a Claude Code session exists for this worktree.
// CreatedAt / LastCommit / SessionPath are populated for non-prunable entries
// when retrieval succeeds; otherwise nil / empty.
type Info struct {
	Path        string
	Branch      string
	Status      Status
	AheadCount  int
	BehindCount int
	DirtyCount  int
	HasSession  bool
	CreatedAt   *time.Time
	LastCommit  *CommitInfo
	SessionPath string
}

// Indicators formats the ahead/behind/dirty counts for display.
// Format: "↑<ahead> ↓<behind>" plus " ✎<dirty>" when Status is dirty
// (the dirty suffix is omitted when DirtyCount is 0).
// Callers wanting a Prunable-specific label should branch before calling.
func (i Info) Indicators() string {
	out := fmt.Sprintf("↑%d ↓%d", i.AheadCount, i.BehindCount)
	if i.Status == StatusDirty && i.DirtyCount > 0 {
		out += fmt.Sprintf(" ✎%d", i.DirtyCount)
	}
	return out
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
		if e.Prunable {
			result = append(result, Info{
				Path:   e.Path,
				Branch: e.Branch,
				Status: StatusPrunable,
			})
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
		if info.HasSession {
			info.SessionPath = SessionLogPath(e.Path)
		}
		if st, err := os.Stat(e.Path); err == nil {
			t := st.ModTime()
			info.CreatedAt = &t
		}
		if sha, subject, ts, err := gitx.LastCommit(e.Path); err == nil {
			info.LastCommit = &CommitInfo{SHA: sha, Subject: subject, Time: ts}
		}
		result = append(result, info)
	}
	return result, nil
}

// Prune cleans up admin files for prunable worktrees attached to mainRepo.
// Wraps `git -C mainRepo worktree prune`.
func Prune(mainRepo string) error {
	if err := gitx.Prune(mainRepo); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}
	return nil
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
