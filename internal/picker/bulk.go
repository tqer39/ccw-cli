package picker

import "github.com/tqer39/ccw-cli/internal/worktree"

// SelectByStatus returns indices of infos whose Status is in the filter set.
// If filter is nil or empty, all indices are returned.
func SelectByStatus(infos []worktree.Info, filter map[worktree.Status]bool) []int {
	out := make([]int, 0, len(infos))
	for i, w := range infos {
		if len(filter) == 0 || filter[w.Status] {
			out = append(out, i)
		}
	}
	return out
}

// HasDirty reports whether any of the given indices references a dirty worktree.
func HasDirty(infos []worktree.Info, indices []int) bool {
	for _, i := range indices {
		if infos[i].Status == worktree.StatusDirty {
			return true
		}
	}
	return false
}

// DropDirty returns indices with StatusDirty entries removed.
func DropDirty(infos []worktree.Info, indices []int) []int {
	out := make([]int, 0, len(indices))
	for _, i := range indices {
		if infos[i].Status != worktree.StatusDirty {
			out = append(out, i)
		}
	}
	return out
}
