// Package picker renders the ccw worktree picker (bubbletea) and returns the
// user's decision as a pure (Action, Selection) value for the caller to act on.
package picker

import (
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// Action is the user's choice returned by Run.
type Action int

// Action values. ActionCancel is the zero value so an uninitialized result
// means "no action taken".
const (
	ActionCancel Action = iota
	ActionResume
	ActionDelete
	ActionNew
)

// String returns the lowercase name of the action.
func (a Action) String() string {
	switch a {
	case ActionResume:
		return "resume"
	case ActionDelete:
		return "delete"
	case ActionNew:
		return "new"
	default:
		return "cancel"
	}
}

// Selection identifies the worktree the user picked.
type Selection struct {
	Path        string
	Branch      string
	Status      worktree.Status
	ForceDelete bool
}

// Icon maps a worktree.Status to a one-rune glyph displayed in the list.
func Icon(s worktree.Status) string {
	switch s {
	case worktree.StatusPushed:
		return "✅"
	case worktree.StatusLocalOnly:
		return "⚠"
	case worktree.StatusDirty:
		return "⛔"
	default:
		return "•"
	}
}
