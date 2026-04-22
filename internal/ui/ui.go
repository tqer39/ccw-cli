// Package ui provides colored CLI output helpers and tool-presence checks.
package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/term"
)

var (
	stdout       io.Writer = os.Stdout
	stderr       io.Writer = os.Stderr
	colorEnabled bool
)

// InitColor evaluates NO_COLOR and the stderr TTY state once. Call from main
// before any Info/Warn/Error/Success/Debug.
func InitColor() {
	if os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
		return
	}
	colorEnabled = term.IsTerminal(int(os.Stderr.Fd()))
}

// SetWriter redirects stdout/stderr output (tests) and disables color.
func SetWriter(out, err io.Writer) {
	stdout = out
	stderr = err
	colorEnabled = false
}

// Info writes to stdout without prefix or color.
func Info(format string, args ...any) {
	_, _ = fmt.Fprintf(stdout, format+"\n", args...)
}

// Warn writes a yellow-prefixed message to stderr.
func Warn(format string, args ...any) {
	write(stderr, "⚠ ", 33, format, args...)
}

// Error writes a red-prefixed message to stderr.
func Error(format string, args ...any) {
	write(stderr, "✖ ", 31, format, args...)
}

// Success writes a green-prefixed message to stderr.
func Success(format string, args ...any) {
	write(stderr, "✓ ", 32, format, args...)
}

// Debug writes a gray-prefixed message to stderr only when CCW_DEBUG=1.
func Debug(format string, args ...any) {
	if os.Getenv("CCW_DEBUG") != "1" {
		return
	}
	write(stderr, "[debug] ", 90, format, args...)
}

// EnsureTool aborts with exit 1 if name is not found in PATH.
func EnsureTool(name, installHint string) {
	if _, err := exec.LookPath(name); err != nil {
		Error("required tool not found: %s. %s", name, installHint)
		os.Exit(1)
	}
}

func write(w io.Writer, prefix string, ansi int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		_, _ = fmt.Fprintf(w, "\x1b[%dm%s%s\x1b[0m\n", ansi, prefix, msg)
		return
	}
	_, _ = fmt.Fprintf(w, "%s%s\n", prefix, msg)
}
