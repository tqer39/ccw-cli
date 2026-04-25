package picker

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
			footer = "💡 Install gh to see PR titles here"
		case m.tip != "":
			footer = "💡 Tip: " + m.tip
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
	return fmt.Sprintf(
		"Selected: %s (%s)\nPath:     %s\n\nWhat to do?\n  [r] run    — start claude in this worktree\n  [d] delete — remove the worktree\n  [b] back   — return to list\n  [q] quit   — cancel\n",
		w.Branch, w.Status, w.Path,
	)
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
	return fmt.Sprintf(
		"Delete worktree %s?\n  path:   %s\n  status: %s\n\nThis will run: %s\n\nConfirm? [y/N]\n",
		w.Branch, w.Path, w.Status, cmd,
	)
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
		return fmt.Sprintf(
			"Prune worktree %s?\n  path:   %s\n\nThis will run: git worktree prune\n\nConfirm? [y/N]\n",
			w.Branch, w.Path,
		)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Prune %d prunable worktrees? (git worktree prune removes all of them at once)\n\n", len(prunables))
	for _, p := range prunables {
		fmt.Fprintf(&b, "  %s %s\n", p.Branch, p.Path)
	}
	b.WriteString("\nThis will run: git worktree prune\n\nConfirm? [y/N]\n")
	return b.String()
}

func (m Model) bulkFilterView() string {
	var b strings.Builder
	b.WriteString("Select statuses to delete (toggle):\n\n")
	for _, s := range []worktree.Status{
		worktree.StatusPushed, worktree.StatusLocalOnly, worktree.StatusDirty,
	} {
		mark := "[ ]"
		if m.bulkFilter[s] {
			mark = "[x]"
		}
		fmt.Fprintf(&b, "  %s %s\n", mark, s)
	}
	b.WriteString("\n  [p] pushed  [l] local-only  [d] dirty  [a] clear\n")
	b.WriteString("  [enter] confirm  [q/esc] back\n")
	return b.String()
}

func (m Model) bulkConfirmView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Delete %d worktrees?\n\n", len(m.bulkTargets))
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
		b.WriteString("\nℹ Prunable entries will be cleaned up via `git worktree prune` after the removals.\n")
	}
	if hasDirty {
		b.WriteString("\n⚠ Dirty worktrees are included. `git worktree remove --force` is required.\n")
		b.WriteString("  [y] yes (include dirty, use --force)\n")
		b.WriteString("  [s] skip dirty (remove clean only)\n")
		b.WriteString("  [N] cancel\n")
	} else {
		b.WriteString("\nConfirm? [y/N]\n")
	}
	return b.String()
}
