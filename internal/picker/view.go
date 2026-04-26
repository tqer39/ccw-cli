package picker

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tqer39/ccw-cli/internal/i18n"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// View implements tea.Model.
func (m Model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	return v
}

func (m Model) viewContent() string {
	switch m.state {
	case stateList:
		base := m.list.View()
		footer := ""
		switch {
		case !m.ghAvailable:
			footer = i18n.T(i18n.KeyPickerFooterInstallGh)
		case m.tip != "":
			footer = i18n.T(i18n.KeyPickerFooterTip, m.tip)
		}
		if footer == "" {
			return base
		}
		return base + "\n\n" + footer
	case stateMenu:
		return m.menuView()
	case stateDeleteConfirm:
		return m.deleteConfirmView()
	case stateBulkFilter:
		return m.bulkFilterView()
	case stateBulkConfirm:
		return m.bulkConfirmView()
	}
	return ""
}

func (m Model) menuView() string {
	w := m.infos[m.selIdx]
	return i18n.T(i18n.KeyPickerActionMenu, w.Branch, w.Status, w.Path)
}

func (m Model) deleteConfirmView() string {
	w := m.infos[m.selIdx]
	if w.Status == worktree.StatusPrunable {
		return m.prunableConfirmView()
	}
	cmd := fmt.Sprintf("git worktree remove %q", w.Path)
	if w.Status == worktree.StatusDirty {
		cmd = fmt.Sprintf("git worktree remove --force %q", w.Path)
	}
	return i18n.T(i18n.KeyPickerDeleteConfirm, w.Branch, w.Path, w.Status, cmd)
}

func (m Model) prunableConfirmView() string {
	var prunables []worktree.Info
	for _, in := range m.infos {
		if in.Status == worktree.StatusPrunable {
			prunables = append(prunables, in)
		}
	}
	if len(prunables) <= 1 {
		w := m.infos[m.selIdx]
		return i18n.T(i18n.KeyPickerPruneSingle, w.Branch, w.Path)
	}
	var b strings.Builder
	b.WriteString(i18n.T(i18n.KeyPickerPruneBulkHead, len(prunables)))
	for _, p := range prunables {
		b.WriteString(i18n.T(i18n.KeyPickerPruneBulkLine, p.Branch, p.Path))
	}
	b.WriteString(i18n.T(i18n.KeyPickerPruneBulkFoot))
	return b.String()
}

func (m Model) bulkFilterView() string {
	var b strings.Builder
	b.WriteString(i18n.T(i18n.KeyPickerBulkFilterHead))
	for _, s := range []worktree.Status{
		worktree.StatusPushed, worktree.StatusLocalOnly, worktree.StatusDirty,
	} {
		mark := "[ ]"
		if m.bulkFilter[s] {
			mark = "[x]"
		}
		b.WriteString(i18n.T(i18n.KeyPickerBulkFilterLine, mark, s))
	}
	b.WriteString(i18n.T(i18n.KeyPickerBulkFilterKeys))
	b.WriteString(i18n.T(i18n.KeyPickerBulkFilterFoot))
	return b.String()
}

func (m Model) bulkConfirmView() string {
	var b strings.Builder
	b.WriteString(i18n.T(i18n.KeyPickerBulkConfirmHead, len(m.bulkTargets)))
	hasDirty := HasDirty(m.infos, m.bulkTargets)
	hasPrunable := HasPrunable(m.infos, m.bulkTargets)
	dirtyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	for _, i := range m.bulkTargets {
		w := m.infos[i]
		line := fmt.Sprintf("  %s %s  %s\n", Badge(w.Status), w.Branch, w.Path)
		if w.Status == worktree.StatusDirty && !noColor() {
			line = dirtyStyle.Render(line)
		}
		b.WriteString(line)
	}
	if hasPrunable {
		b.WriteString(i18n.T(i18n.KeyPickerBulkPruneNote))
	}
	if hasDirty {
		b.WriteString(i18n.T(i18n.KeyPickerBulkDirtyWarn))
	} else {
		b.WriteString(i18n.T(i18n.KeyPickerBulkConfirmYN))
	}
	return b.String()
}
