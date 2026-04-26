package listmode

import (
	"errors"
	"testing"
	"time"

	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestBuild_HappyPath(t *testing.T) {
	ts := time.Now()
	wts := []worktree.Info{{
		Path:        "/abs/.claude/worktrees/ccw-x",
		Branch:      "worktree-ccw-x",
		Status:      worktree.StatusPushed,
		AheadCount:  0,
		BehindCount: 0,
		HasSession:  true,
		SessionPath: "/log.jsonl",
		CreatedAt:   &ts,
		LastCommit:  &worktree.CommitInfo{SHA: "abc1234", Subject: "init", Time: ts},
	}}
	prs := map[string]gh.PRInfo{
		"worktree-ccw-x": {Number: 42, Title: "feat", State: "OPEN"},
	}

	b := Builder{
		ListWorktrees: func(string) ([]worktree.Info, error) { return wts, nil },
		ResolveRepo: func(string) (RepoInfo, error) {
			return RepoInfo{Owner: "tqer39", Name: "ccw-cli", DefaultBranch: "main", MainPath: "/abs"}, nil
		},
		FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return prs, nil },
		GhAvailable: func() bool { return true },
	}
	out, warns, err := b.Build("/abs", Options{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("warns = %v, want none", warns)
	}
	if out.Version != 1 {
		t.Errorf("Version = %d", out.Version)
	}
	if len(out.Worktrees) != 1 {
		t.Fatalf("Worktrees len = %d", len(out.Worktrees))
	}
	w := out.Worktrees[0]
	if w.Name != "ccw-x" {
		t.Errorf("Name = %q", w.Name)
	}
	if w.Status != "pushed" {
		t.Errorf("Status = %q", w.Status)
	}
	if w.PR == nil || w.PR.Number != 42 {
		t.Errorf("PR = %+v", w.PR)
	}
	if w.PR.URL != "https://github.com/tqer39/ccw-cli/pull/42" {
		t.Errorf("PR.URL = %q", w.PR.URL)
	}
	if !w.Session.Exists || w.Session.LogPath == nil || *w.Session.LogPath != "/log.jsonl" {
		t.Errorf("Session = %+v", w.Session)
	}
}

func TestBuild_GhUnavailable_PRNullPlusWarning(t *testing.T) {
	b := Builder{
		ListWorktrees: func(string) ([]worktree.Info, error) {
			return []worktree.Info{{Path: "/a/.claude/worktrees/x", Branch: "b", Status: worktree.StatusPushed}}, nil
		},
		ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{Owner: "o", Name: "r", MainPath: "/a"}, nil },
		FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, nil },
		GhAvailable: func() bool { return false },
	}
	out, warns, err := b.Build("/a", Options{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(warns) != 1 {
		t.Errorf("warns = %v, want exactly 1", warns)
	}
	if out.Worktrees[0].PR != nil {
		t.Errorf("PR = %+v, want nil", out.Worktrees[0].PR)
	}
}

func TestBuild_NoPROptionSkipsFetch(t *testing.T) {
	called := false
	b := Builder{
		ListWorktrees: func(string) ([]worktree.Info, error) { return nil, nil },
		ResolveRepo:   func(string) (RepoInfo, error) { return RepoInfo{}, nil },
		FetchPRs: func([]string) (map[string]gh.PRInfo, error) {
			called = true
			return nil, nil
		},
		GhAvailable: func() bool { return true },
	}
	if _, _, err := b.Build("/a", Options{NoPR: true}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if called {
		t.Error("FetchPRs called despite NoPR=true")
	}
}

func TestBuild_NoSessionOptionForcesEmpty(t *testing.T) {
	b := Builder{
		ListWorktrees: func(string) ([]worktree.Info, error) {
			return []worktree.Info{{Path: "/a/.claude/worktrees/x", Branch: "b", Status: worktree.StatusPushed, HasSession: true, SessionPath: "/p"}}, nil
		},
		ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{}, nil },
		FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, nil },
		GhAvailable: func() bool { return true },
	}
	out, _, _ := b.Build("/a", Options{NoSession: true})
	if out.Worktrees[0].Session.Exists {
		t.Error("Session.Exists = true despite NoSession=true")
	}
}

func TestBuild_PRFetchErrorBecomesWarning(t *testing.T) {
	b := Builder{
		ListWorktrees: func(string) ([]worktree.Info, error) {
			return []worktree.Info{{Path: "/a/.claude/worktrees/x", Branch: "b", Status: worktree.StatusPushed}}, nil
		},
		ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{Owner: "o", Name: "r"}, nil },
		FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, errors.New("rate limit") },
		GhAvailable: func() bool { return true },
	}
	out, warns, err := b.Build("/a", Options{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(warns) != 1 {
		t.Errorf("want 1 warning, got %v", warns)
	}
	if out.Worktrees[0].PR != nil {
		t.Errorf("PR not nil on fetch failure")
	}
}

func TestBuild_PrunableSkipsAheadAndCommit(t *testing.T) {
	b := Builder{
		ListWorktrees: func(string) ([]worktree.Info, error) {
			return []worktree.Info{{Path: "/a/.claude/worktrees/p", Branch: "b", Status: worktree.StatusPrunable}}, nil
		},
		ResolveRepo: func(string) (RepoInfo, error) { return RepoInfo{}, nil },
		FetchPRs:    func([]string) (map[string]gh.PRInfo, error) { return nil, nil },
		GhAvailable: func() bool { return true },
	}
	out, _, _ := b.Build("/a", Options{})
	w := out.Worktrees[0]
	if w.Status != "prunable" {
		t.Errorf("Status = %q", w.Status)
	}
	if w.LastCommit != nil || w.CreatedAt != nil {
		t.Errorf("commit/created should be nil for prunable, got %+v %+v", w.LastCommit, w.CreatedAt)
	}
}
