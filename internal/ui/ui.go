// Package ui provides colored CLI output helpers and tool-presence checks.
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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

// PromptYN writes question + " [y/N]: " to out and reads one line from in.
// Returns true for y / yes (case-insensitive); false for anything else or EOF.
func PromptYN(in io.Reader, out io.Writer, question string) (bool, error) {
	_, _ = fmt.Fprintf(out, "%s [y/N]: ", question)
	line, err := readLine(in)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("prompt yn: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// PromptChoice writes question + " " to out and reads one line. Returns the
// first rune of the trimmed lowercased answer if it is in valid; otherwise
// an error.
func PromptChoice(in io.Reader, out io.Writer, question string, valid []rune) (rune, error) {
	_, _ = fmt.Fprintf(out, "%s ", question)
	line, err := readLine(in)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("prompt choice: %w", err)
	}
	trimmed := strings.ToLower(strings.TrimSpace(line))
	if trimmed == "" {
		return 0, errors.New("empty choice")
	}
	got := []rune(trimmed)[0]
	for _, v := range valid {
		if got == v {
			return got, nil
		}
	}
	return 0, fmt.Errorf("invalid choice: %q", string(got))
}

// IsInteractive returns true if f is a terminal.
func IsInteractive(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

func readLine(in io.Reader) (string, error) {
	r := bufio.NewReader(in)
	s, err := r.ReadString('\n')
	if err != nil {
		return s, fmt.Errorf("read line: %w", err)
	}
	return s, nil
}

func write(w io.Writer, prefix string, ansi int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		_, _ = fmt.Fprintf(w, "\x1b[%dm%s%s\x1b[0m\n", ansi, prefix, msg)
		return
	}
	_, _ = fmt.Fprintf(w, "%s%s\n", prefix, msg)
}
