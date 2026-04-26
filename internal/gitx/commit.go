package gitx

import (
	"fmt"
	"strings"
	"time"
)

// LastCommit returns short SHA (7 chars), subject line, and author time of HEAD
// at the working tree wt. Errors when the repo has no commits.
func LastCommit(wt string) (string, string, time.Time, error) {
	out, err := Output(wt, "log", "-1", "--no-color", "--format=%h%x1f%s%x1f%aI", "HEAD")
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("last commit: %w", err)
	}
	parts := strings.SplitN(strings.TrimSpace(out), "\x1f", 3)
	if len(parts) != 3 {
		return "", "", time.Time{}, fmt.Errorf("last commit: malformed output %q", out)
	}
	ts, err := time.Parse(time.RFC3339, parts[2])
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("last commit: parse time %q: %w", parts[2], err)
	}
	return parts[0], parts[1], ts, nil
}
