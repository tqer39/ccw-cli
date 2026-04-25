package gitx

import (
	"fmt"
	"strings"
)

// ParseOriginURL extracts (owner, repo) from a git remote URL.
// Supports SSH (git@host:owner/repo[.git]) and HTTPS (https://host/owner/repo[.git]) forms.
// Nested path segments (e.g. GitLab subgroups) collapse to the last two segments.
// Returns an error for empty, malformed, or single-segment paths.
func ParseOriginURL(rawURL string) (string, string, error) {
	url := strings.TrimSpace(rawURL)
	if url == "" {
		return "", "", fmt.Errorf("empty origin url")
	}
	var path string
	switch {
	case strings.HasPrefix(url, "git@"):
		idx := strings.Index(url, ":")
		if idx < 0 {
			return "", "", fmt.Errorf("malformed ssh url: %q", rawURL)
		}
		path = url[idx+1:]
	case strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "ssh://") || strings.HasPrefix(url, "git://"):
		idx := strings.Index(url, "://")
		rest := url[idx+3:]
		slash := strings.Index(rest, "/")
		if slash < 0 {
			return "", "", fmt.Errorf("malformed url: %q", rawURL)
		}
		path = rest[slash+1:]
	default:
		return "", "", fmt.Errorf("unsupported url scheme: %q", rawURL)
	}
	path = strings.TrimSuffix(strings.TrimSuffix(strings.Trim(path, "/"), ".git"), "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" || parts[len(parts)-1] == "" {
		return "", "", fmt.Errorf("origin url has fewer than 2 path segments: %q", rawURL)
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	return owner, repo, nil
}

// OriginURL returns the URL of the `origin` remote, or "" when not configured.
// The empty/no-origin case is treated as a normal branch (no error).
func OriginURL(mainRepo string) (string, error) {
	out, err := OutputSilent(mainRepo, "remote", "get-url", "origin")
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(out), nil
}
