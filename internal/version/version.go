// Package version exposes build-time version metadata injected via -ldflags.
package version

import "fmt"

var (
	// Version is the release tag (e.g. "v0.1.0"). Set via -ldflags.
	Version = "dev"
	// Commit is the git commit SHA. Set via -ldflags.
	Commit = "none"
	// Date is the build timestamp. Set via -ldflags.
	Date = "unknown"
)

// String returns a human-readable one-line description of the build.
func String() string {
	return fmt.Sprintf("ccw %s (commit: %s, built: %s)", Version, Commit, Date)
}
