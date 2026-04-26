package picker

import (
	"fmt"
	"io"
	"path/filepath"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// rowDelegate renders worktree items as three lines: header (resume + tree
// icon + worktree name + status badge + indicators), branch, pr.
type rowDelegate struct {
	prUnavailable bool
}

func (d rowDelegate) Height() int                             { return 3 }
func (d rowDelegate) Spacing() int                            { return 1 }
func (d rowDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d rowDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	li, ok := item.(listItem)
	if !ok {
		return
	}
	selected := index == m.Index()
	_, _ = fmt.Fprint(w, renderRow(li, m.Width(), d.prUnavailable, selected))
}

func renderRow(li listItem, width int, prUnavailable bool, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	switch li.tag {
	case tagNew, tagQuit, tagDeleteAll, tagCleanPushed, tagCustomSelect:
		return prefix + li.title + "\n  " + li.desc
	}
	wt := li.wt
	name := filepath.Base(wt.Path)
	resume := ResumeBadge(wt.HasSession)
	sep := Separator()

	// Reserve a 4-cell right-edge margin so IDE-embedded terminals (Cursor,
	// cmux, ...) that report Width slightly larger than the visible area
	// don't clip the right edge. Falls back to the raw width when it is too
	// small to shrink meaningfully.
	effectiveWidth := width
	if width > 4 {
		effectiveWidth = width - 4
	}

	var header string
	if wt.Status == worktree.StatusPrunable {
		header = prefix + resume + sep + "🌲 " + name + sep + MissingOnDisk()
	} else {
		status := Badge(wt.Status)
		indicators := wt.Indicators()
		header = prefix + resume + sep + "🌲 " + name + sep + status + sep + indicators
	}

	branchLine := fmt.Sprintf("    branch:  %s", wt.Branch)
	prCell := ""
	if !prUnavailable {
		prCell = renderPRCell(li.pr)
	}
	prLine := "    pr:      " + prCell

	if width > 0 {
		header = truncateToWidth(header, effectiveWidth)
		branchLine = truncateToWidth(branchLine, effectiveWidth)
		prLine = truncateToWidth(prLine, effectiveWidth)
	}

	return header + "\n" + branchLine + "\n" + prLine
}

// renderPRCell builds the PR portion of the row: either a state-tinted
// `[STATE] #N "title"` cell or a dim `(no PR)` placeholder.
func renderPRCell(pr *gh.PRInfo) string {
	if pr == nil {
		if noColor() {
			return "(no PR)"
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(no PR)")
	}
	title := pr.Title
	if len(title) > 30 {
		title = title[:29] + "…"
	}
	inner := fmt.Sprintf("%s #%d %q", PRBadge(pr.State), pr.Number, title)
	return PRCellStyle(pr.State).Render(inner)
}

// truncateToWidth trims the visible width of s to n cells.
func truncateToWidth(s string, n int) string {
	if lipgloss.Width(s) <= n {
		return s
	}
	for len(s) > 0 && lipgloss.Width(s) > n {
		s = s[:len(s)-1]
	}
	return s
}
