// fake claude binary used in resume_flow_test.go.
// Writes os.Args[1:] to $CCW_FAKE_CLAUDE_LOG (newline-separated, with a "---"
// separator between calls) and exits with $CCW_FAKE_CLAUDE_EXIT (default 0).
package main

import (
	"os"
	"strconv"
	"strings"
)

func main() {
	logPath := os.Getenv("CCW_FAKE_CLAUDE_LOG")
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = f.WriteString(strings.Join(os.Args[1:], "\n") + "\n---\n")
			_ = f.Close()
		}
	}
	exit := 0
	if v := os.Getenv("CCW_FAKE_CLAUDE_EXIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			exit = n
		}
	}
	os.Exit(exit)
}
