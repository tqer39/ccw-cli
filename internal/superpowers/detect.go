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

// DetectInstalled returns true if ~/.claude/plugins/cache/*/superpowers is a
// directory. home is taken as an explicit parameter for testability.
func DetectInstalled(home string) (bool, error) {
	pattern := filepath.Join(home, ".claude", "plugins", "cache", "*", "superpowers")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, fmt.Errorf("glob superpowers: %w", err)
	}
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err == nil && fi.IsDir() {
			return true, nil
		}
	}
	return false, nil
}

// installRunner is overridable in tests.
var installRunner = func() error {
	cmd := exec.Command("claude", "plugin", "install", "claude-plugins-official/superpowers")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude plugin install: %w", err)
	}
	return nil
}

// EnsureInstalled returns nil if superpowers is detected under home; otherwise
// in interactive mode it prompts the user to run the plugin installer.
func EnsureInstalled(in io.Reader, out io.Writer, home string, interactive bool) error {
	ok, err := DetectInstalled(home)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	_, _ = fmt.Fprintln(out, "⚠ missing dependency: superpowers plugin (required for -s)")
	_, _ = fmt.Fprintln(out, "The following command will install it:")
	_, _ = fmt.Fprintln(out, "  claude plugin install claude-plugins-official/superpowers")
	_, _ = fmt.Fprintln(out, "(reference: https://docs.claude.com/en/docs/claude-code/plugins )")

	if !interactive {
		return errors.New("superpowers plugin not installed (non-interactive)")
	}

	yes, err := ui.PromptYN(in, out, "Run now?")
	if err != nil {
		return fmt.Errorf("prompt: %w", err)
	}
	if !yes {
		return errors.New("cancelled by user")
	}
	if err := installRunner(); err != nil {
		return fmt.Errorf("plugin install failed: %w", err)
	}
	return nil
}
