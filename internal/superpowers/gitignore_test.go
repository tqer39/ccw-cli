package superpowers

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	runGit("init", "-q", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "test")
	return dir
}

func TestEnsureGitignore_AlreadyIgnored(t *testing.T) {
	repo := setupRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("docs/superpowers/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := EnsureGitignore(strings.NewReader(""), &out, repo, true); err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if string(body) != "docs/superpowers/\n" {
		t.Errorf(".gitignore changed unexpectedly:\n%s", body)
	}
}

func TestEnsureGitignore_NonInteractiveSkips(t *testing.T) {
	repo := setupRepo(t)
	var out bytes.Buffer
	if err := EnsureGitignore(strings.NewReader(""), &out, repo, false); err != nil {
		t.Fatalf("EnsureGitignore non-interactive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf(".gitignore created in non-interactive mode: %v", err)
	}
}

func TestEnsureGitignore_UserAcceptsAppendsBlock(t *testing.T) {
	repo := setupRepo(t)
	var out bytes.Buffer
	if err := EnsureGitignore(strings.NewReader("y\n"), &out, repo, true); err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(body), "# superpowers workflow artifacts") {
		t.Errorf(".gitignore missing comment marker:\n%s", body)
	}
	if !strings.Contains(string(body), "docs/superpowers/") {
		t.Errorf(".gitignore missing path entry:\n%s", body)
	}
}

func TestEnsureGitignore_UserDeclinesNoChange(t *testing.T) {
	repo := setupRepo(t)
	var out bytes.Buffer
	if err := EnsureGitignore(strings.NewReader("n\n"), &out, repo, true); err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf(".gitignore created despite decline: %v", err)
	}
}

func TestEnsureGitignore_PreservesExistingContent(t *testing.T) {
	repo := setupRepo(t)
	existing := "node_modules/\n"
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := EnsureGitignore(strings.NewReader("y\n"), &out, repo, true); err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if !strings.HasPrefix(string(body), existing) {
		t.Errorf(".gitignore prefix changed:\n%s", body)
	}
	if !strings.Contains(string(body), "docs/superpowers/") {
		t.Errorf(".gitignore missing appended entry:\n%s", body)
	}
}

func TestEnsureGitignore_AppendsLeadingNewlineWhenMissing(t *testing.T) {
	repo := setupRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("foo"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := EnsureGitignore(strings.NewReader("y\n"), &out, repo, true); err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if !strings.HasPrefix(string(body), "foo\n") {
		t.Errorf(".gitignore did not preserve leading content + newline:\n%s", body)
	}
}
