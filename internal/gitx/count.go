package gitx

import (
	"fmt"
	"strconv"
	"strings"
)

// AheadBehind returns (ahead, behind) commits for HEAD against its upstream.
// Returns (0, 0, nil) when no upstream is configured.
func AheadBehind(dir string) (int, int, error) {
	out, err := OutputSilent(dir, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if err != nil {
		return 0, 0, nil
	}
	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("ahead-behind: unexpected output %q", out)
	}
	behind, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("ahead-behind: parse behind %q: %w", fields[0], err)
	}
	ahead, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("ahead-behind: parse ahead %q: %w", fields[1], err)
	}
	return ahead, behind, nil
}

// DirtyCount returns the number of entries in `git status --porcelain`.
func DirtyCount(dir string) (int, error) {
	out, err := Output(dir, "status", "--porcelain")
	if err != nil {
		return 0, fmt.Errorf("dirty-count: %w", err)
	}
	out = strings.TrimRight(out, "\n")
	if out == "" {
		return 0, nil
	}
	return strings.Count(out, "\n") + 1, nil
}
