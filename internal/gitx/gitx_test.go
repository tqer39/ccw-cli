package gitx

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	mustRun(t, dir, "git", "init", "-q", "-b", "main")
	mustRun(t, dir, "git", "config", "user.email", "test@example.com")
	mustRun(t, dir, "git", "config", "user.name", "test")
	mustRun(t, dir, "git", "config", "commit.gpgsign", "false")
	return dir
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestRun_Success(t *testing.T) {
	dir := initRepo(t)
	if err := Run(dir, "status"); err != nil {
		t.Fatalf("Run status: %v", err)
	}
}

func TestRun_FailsOnNonRepo(t *testing.T) {
	dir := t.TempDir()
	if err := Run(dir, "status"); err == nil {
		t.Fatal("Run status in non-repo: want error, got nil")
	}
}

func TestOutput_ReturnsTrimmedStdout(t *testing.T) {
	dir := initRepo(t)
	got, err := Output(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if got != "true" {
		t.Errorf("Output = %q, want %q", got, "true")
	}
}

func TestOutput_ErrorOnFailure(t *testing.T) {
	dir := t.TempDir()
	_, err := Output(dir, "rev-parse", "--is-inside-work-tree")
	if err == nil {
		t.Fatal("Output in non-repo: want error, got nil")
	}
}

func TestOutputSilent_SwallowsStderr(t *testing.T) {
	dir := initRepo(t)
	got, err := OutputSilent(dir, "rev-parse", "--abbrev-ref", "@{u}")
	if err == nil {
		t.Fatal("OutputSilent @{u}: want error, got nil")
	}
	if strings.Contains(got, "fatal") {
		t.Errorf("OutputSilent stdout contains stderr text: %q", got)
	}
}

func TestOutputSilent_DirAppliedAsCArg(t *testing.T) {
	dir := initRepo(t)
	wd := t.TempDir()
	t.Chdir(wd)
	got, err := Output(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	want := filepath.Base(dir)
	if !strings.HasSuffix(got, want) {
		t.Errorf("toplevel = %q, want suffix %q", got, want)
	}
}
