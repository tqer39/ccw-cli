package gitx

import (
	"testing"
	"time"
)

func TestLastCommit_HappyPath(t *testing.T) {
	dir := initRepo(t)
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "first commit")

	sha, subject, ts, err := LastCommit(dir)
	if err != nil {
		t.Fatalf("LastCommit: %v", err)
	}
	if len(sha) != 7 {
		t.Errorf("sha length = %d, want 7 (got %q)", len(sha), sha)
	}
	if subject != "first commit" {
		t.Errorf("subject = %q, want %q", subject, "first commit")
	}
	if ts.IsZero() {
		t.Errorf("time is zero")
	}
	if time.Since(ts) > time.Minute {
		t.Errorf("time too old: %v", ts)
	}
}

func TestLastCommit_EmptyRepo(t *testing.T) {
	dir := initRepo(t)
	if _, _, _, err := LastCommit(dir); err == nil {
		t.Fatal("LastCommit on empty repo: want error, got nil")
	}
}

func TestLastCommit_SubjectWithSpaces(t *testing.T) {
	dir := initRepo(t)
	mustRun(t, dir, "git", "commit", "--allow-empty", "-m", "feat(x): add multi word subject")

	_, subject, _, err := LastCommit(dir)
	if err != nil {
		t.Fatalf("LastCommit: %v", err)
	}
	if subject != "feat(x): add multi word subject" {
		t.Errorf("subject = %q, want full multi-word", subject)
	}
}
