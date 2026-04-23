// Package claude wraps launching the `claude` CLI in ccw-appropriate ways
// (new worktree session vs. resume existing worktree).
package claude

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// BuildNewArgs constructs argv (excluding the program name) for
// `claude --permission-mode auto --worktree [extra...] [-- preamble]`.
func BuildNewArgs(preamble string, extra []string) []string {
	args := make([]string, 0, 3+len(extra)+2)
	args = append(args, "--permission-mode", "auto", "--worktree")
	args = append(args, extra...)
	if preamble != "" {
		args = append(args, "--", preamble)
	}
	return args
}

// BuildResumeArgs constructs argv for `claude --permission-mode auto [extra...]`.
func BuildResumeArgs(extra []string) []string {
	args := make([]string, 0, 2+len(extra))
	args = append(args, "--permission-mode", "auto")
	return append(args, extra...)
}

// LaunchNew execs claude with BuildNewArgs in cwd. Returns claude's exit code
// (0 on success, the child exit code on non-zero exit, -1 on exec error).
func LaunchNew(cwd, preamble string, extra []string) (int, error) {
	return runClaude(cwd, BuildNewArgs(preamble, extra))
}

// Resume execs claude with BuildResumeArgs in cwd.
func Resume(cwd string, extra []string) (int, error) {
	return runClaude(cwd, BuildResumeArgs(extra))
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
