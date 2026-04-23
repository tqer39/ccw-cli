package picker

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

// menuView is a stub filled in by Task 4.
func (m Model) menuView() string { return "" }

// deleteConfirmView is a stub filled in by Task 5.
func (m Model) deleteConfirmView() string { return "" }
