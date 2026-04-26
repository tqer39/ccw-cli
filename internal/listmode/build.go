package listmode

import (
	"fmt"
	"path/filepath"

	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// Builder bundles the dependencies Build needs. Tests inject fakes; production
// callers wire ListWorktrees / ResolveRepo / FetchPRs / GhAvailable to the
// real packages.
type Builder struct {
	ListWorktrees func(mainRepo string) ([]worktree.Info, error)
	ResolveRepo   func(mainRepo string) (RepoInfo, error)
	FetchPRs      func(branches []string) (map[string]gh.PRInfo, error)
	GhAvailable   func() bool
}

// Build assembles an *Output from the given main repo. Returns warnings
// for non-fatal degradations (gh missing, PR fetch failures, etc).
func (b Builder) Build(mainRepo string, opts Options) (*Output, []Warning, error) {
	repo, err := b.ResolveRepo(mainRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve repo: %w", err)
	}

	infos, err := b.ListWorktrees(mainRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("list worktrees: %w", err)
	}

	var warns []Warning
	prs := map[string]gh.PRInfo{}
	if !opts.NoPR {
		switch {
		case !b.GhAvailable():
			warns = append(warns, Warning{Message: "gh not available, PR info disabled"})
		default:
			branches := make([]string, 0, len(infos))
			for _, info := range infos {
				if info.Branch != "" {
					branches = append(branches, info.Branch)
				}
			}
			fetched, err := b.FetchPRs(branches)
			if err != nil {
				warns = append(warns, Warning{Message: fmt.Sprintf("gh pr fetch failed: %v", err)})
			} else {
				prs = fetched
			}
		}
	}

	out := &Output{
		Version:   1,
		Repo:      repo,
		Worktrees: make([]WorktreeEntry, 0, len(infos)),
	}
	for _, info := range infos {
		out.Worktrees = append(out.Worktrees, buildEntry(info, repo, prs, opts))
	}
	return out, warns, nil
}

func buildEntry(info worktree.Info, repo RepoInfo, prs map[string]gh.PRInfo, opts Options) WorktreeEntry {
	entry := WorktreeEntry{
		Name:          filepath.Base(info.Path),
		Path:          info.Path,
		Branch:        info.Branch,
		Status:        info.Status.String(),
		Ahead:         info.AheadCount,
		Behind:        info.BehindCount,
		Dirty:         info.Status == worktree.StatusDirty,
		DefaultBranch: repo.DefaultBranch,
		CreatedAt:     info.CreatedAt,
	}
	if info.LastCommit != nil {
		entry.LastCommit = &CommitInfo{
			SHA:     info.LastCommit.SHA,
			Subject: info.LastCommit.Subject,
			Time:    info.LastCommit.Time,
		}
	}
	if pr, ok := prs[info.Branch]; ok && info.Branch != "" {
		entry.PR = &PRInfo{
			State:  pr.State,
			Number: pr.Number,
			URL:    prURL(repo, pr.Number),
			Title:  pr.Title,
		}
	}
	if !opts.NoSession && info.HasSession {
		path := info.SessionPath
		entry.Session = SessionInfo{Exists: true, LogPath: &path}
	}
	return entry
}

func prURL(repo RepoInfo, number int) string {
	if repo.Owner == "" || repo.Name == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", repo.Owner, repo.Name, number)
}
