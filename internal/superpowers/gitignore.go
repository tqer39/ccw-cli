package superpowers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tqer39/ccw-cli/internal/ui"
)

const gitignoreBlock = "\n# superpowers workflow artifacts\ndocs/superpowers/\n"

// EnsureGitignore verifies that docs/superpowers/ is ignored by git. If not,
// and in interactive mode, it prompts and appends the canonical block to
// .gitignore. In non-interactive mode it silently leaves the file unchanged,
// matching bash 版 ensure_gitignore().
func EnsureGitignore(in io.Reader, out io.Writer, mainRepo string, interactive bool) error {
	cmd := exec.Command("git", "-C", mainRepo, "check-ignore", "-q", "docs/superpowers/")
	if err := cmd.Run(); err == nil {
		return nil
	}

	_, _ = fmt.Fprintln(out, "⚠ docs/superpowers/ is not ignored by git.")
	_, _ = fmt.Fprintln(out, "The following block will be appended to .gitignore:")
	_, _ = fmt.Fprintln(out, "  # superpowers workflow artifacts")
	_, _ = fmt.Fprintln(out, "  docs/superpowers/")

	if !interactive {
		return nil
	}

	yes, err := ui.PromptYN(in, out, "Add to .gitignore?")
	if err != nil {
		return fmt.Errorf("prompt: %w", err)
	}
	if !yes {
		return nil
	}
	return appendIgnoreBlock(filepath.Join(mainRepo, ".gitignore"))
}

func appendIgnoreBlock(path string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	needsLeadingNewline := len(existing) > 0 && existing[len(existing)-1] != '\n'

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()

	if needsLeadingNewline {
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
	}
	if _, err := f.WriteString(gitignoreBlock); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}
