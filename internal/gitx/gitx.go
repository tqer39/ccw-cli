// Package gitx wraps common `git` command invocations used throughout ccw.
//
// All helpers accept an optional `dir` (working directory). If non-empty,
// it is passed as `git -C <dir>` so callers need not Chdir.
package gitx

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Run executes `git [-C dir] args...` and discards all output. Returns an
// error wrapping the exec failure when git exits non-zero.
func Run(dir string, args ...string) error {
	cmd := exec.Command("git", withDir(dir, args)...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

// Output executes git and returns trimmed stdout. stderr is discarded.
func Output(dir string, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", withDir(dir, args)...)
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

// OutputSilent is like Output but returns (stdout, err) even on failure,
// so callers can treat non-zero exit as a normal branch.
func OutputSilent(dir string, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", withDir(dir, args)...)
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard
	err := cmd.Run()
	out := strings.TrimRight(buf.String(), "\n")
	if err != nil {
		return out, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

func withDir(dir string, args []string) []string {
	if dir == "" {
		return args
	}
	out := make([]string, 0, len(args)+2)
	out = append(out, "-C", dir)
	return append(out, args...)
}
