// Package claude wraps launching the `claude` CLI in ccw-appropriate ways
// (new worktree session vs. continue existing worktree).
package claude

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// BuildNewArgs constructs argv (excluding the program name) for
// `claude --permission-mode auto --worktree <name> -n <name> [extra...] [-- <preamble>]`.
// When preamble is non-empty it is appended after a `--` separator so claude
// treats it as the first user prompt.
func BuildNewArgs(name, preamble string, extra []string) []string {
	return buildLaunchArgs(true, name, preamble, extra)
}

// BuildInWorktreeArgs is BuildNewArgs without `--worktree`. Use when cwd is
// already an existing worktree, since passing `--worktree <name>` from inside
// a worktree risks a name-collision error against the existing git registration.
func BuildInWorktreeArgs(name, preamble string, extra []string) []string {
	return buildLaunchArgs(false, name, preamble, extra)
}

func buildLaunchArgs(includeWorktree bool, name, preamble string, extra []string) []string {
	args := make([]string, 0, 6+len(extra)+2)
	args = append(args, "--permission-mode", "auto")
	if includeWorktree {
		args = append(args, "--worktree", name)
	}
	args = append(args, "-n", name)
	args = append(args, extra...)
	if preamble != "" {
		args = append(args, "--", preamble)
	}
	return args
}

// BuildContinueArgs constructs argv for `claude --permission-mode auto --continue [extra...]`.
func BuildContinueArgs(extra []string) []string {
	args := make([]string, 0, 3+len(extra))
	args = append(args, "--permission-mode", "auto", "--continue")
	return append(args, extra...)
}

// LaunchNew execs claude with BuildNewArgs in cwd. Returns claude's exit code
// (0 on success, the child exit code on non-zero exit, -1 on exec error).
func LaunchNew(cwd, name, preamble string, extra []string) (int, error) {
	return runClaude(cwd, BuildNewArgs(name, preamble, extra))
}

// LaunchInWorktree execs claude with BuildInWorktreeArgs in cwd (no `--worktree`).
func LaunchInWorktree(cwd, name, preamble string, extra []string) (int, error) {
	return runClaude(cwd, BuildInWorktreeArgs(name, preamble, extra))
}

// Continue execs claude with BuildContinueArgs in cwd.
func Continue(cwd string, extra []string) (int, error) {
	return runClaude(cwd, BuildContinueArgs(extra))
}

func runClaude(cwd string, args []string) (int, error) {
	cmd := exec.Command("claude", args...)
	cmd.Dir = cwd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return -1, fmt.Errorf("run claude: %w", err)
}
