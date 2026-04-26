package listmode

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestOutput_JSONShape(t *testing.T) {
	ts := time.Date(2026, 4, 26, 4, 28, 0, 0, time.UTC)
	logPath := "/log.jsonl"
	out := Output{
		Version: 1,
		Repo: RepoInfo{
			Owner:         "tqer39",
			Name:          "ccw-cli",
			DefaultBranch: "main",
			MainPath:      "/abs",
		},
		Worktrees: []WorktreeEntry{{
			Name:          "ccw-foo",
			Path:          "/abs/.claude/worktrees/ccw-foo",
			Branch:        "worktree-ccw-foo",
			Status:        "pushed",
			Ahead:         0,
			Behind:        0,
			Dirty:         false,
			DefaultBranch: "main",
			CreatedAt:     &ts,
			LastCommit: &CommitInfo{
				SHA:     "9d3dc6e",
				Subject: "feat: x",
				Time:    ts,
			},
			PR: &PRInfo{
				State:  "OPEN",
				Number: 42,
				URL:    "https://github.com/tqer39/ccw-cli/pull/42",
				Title:  "feat: ...",
			},
			Session: SessionInfo{Exists: true, LogPath: &logPath},
		}},
	}

	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(b)
	for _, want := range []string{
		`"version":1`,
		`"owner":"tqer39"`,
		`"default_branch":"main"`,
		`"worktrees":[`,
		`"status":"pushed"`,
		`"ahead":0`,
		`"dirty":false`,
		`"pr":{`,
		`"state":"OPEN"`,
		`"session":{"exists":true,"log_path":"/log.jsonl"}`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("JSON missing %q\nfull: %s", want, s)
		}
	}
}

func TestOutput_EmptyWorktreesIsArrayNotNull(t *testing.T) {
	out := Output{Version: 1, Repo: RepoInfo{}, Worktrees: []WorktreeEntry{}}
	b, _ := json.Marshal(out)
	if !strings.Contains(string(b), `"worktrees":[]`) {
		t.Errorf("want empty array, got %s", string(b))
	}
}

func TestPRInfoNullsCleanly(t *testing.T) {
	w := WorktreeEntry{Name: "x", PR: nil}
	b, _ := json.Marshal(w)
	if !strings.Contains(string(b), `"pr":null`) {
		t.Errorf("want pr:null, got %s", string(b))
	}
}
