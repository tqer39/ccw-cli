package tests

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestListMode_EmptyRepoEmitsHeaderOnly(t *testing.T) {
	binDir := t.TempDir()
	buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

	repo := t.TempDir()
	initRepo(t, repo)

	cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "--no-pr", "--no-session")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run ccw -L: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "NAME") {
		t.Errorf("expected header, got: %s", out)
	}
}

func TestListMode_JSONShape(t *testing.T) {
	binDir := t.TempDir()
	buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

	repo := t.TempDir()
	initRepo(t, repo)

	cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "--json", "--no-pr", "--no-session")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run ccw -L --json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v\n%s", err, out)
	}
	if v, _ := parsed["version"].(float64); v != 1 {
		t.Errorf("version = %v", parsed["version"])
	}
	if _, ok := parsed["repo"]; !ok {
		t.Error("missing repo key")
	}
	if _, ok := parsed["worktrees"]; !ok {
		t.Error("missing worktrees key")
	}
}

func TestListMode_DirOverridesCwd(t *testing.T) {
	binDir := t.TempDir()
	buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

	repo := t.TempDir()
	initRepo(t, repo)

	cwd := t.TempDir() // not a repo
	cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "-d", repo, "--no-pr", "--no-session")
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run ccw -L -d: %v\n%s", err, out)
	}
}

func TestListMode_DirInvalidExits1(t *testing.T) {
	binDir := t.TempDir()
	buildBinary(t, "../cmd/ccw", filepath.Join(binDir, "ccw"))

	cmd := exec.Command(filepath.Join(binDir, "ccw"), "-L", "-d", "/nonexistent-dir-xyz")
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("want non-zero exit, got success\n%s", out)
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 1 {
			t.Errorf("exit code = %d, want 1", exitErr.ExitCode())
		}
	}
}
