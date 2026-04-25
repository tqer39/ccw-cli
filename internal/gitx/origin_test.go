package gitx

import "testing"

func TestOriginURL_Configured(t *testing.T) {
	dir := initRepo(t)
	mustRun(t, dir, "git", "remote", "add", "origin", "git@github.com:tqer39/ccw-cli.git")
	got, err := OriginURL(dir)
	if err != nil {
		t.Fatalf("OriginURL: %v", err)
	}
	if got != "git@github.com:tqer39/ccw-cli.git" {
		t.Errorf("OriginURL = %q, want %q", got, "git@github.com:tqer39/ccw-cli.git")
	}
}

func TestOriginURL_NotConfigured(t *testing.T) {
	dir := initRepo(t)
	got, err := OriginURL(dir)
	if err != nil {
		t.Fatalf("OriginURL on no-origin repo: want nil error, got %v", err)
	}
	if got != "" {
		t.Errorf("OriginURL = %q, want \"\"", got)
	}
}

func TestParseOriginURL(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		owner     string
		repo      string
		wantError bool
	}{
		{"ssh github", "git@github.com:tqer39/ccw-cli.git", "tqer39", "ccw-cli", false},
		{"ssh github no .git", "git@github.com:tqer39/ccw-cli", "tqer39", "ccw-cli", false},
		{"https github", "https://github.com/tqer39/ccw-cli.git", "tqer39", "ccw-cli", false},
		{"https github no .git", "https://github.com/tqer39/ccw-cli", "tqer39", "ccw-cli", false},
		{"https with trailing slash", "https://github.com/tqer39/ccw-cli/", "tqer39", "ccw-cli", false},
		{"gitlab nested", "https://gitlab.com/group/sub/repo.git", "sub", "repo", false},
		{"ssh gitlab nested", "git@gitlab.com:group/sub/repo.git", "sub", "repo", false},
		{"empty", "", "", "", true},
		{"only host", "git@github.com:", "", "", true},
		{"single segment", "https://example.com/repo.git", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, err := ParseOriginURL(tc.url)
			if tc.wantError {
				if err == nil {
					t.Fatalf("ParseOriginURL(%q) want error, got owner=%q repo=%q", tc.url, owner, repo)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseOriginURL(%q) unexpected error: %v", tc.url, err)
			}
			if owner != tc.owner || repo != tc.repo {
				t.Errorf("ParseOriginURL(%q) = (%q, %q), want (%q, %q)", tc.url, owner, repo, tc.owner, tc.repo)
			}
		})
	}
}
