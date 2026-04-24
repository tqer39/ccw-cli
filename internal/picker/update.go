package picker

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case prFetchedMsg:
		m.prs = msg.prs
		m.applyPRsToItems()
		return m, nil
	case prFetchErrMsg:
		m.prUnavailable = true
		m.applyPRsToItems()
		if os.Getenv("CCW_DEBUG") != "" {
			_, _ = fmt.Fprintln(os.Stderr, "ccw: PR fetch failed:", msg.err)
		}
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
	case stateBulkFilter:
		return m.updateBulkFilter(msg)
	case stateBulkConfirm:
		return m.updateBulkConfirm(msg)
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
		case tagDeleteAll:
			m.bulkTargets = SelectByStatus(m.infos, nil)
			if len(m.bulkTargets) == 0 {
				return m, nil
			}
			m.state = stateBulkConfirm
			return m, nil
		case tagCleanPushed:
			filter := map[worktree.Status]bool{worktree.StatusPushed: true}
			m.bulkTargets = SelectByStatus(m.infos, filter)
			if len(m.bulkTargets) == 0 {
				return m, nil
			}
			m.state = stateBulkConfirm
			return m, nil
		case tagCustomSelect:
			m.bulkFilter = map[worktree.Status]bool{}
			m.state = stateBulkFilter
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "R":
		m.action = ActionResume
		m.selection = m.currentSelection()
		return m, tea.Quit
	case "d", "D":
		m.state = stateDeleteConfirm
		return m, nil
	case "b", "B", "esc":
		m.state = stateList
		return m, nil
	case "q", "Q", "ctrl+c":
		m.action = ActionCancel
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) currentSelection() Selection {
	w := m.infos[m.selIdx]
	return Selection{Path: w.Path, Branch: w.Branch, Status: w.Status}
}

func (m Model) updateBulkFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "p", "P":
		m.bulkFilter[worktree.StatusPushed] = !m.bulkFilter[worktree.StatusPushed]
		return m, nil
	case "l", "L":
		m.bulkFilter[worktree.StatusLocalOnly] = !m.bulkFilter[worktree.StatusLocalOnly]
		return m, nil
	case "d", "D":
		m.bulkFilter[worktree.StatusDirty] = !m.bulkFilter[worktree.StatusDirty]
		return m, nil
	case "a", "A":
		m.bulkFilter = map[worktree.Status]bool{}
		return m, nil
	case "enter":
		m.bulkTargets = SelectByStatus(m.infos, m.bulkFilter)
		if len(m.bulkTargets) == 0 {
			m.state = stateList
			return m, nil
		}
		m.state = stateBulkConfirm
		return m, nil
	case "q", "esc", "ctrl+c":
		m.state = stateList
		return m, nil
	}
	return m, nil
}

func (m Model) updateBulkConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	hasDirty := HasDirty(m.infos, m.bulkTargets)
	switch msg.String() {
	case "y", "Y":
		m.bulkForce = hasDirty
		m.action = ActionBulkDelete
		return m, tea.Quit
	case "s", "S":
		if !hasDirty {
			return m, nil
		}
		m.bulkTargets = DropDirty(m.infos, m.bulkTargets)
		m.bulkForce = false
		m.action = ActionBulkDelete
		return m, tea.Quit
	case "n", "N", "esc":
		m.state = stateList
		return m, nil
	case "q", "Q", "ctrl+c":
		m.action = ActionCancel
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		sel := m.currentSelection()
		sel.ForceDelete = sel.Status == worktree.StatusDirty
		m.selection = sel
		m.action = ActionDelete
		return m, tea.Quit
	case "n", "N", "b", "B", "esc":
		m.state = stateList
		return m, nil
	case "q", "Q", "ctrl+c":
		m.action = ActionCancel
		return m, tea.Quit
	}
	return m, nil
}
