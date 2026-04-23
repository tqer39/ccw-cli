package picker

import (
	"testing"

	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestIcon(t *testing.T) {
	cases := []struct {
		s    worktree.Status
		want string
	}{
		{worktree.StatusPushed, "✅"},
		{worktree.StatusLocalOnly, "⚠"},
		{worktree.StatusDirty, "⛔"},
		{worktree.Status(99), "•"},
	}
	for _, tc := range cases {
		if got := Icon(tc.s); got != tc.want {
			t.Errorf("Icon(%s) = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestActionString(t *testing.T) {
	cases := []struct {
		a    Action
		want string
	}{
		{ActionCancel, "cancel"},
		{ActionResume, "resume"},
		{ActionDelete, "delete"},
		{ActionNew, "new"},
	}
	for _, tc := range cases {
		if got := tc.a.String(); got != tc.want {
			t.Errorf("Action(%d).String() = %q, want %q", tc.a, got, tc.want)
		}
	}
}
