package picker

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateList:
		return m.updateList(msg)
	case stateMenu:
		return m.updateMenu(msg)
	case stateDeleteConfirm:
		return m.updateDeleteConfirm(msg)
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.action = ActionCancel
		return m, tea.Quit
	case "enter":
		item, ok := m.list.SelectedItem().(listItem)
		if !ok {
			return m, nil
		}
		switch item.tag {
		case tagNew:
			m.action = ActionNew
			return m, tea.Quit
		case tagQuit:
			m.action = ActionCancel
			return m, tea.Quit
		case tagWorktree:
			m.selIdx = item.idx
			m.state = stateMenu
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// updateMenu is a stub filled in by Task 4.
func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	_ = msg
	return m, nil
}

// updateDeleteConfirm is a stub filled in by Task 5.
func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	_ = msg
	return m, nil
}
