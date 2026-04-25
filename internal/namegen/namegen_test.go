package namegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/tqer39/ccw-cli/internal/gitx"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Anthropic", "anthropic"},
		{"My Org", "my-org"},
		{"_underscore_", "underscore"},
		{"--double--dash--", "double-dash"},
		{"a..b..c", "a-b-c"},
		{"", ""},
		{"a", "a"},
		{"123", "123"},
		{"日本語repo", "repo"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := normalize(tc.in)
			if got != tc.want {
				t.Errorf("normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildName(t *testing.T) {
	cases := []struct {
		name      string
		owner     string
		repo      string
		tail      string
		taken     map[string]bool
		want      string
		wantError bool
	}{
		{
			name:  "no collision",
			owner: "tqer39", repo: "ccw-cli", tail: "260426-143055",
			taken: map[string]bool{},
			want:  "ccw-tqer39-ccw-cli-260426-143055",
		},
		{
			name:  "one collision",
			owner: "tqer39", repo: "ccw-cli", tail: "260426-143055",
			taken: map[string]bool{"ccw-tqer39-ccw-cli-260426-143055": true},
			want:  "ccw-tqer39-ccw-cli-260426-143055-2",
		},
		{
			name:  "two collisions",
			owner: "tqer39", repo: "ccw-cli", tail: "260426-143055",
			taken: map[string]bool{
				"ccw-tqer39-ccw-cli-260426-143055":   true,
				"ccw-tqer39-ccw-cli-260426-143055-2": true,
			},
			want: "ccw-tqer39-ccw-cli-260426-143055-3",
		},
		{
			name:  "normalization applied",
			owner: "Anthropic", repo: "Claude.Code", tail: "260426-143055",
			taken: map[string]bool{},
			want:  "ccw-anthropic-claude-code-260426-143055",
		},
		{
			name:  "empty owner errors",
			owner: "", repo: "ccw-cli", tail: "260426-143055",
			taken:     map[string]bool{},
			wantError: true,
		},
		{
			name:  "empty repo errors",
			owner: "tqer39", repo: "", tail: "260426-143055",
			taken:     map[string]bool{},
			wantError: true,
		},
		{
			name:  "empty tail errors",
			owner: "tqer39", repo: "ccw-cli", tail: "",
			taken:     map[string]bool{},
			wantError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildName(tc.owner, tc.repo, tc.tail, tc.taken)
			if tc.wantError {
				if err == nil {
					t.Fatalf("buildName(%q,%q,%q) want error, got %q", tc.owner, tc.repo, tc.tail, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildName: %v", err)
			}
			if got != tc.want {
				t.Errorf("buildName = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildName_ManyCollisions(t *testing.T) {
	taken := map[string]bool{}
	base := "ccw-x-y-260426-143055"
	taken[base] = true
	for i := 2; i <= 99; i++ {
		taken[base+"-"+strconv.Itoa(i)] = true
	}
	if _, err := buildName("x", "y", "260426-143055", taken); err == nil {
		t.Fatal("buildName at 99-collision cap: want error, got nil")
	}
}

type fakes struct {
	origin    string
	originErr bool
	now       time.Time
	worktrees []gitx.WorktreeEntry
}

// testNow / testTimestamp are the canonical fixed clock used across Generate tests.
var testNow = time.Date(2026, 4, 26, 14, 30, 55, 0, time.Local)

const testTimestamp = "260426-143055"

func withFakes(t *testing.T, f fakes) {
	t.Helper()
	origOrigin := originURLFn
	origList := worktreeListFn
	origNow := nowFn
	t.Cleanup(func() {
		originURLFn = origOrigin
		worktreeListFn = origList
		nowFn = origNow
	})
	originURLFn = func(string) (string, error) {
		if f.originErr {
			return "", fmt.Errorf("fake: origin url error")
		}
		return f.origin, nil
	}
	worktreeListFn = func(string) ([]gitx.WorktreeEntry, error) { return f.worktrees, nil }
	if !f.now.IsZero() {
		nowFn = func() time.Time { return f.now }
	} else {
		nowFn = func() time.Time { return testNow }
	}
}

func mustMkdir(t *testing.T, root, rel string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
}

func TestGenerate_HappyPath(t *testing.T) {
	withFakes(t, fakes{
		origin: "git@github.com:tqer39/ccw-cli.git",
		now:    testNow,
	})
	got, err := Generate(t.TempDir())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "ccw-tqer39-ccw-cli-" + testTimestamp
	if got != want {
		t.Errorf("Generate = %q, want %q", got, want)
	}
}

func TestGenerate_NoOriginFallback(t *testing.T) {
	withFakes(t, fakes{
		origin: "",
		now:    testNow,
	})
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "myrepo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	got, err := Generate(repoPath)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "ccw-local-myrepo-" + testTimestamp
	if got != want {
		t.Errorf("Generate = %q, want %q", got, want)
	}
}

func TestGenerate_OriginURLError(t *testing.T) {
	withFakes(t, fakes{
		originErr: true,
		now:       testNow,
	})
	if _, err := Generate(t.TempDir()); err == nil {
		t.Fatal("Generate with origin-url error: want error, got nil")
	}
}

func TestGenerate_CollisionWithExistingDir(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, repo, ".claude/worktrees/ccw-tqer39-ccw-cli-"+testTimestamp)
	withFakes(t, fakes{
		origin: "git@github.com:tqer39/ccw-cli.git",
		now:    testNow,
	})
	got, err := Generate(repo)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "ccw-tqer39-ccw-cli-" + testTimestamp + "-2"
	if got != want {
		t.Errorf("Generate = %q, want %q", got, want)
	}
}

// TestGenerate_CollisionWithGitWorktree exercises the spec rule that names
// registered with `git worktree list` count as taken even when no matching
// .claude/worktrees directory exists.
func TestGenerate_CollisionWithGitWorktree(t *testing.T) {
	repo := t.TempDir()
	withFakes(t, fakes{
		origin: "git@github.com:tqer39/ccw-cli.git",
		now:    testNow,
		worktrees: []gitx.WorktreeEntry{
			{Path: "/tmp/elsewhere/ccw-tqer39-ccw-cli-" + testTimestamp},
		},
	})
	got, err := Generate(repo)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "ccw-tqer39-ccw-cli-" + testTimestamp + "-2"
	if got != want {
		t.Errorf("Generate = %q, want %q", got, want)
	}
}
