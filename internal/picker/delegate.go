package picker

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// rowDelegate renders items as two lines: meta (badge/branch/indicators/→/PR)
// on top, path below.
type rowDelegate struct {
	prUnavailable bool
}

func (d rowDelegate) Height() int                             { return 2 }
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

// arrowGlyph returns the worktree→PR separator. ASCII `->` under NO_COLOR
// keeps width predictable across terminals without Unicode.
func arrowGlyph() string {
	if noColor() {
		return "->"
	}
	return "→"
}

// renderRow is a pure function used by the delegate and tests.
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
	badge := Badge(wt.Status)
	indicators := fmt.Sprintf("↑%d ↓%d", wt.AheadCount, wt.BehindCount)
	if wt.Status == worktree.StatusDirty {
		indicators += fmt.Sprintf(" ✎%d", wt.DirtyCount)
	}

	meta := strings.TrimRight(fmt.Sprintf("%s%s  %-24s %s", prefix, badge, wt.Branch, indicators), " ")
	top := meta
	if !prUnavailable {
		top = meta + "  " + arrowGlyph() + "  " + renderPRCell(li.pr)
	}
	if width > 0 && lipgloss.Width(top) > width {
		top = truncateToWidth(top, width)
	}
	return top + "\n  " + wt.Path
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
	// Naive byte-trim fallback: good enough for ASCII-heavy rows.
	for len(s) > 0 && lipgloss.Width(s) > n {
		s = s[:len(s)-1]
	}
	return s
}
