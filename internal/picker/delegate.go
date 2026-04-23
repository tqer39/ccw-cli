package picker

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// rowDelegate renders items as two lines: badge/branch/indicators/PR on top, path below.
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
	var pr string
	if !prUnavailable {
		if li.pr != nil {
			title := li.pr.Title
			if len(title) > 30 {
				title = title[:29] + "…"
			}
			pr = fmt.Sprintf("#%d %s \"%s\"", li.pr.Number, strings.ToLower(li.pr.State), title)
		} else {
			pr = "(no PR)"
		}
	}
	top := strings.TrimRight(fmt.Sprintf("%s%s  %-24s %s  %s", prefix, badge, wt.Branch, indicators, pr), " ")
	if width > 0 && lipgloss.Width(top) > width {
		top = truncateToWidth(top, width)
	}
	return top + "\n  " + wt.Path
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
