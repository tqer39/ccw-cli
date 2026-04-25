package picker

import (
	"reflect"
	"testing"

	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestSelectByStatus(t *testing.T) {
	infos := []worktree.Info{
		{Branch: "a", Status: worktree.StatusPushed},
		{Branch: "b", Status: worktree.StatusLocalOnly},
		{Branch: "c", Status: worktree.StatusDirty},
	}
	cases := []struct {
		name   string
		filter map[worktree.Status]bool
		want   []int
	}{
		{"all", nil, []int{0, 1, 2}},
		{"pushed", map[worktree.Status]bool{worktree.StatusPushed: true}, []int{0}},
		{"dirty+local", map[worktree.Status]bool{
			worktree.StatusDirty: true, worktree.StatusLocalOnly: true,
		}, []int{1, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SelectByStatus(infos, c.filter)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestHasDirty_And_DropDirty(t *testing.T) {
	infos := []worktree.Info{
		{Status: worktree.StatusPushed},
		{Status: worktree.StatusDirty},
	}
	if !HasDirty(infos, []int{0, 1}) {
		t.Error("want true")
	}
	if HasDirty(infos, []int{0}) {
		t.Error("want false")
	}
	if got := DropDirty(infos, []int{0, 1}); !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("DropDirty: %v", got)
	}
}

func TestHasPrunable_True(t *testing.T) {
	infos := []worktree.Info{
		{Path: "/a", Status: worktree.StatusPushed},
		{Path: "/b", Status: worktree.StatusPrunable},
	}
	if !HasPrunable(infos, []int{0, 1}) {
		t.Error("HasPrunable should return true when a prunable index is included")
	}
}

func TestHasPrunable_False(t *testing.T) {
	infos := []worktree.Info{
		{Path: "/a", Status: worktree.StatusPushed},
		{Path: "/b", Status: worktree.StatusDirty},
	}
	if HasPrunable(infos, []int{0, 1}) {
		t.Error("HasPrunable should return false without prunable")
	}
}
