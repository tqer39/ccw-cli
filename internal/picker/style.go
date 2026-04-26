package picker

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
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
	case worktree.StatusPrunable:
		style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
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
	case worktree.StatusPrunable:
		return "PRUNE ", "[prune] "
	default:
		return "??????", "[?]     "
	}
}

func noColor() bool { return os.Getenv("NO_COLOR") != "" }

// Separator returns the dim-grey vertical bar used to join inline header
// segments (resume, name, status, indicators). Falls back to plain " | "
// when noColor() is true.
func Separator() string {
	if noColor() {
		return " | "
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" | ")
}

// MissingOnDisk returns the dim-grey "(missing on disk)" label used in place
// of status / indicators for prunable worktrees whose physical directory is
// gone. Falls back to plain text when noColor() is true.
func MissingOnDisk() string {
	const label = "(missing on disk)"
	if noColor() {
		return label
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(label)
}

// PRBadge renders a PR state badge. Upstream states from `gh pr list` are
// OPEN / DRAFT / MERGED / CLOSED; unknown values fall back to a lowercased
// bracketed label with no color.
func PRBadge(state string) string {
	if noColor() {
		return "[" + strings.ToLower(state) + "]"
	}
	fg, bg, ok := prBadgeColor(state)
	if !ok {
		return "[" + strings.ToLower(state) + "]"
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Render("[" + strings.ToUpper(state) + "]")
}

// PRCellStyle returns the style that wraps the full PR cell
// (badge + `#N "title"`) with a dim state-tinted background.
// Returns an empty style when NO_COLOR is set or state is unknown.
func PRCellStyle(state string) lipgloss.Style {
	if noColor() {
		return lipgloss.NewStyle()
	}
	bg, ok := prCellBackground(state)
	if !ok {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(bg))
}

func prBadgeColor(state string) (fg, bg string, ok bool) {
	switch state {
	case "OPEN":
		return "0", "2", true
	case "DRAFT":
		return "15", "8", true
	case "MERGED":
		return "15", "5", true
	case "CLOSED":
		return "15", "1", true
	}
	return "", "", false
}

func prCellBackground(state string) (string, bool) {
	switch state {
	case "OPEN":
		return "22", true
	case "DRAFT":
		return "237", true
	case "MERGED":
		return "53", true
	case "CLOSED":
		return "52", true
	}
	return "", false
}

// ResumeBadge renders a RESUME / NEW badge. Under NO_COLOR returns
// "[RESUME]" / "[NEW]   " (padded to equal width).
func ResumeBadge(hasSession bool) string {
	if noColor() {
		if hasSession {
			return "[RESUME]"
		}
		return "[NEW]   "
	}
	if hasSession {
		return lipgloss.NewStyle().
			Padding(0, 1).Bold(true).
			Background(lipgloss.Color("14")).
			Foreground(lipgloss.Color("0")).
			Render("💬 RESUME")
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15")).
		Render("⚡ NEW   ")
}
