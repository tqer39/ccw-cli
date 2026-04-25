package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T, target, out string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", out, target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", target, err, output)
	}
}

func setupFakeEnv(t *testing.T) (binDir, logPath, home string) {
	t.Helper()
	binDir = t.TempDir()
	home = t.TempDir()
	buildBinary(t, "./fakes/fake_claude", filepath.Join(binDir, "claude"))
	buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))
	logPath = filepath.Join(t.TempDir(), "claude.log")
	return
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
		{"commit", "--allow-empty", "-q", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func runCcw(t *testing.T, binDir, repo, log, home string, args ...string) string {
	t.Helper()
	cmd := exec.Command(filepath.Join(binDir, "ccw"), args...)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"HOME="+home,
		"CCW_FAKE_CLAUDE_LOG="+log,
		"NO_COLOR=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ccw %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func readLog(t *testing.T, p string) []string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n---\n")
}

func indexOf(s []string, target string) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}

func TestResumeFlow_NewWorktreePassesNameToBoth(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	binDir, log, home := setupFakeEnv(t)
	repo := t.TempDir()
	initRepo(t, repo)

	_ = runCcw(t, binDir, repo, log, home, "-n")

	calls := readLog(t, log)
	if len(calls) < 1 {
		t.Fatalf("expected at least 1 claude call, got %d", len(calls))
	}
	first := calls[0]
	if !strings.Contains(first, "--worktree\n") {
		t.Errorf("first call missing --worktree:\n%s", first)
	}
	if !strings.Contains(first, "\n-n\n") {
		t.Errorf("first call missing -n:\n%s", first)
	}
	args := strings.Split(first, "\n")
	idxWT := indexOf(args, "--worktree")
	idxN := indexOf(args, "-n")
	if idxWT < 0 || idxN < 0 || idxWT+1 >= len(args) || idxN+1 >= len(args) {
		t.Fatalf("malformed args:\n%s", first)
	}
	if args[idxWT+1] != args[idxN+1] {
		t.Errorf("--worktree %q != -n %q", args[idxWT+1], args[idxN+1])
	}
}
