// Package gh wraps the `gh` CLI for ccw's optional PR display. All exec
// calls go through the Runner interface so tests can inject a fake.
package gh

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// PRInfo is a minimal projection of `gh pr list` used by the picker.
type PRInfo struct {
	Number int
	Title  string
	State  string // "OPEN" | "CLOSED" | "MERGED" | "DRAFT"
}

// Runner abstracts the `gh` binary so tests can swap in a fake.
type Runner interface {
	LookPath() error
	AuthStatus() error
	PRListJSON() (string, error)
}

// DefaultRunner executes the real `gh` binary.
type DefaultRunner struct{}

// LookPath implements Runner by searching PATH for `gh`.
func (DefaultRunner) LookPath() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("look gh: %w", err)
	}
	return nil
}

// AuthStatus implements Runner by running `gh auth status`.
func (DefaultRunner) AuthStatus() error {
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh auth status: %w", err)
	}
	return nil
}

// PRListJSON implements Runner by running `gh pr list --json ...`.
func (DefaultRunner) PRListJSON() (string, error) {
	cmd := exec.Command("gh", "pr", "list",
		"--state", "all",
		"--limit", "200",
		"--json", "number,title,state,headRefName")
	out, err := cmd.Output()
	if err != nil {
		return string(out), fmt.Errorf("gh pr list: %w", err)
	}
	return string(out), nil
}

// Available reports whether gh is installed and authenticated.
func Available() bool { return AvailableWith(DefaultRunner{}) }

// AvailableWith lets callers inject a Runner (primarily for tests).
func AvailableWith(r Runner) bool {
	if err := r.LookPath(); err != nil {
		return false
	}
	return r.AuthStatus() == nil
}

// PRStatus fetches PR info for given branches using the default runner.
func PRStatus(branches []string) (map[string]PRInfo, error) {
	return PRStatusWith(DefaultRunner{}, branches)
}

// PRStatusWith lets callers inject a Runner.
func PRStatusWith(r Runner, branches []string) (map[string]PRInfo, error) {
	out, err := r.PRListJSON()
	if err != nil {
		return nil, fmt.Errorf("pr list: %w", err)
	}
	var raw []struct {
		Number      int    `json:"number"`
		Title       string `json:"title"`
		State       string `json:"state"`
		HeadRefName string `json:"headRefName"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("parse pr list: %w", err)
	}
	want := make(map[string]struct{}, len(branches))
	for _, b := range branches {
		want[b] = struct{}{}
	}
	result := make(map[string]PRInfo)
	for _, e := range raw {
		if _, ok := want[e.HeadRefName]; !ok {
			continue
		}
		result[e.HeadRefName] = PRInfo{Number: e.Number, Title: e.Title, State: e.State}
	}
	return result, nil
}
