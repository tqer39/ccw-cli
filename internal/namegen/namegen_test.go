package namegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
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
		shorthash string
		taken     map[string]bool
		want      string
		wantError bool
	}{
		{
			name:  "no collision",
			owner: "tqer39", repo: "ccw-cli", shorthash: "a3f2b1",
			taken: map[string]bool{},
			want:  "ccw-tqer39-ccw-cli-a3f2b1",
		},
		{
			name:  "one collision",
			owner: "tqer39", repo: "ccw-cli", shorthash: "a3f2b1",
			taken: map[string]bool{"ccw-tqer39-ccw-cli-a3f2b1": true},
			want:  "ccw-tqer39-ccw-cli-a3f2b1-2",
		},
		{
			name:  "two collisions",
			owner: "tqer39", repo: "ccw-cli", shorthash: "a3f2b1",
			taken: map[string]bool{
				"ccw-tqer39-ccw-cli-a3f2b1":   true,
				"ccw-tqer39-ccw-cli-a3f2b1-2": true,
			},
			want: "ccw-tqer39-ccw-cli-a3f2b1-3",
		},
		{
			name:  "normalization applied",
			owner: "Anthropic", repo: "Claude.Code", shorthash: "9F8E7D",
			taken: map[string]bool{},
			want:  "ccw-anthropic-claude-code-9f8e7d",
		},
		{
			name:  "empty owner errors",
			owner: "", repo: "ccw-cli", shorthash: "a3f2b1",
			taken:     map[string]bool{},
			wantError: true,
		},
		{
			name:  "empty repo errors",
			owner: "tqer39", repo: "", shorthash: "a3f2b1",
			taken:     map[string]bool{},
			wantError: true,
		},
		{
			name:  "empty shorthash errors",
			owner: "tqer39", repo: "ccw-cli", shorthash: "",
			taken:     map[string]bool{},
			wantError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildName(tc.owner, tc.repo, tc.shorthash, tc.taken)
			if tc.wantError {
				if err == nil {
					t.Fatalf("buildName(%q,%q,%q) want error, got %q", tc.owner, tc.repo, tc.shorthash, got)
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
	base := "ccw-x-y-aaaaaa"
	taken[base] = true
	for i := 2; i <= 99; i++ {
		taken[base+"-"+strconv.Itoa(i)] = true
	}
	if _, err := buildName("x", "y", "aaaaaa", taken); err == nil {
		t.Fatal("buildName at 99-collision cap: want error, got nil")
	}
}

type fakes struct {
	origin       string
	branch       string
	branchError  bool
	shorthash    string
	shorthashErr bool
}

func withFakes(t *testing.T, f fakes) {
	t.Helper()
	origOrigin := originURLFn
	origBranch := defaultBranchFn
	origHash := shortHashFn
	t.Cleanup(func() {
		originURLFn = origOrigin
		defaultBranchFn = origBranch
		shortHashFn = origHash
	})
	originURLFn = func(string) (string, error) { return f.origin, nil }
	defaultBranchFn = func(string) (string, error) {
		if f.branchError {
			return "", fmt.Errorf("fake: no default branch")
		}
		return f.branch, nil
	}
	shortHashFn = func(string, string, int) (string, error) {
		if f.shorthashErr {
			return "", fmt.Errorf("fake: no commits")
		}
		return f.shorthash, nil
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
		origin:    "git@github.com:tqer39/ccw-cli.git",
		branch:    "main",
		shorthash: "a3f2b1",
	})
	got, err := Generate(t.TempDir())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "ccw-tqer39-ccw-cli-a3f2b1" {
		t.Errorf("Generate = %q, want %q", got, "ccw-tqer39-ccw-cli-a3f2b1")
	}
}

func TestGenerate_NoOriginFallback(t *testing.T) {
	withFakes(t, fakes{
		origin:    "",
		branch:    "main",
		shorthash: "a3f2b1",
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
	if got != "ccw-local-myrepo-a3f2b1" {
		t.Errorf("Generate = %q, want %q", got, "ccw-local-myrepo-a3f2b1")
	}
}

func TestGenerate_DefaultBranchError(t *testing.T) {
	withFakes(t, fakes{
		origin:      "git@github.com:tqer39/ccw-cli.git",
		branchError: true,
	})
	if _, err := Generate(t.TempDir()); err == nil {
		t.Fatal("Generate with default-branch error: want error, got nil")
	}
}

func TestGenerate_CollisionWithExistingDir(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, repo, ".claude/worktrees/ccw-tqer39-ccw-cli-a3f2b1")
	withFakes(t, fakes{
		origin:    "git@github.com:tqer39/ccw-cli.git",
		branch:    "main",
		shorthash: "a3f2b1",
	})
	got, err := Generate(repo)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "ccw-tqer39-ccw-cli-a3f2b1-2" {
		t.Errorf("Generate = %q, want %q", got, "ccw-tqer39-ccw-cli-a3f2b1-2")
	}
}
