package gh_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tqer39/ccw-cli/internal/gh"
)

type fakeErr struct{ s string }

func (e *fakeErr) Error() string { return e.s }

var errNotFound = &fakeErr{"not found"}

type fakeRunner struct {
	lookErr error
	authOK  bool
	prJSON  string
	prErr   error
}

func (f *fakeRunner) LookPath() error { return f.lookErr }
func (f *fakeRunner) AuthStatus() error {
	if f.authOK {
		return nil
	}
	return &fakeErr{"not authed"}
}
func (f *fakeRunner) PRListJSON() (string, error) { return f.prJSON, f.prErr }

func TestAvailable_FakeRunner(t *testing.T) {
	cases := []struct {
		name    string
		lookErr error
		authOK  bool
		want    bool
	}{
		{"no binary", errNotFound, false, false},
		{"not authed", nil, false, false},
		{"ok", nil, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := &fakeRunner{lookErr: c.lookErr, authOK: c.authOK}
			if got := gh.AvailableWith(r); got != c.want {
				t.Errorf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestPRStatus_Success(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "pr_list.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	r := &fakeRunner{prJSON: string(data)}
	got, err := gh.PRStatusWith(r, []string{
		"shimmying-frolicking-kahan",
		"playful-swashbuckling-ai",
		"unrelated-branch",
	})
	if err != nil {
		t.Fatalf("PRStatusWith: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 matched entries, got %d (%v)", len(got), got)
	}
	if pr := got["shimmying-frolicking-kahan"]; pr.Number != 12 || pr.State != "MERGED" {
		t.Errorf("kahan: got %+v", pr)
	}
	if pr := got["playful-swashbuckling-ai"]; pr.Number != 42 || pr.State != "OPEN" {
		t.Errorf("pirate: got %+v", pr)
	}
}

func TestPRStatus_RunnerError(t *testing.T) {
	r := &fakeRunner{prErr: &fakeErr{"rate limit"}}
	_, err := gh.PRStatusWith(r, []string{"any"})
	if err == nil {
		t.Fatal("want error when runner fails")
	}
}

func TestPRStatusWithTimeout_Success(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "pr_list.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	r := &fakeRunner{prJSON: string(data)}
	got, err := gh.PRStatusWithTimeout(r, 1*time.Second, []string{"shimmying-frolicking-kahan"})
	if err != nil {
		t.Fatalf("PRStatusWithTimeout: %v", err)
	}
	if got["shimmying-frolicking-kahan"].Number != 12 {
		t.Errorf("got = %+v", got)
	}
}

func TestPRStatusWithTimeout_RunnerErrorReturnsError(t *testing.T) {
	r := &fakeRunner{prErr: errors.New("boom")}
	if _, err := gh.PRStatusWithTimeout(r, 1*time.Second, []string{"x"}); err == nil {
		t.Fatal("want error, got nil")
	}
}
