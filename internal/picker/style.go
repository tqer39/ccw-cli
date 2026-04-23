package picker

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// Badge renders a fixed-width status badge for a Status.
// Respects NO_COLOR=1 by returning a plain-text bracketed label.
func Badge(s worktree.Status) string {
	label, plain := badgeLabel(s)
	if noColor() {
		return plain
	}
	style := lipgloss.NewStyle().Padding(0, 1).Bold(true)
	switch s {
	case worktree.StatusPushed:
		style = style.Background(lipgloss.Color("10")).Foreground(lipgloss.Color("0"))
	case worktree.StatusLocalOnly:
		style = style.Background(lipgloss.Color("11")).Foreground(lipgloss.Color("0"))
	case worktree.StatusDirty:
		style = style.Background(lipgloss.Color("9")).Foreground(lipgloss.Color("15"))
	}
	return style.Render(label)
}

func badgeLabel(s worktree.Status) (colored, plain string) {
	switch s {
	case worktree.StatusPushed:
		return "PUSHED", "[pushed]"
	case worktree.StatusLocalOnly:
		return "LOCAL ", "[local] "
	case worktree.StatusDirty:
		return "DIRTY ", "[dirty] "
	default:
		return "??????", "[?]     "
	}
}

func noColor() bool { return os.Getenv("NO_COLOR") != "" }
