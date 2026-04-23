package picker

import (
	"fmt"

	"github.com/tqer39/ccw-cli/internal/worktree"
)

// View implements tea.Model.
func (m Model) View() string {
	switch m.state {
	case stateList:
		return m.list.View()
	case stateMenu:
		return m.menuView()
	case stateDeleteConfirm:
		return m.deleteConfirmView()
	}
	return ""
}

func (m Model) menuView() string {
	w := m.infos[m.selIdx]
	return fmt.Sprintf(
		"Selected: %s (%s)\nPath:     %s\n\nWhat to do?\n  [r] resume — open claude in this worktree\n  [d] delete — remove the worktree\n  [b] back   — return to list\n  [q] quit   — cancel\n",
		w.Branch, w.Status, w.Path,
	)
}

func (m Model) deleteConfirmView() string {
	w := m.infos[m.selIdx]
	cmd := fmt.Sprintf("git worktree remove %q", w.Path)
	if w.Status == worktree.StatusDirty {
		cmd = fmt.Sprintf("git worktree remove --force %q", w.Path)
	}
	return fmt.Sprintf(
		"Delete worktree %s?\n  path:   %s\n  status: %s\n\nThis will run: %s\n\nConfirm? [y/N]\n",
		w.Branch, w.Path, w.Status, cmd,
	)
}
