package picker

import "fmt"

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

// deleteConfirmView is a stub filled in by Task 5.
func (m Model) deleteConfirmView() string { return "" }
