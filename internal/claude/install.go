package claude

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/tqer39/ccw-cli/internal/ui"
)

// lookPath / installers are package-level variables so tests can override.
var (
	lookPath = exec.LookPath

	npmInstaller = func() error {
		cmd := exec.Command("npm", "install", "-g", "@anthropic-ai/claude-code")
		wireStdio(cmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm install: %w", err)
		}
		return nil
	}

	brewInstaller = func() error {
		cmd := exec.Command("brew", "install", "claude-code")
		wireStdio(cmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew install: %w", err)
		}
		return nil
	}
)

// EnsureInstalled returns nil if `claude` is on PATH. If missing, and in
// interactive mode, prompts to install via npm or brew.
func EnsureInstalled(in io.Reader, out io.Writer, interactive bool) error {
	if _, err := lookPath("claude"); err == nil {
		return nil
	}
	if !interactive {
		return errors.New("claude not found (non-interactive)")
	}

	_, _ = fmt.Fprintln(out, "⚠ missing dependency: claude (Claude Code CLI)")
	_, _ = fmt.Fprintln(out, "Choose installer:")
	_, _ = fmt.Fprintln(out, "  [1] npm  (npm install -g @anthropic-ai/claude-code)")
	_, _ = fmt.Fprintln(out, "  [2] brew (brew install claude-code)")
	_, _ = fmt.Fprintln(out, "  [q] quit (install manually and rerun)")

	choice, err := ui.PromptChoice(in, out, "Select [1/2/q]:", []rune{'1', '2', 'q'})
	if err != nil {
		return fmt.Errorf("prompt: %w", err)
	}
	switch choice {
	case '1':
		if _, err := lookPath("npm"); err != nil {
			return errors.New("npm not found; install Node.js first: https://nodejs.org/")
		}
		if err := npmInstaller(); err != nil {
			return fmt.Errorf("npm install failed: %w", err)
		}
		return nil
	case '2':
		if _, err := lookPath("brew"); err != nil {
			return errors.New("brew not found; install Homebrew first: https://brew.sh/")
		}
		if err := brewInstaller(); err != nil {
			return fmt.Errorf("brew install failed: %w", err)
		}
		return nil
	}
	return errors.New("cancelled by user")
}

func wireStdio(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}
