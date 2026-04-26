// Package listmode produces machine-readable summaries of ccw-managed
// worktrees for the `ccw -L` non-interactive list command.
package listmode

import "time"

// Output is the top-level JSON shape, version-pinned to 1.
type Output struct {
	Version   int             `json:"version"`
	Repo      RepoInfo        `json:"repo"`
	Worktrees []WorktreeEntry `json:"worktrees"`
}

// RepoInfo describes the main repository the listing came from.
type RepoInfo struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	MainPath      string `json:"main_path"`
}

// WorktreeEntry is one ccw-managed worktree.
type WorktreeEntry struct {
	Name          string      `json:"name"`
	Path          string      `json:"path"`
	Branch        string      `json:"branch"`
	Status        string      `json:"status"`
	Ahead         int         `json:"ahead"`
	Behind        int         `json:"behind"`
	Dirty         bool        `json:"dirty"`
	DefaultBranch string      `json:"default_branch"`
	CreatedAt     *time.Time  `json:"created_at"`
	LastCommit    *CommitInfo `json:"last_commit"`
	PR            *PRInfo     `json:"pr"`
	Session       SessionInfo `json:"session"`
}

// CommitInfo describes a worktree's HEAD commit.
type CommitInfo struct {
	SHA     string    `json:"sha"`
	Subject string    `json:"subject"`
	Time    time.Time `json:"time"`
}

// PRInfo is the GitHub pull request associated with a worktree's branch.
type PRInfo struct {
	State  string `json:"state"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

// SessionInfo summarizes Claude Code session presence.
type SessionInfo struct {
	Exists  bool    `json:"exists"`
	LogPath *string `json:"log_path"`
}

// Options control optional data gathering during Build.
type Options struct {
	NoPR      bool
	NoSession bool
}

// Warning is a non-fatal diagnostic emitted during Build (printed to stderr
// by the caller).
type Warning struct {
	Message string
}
