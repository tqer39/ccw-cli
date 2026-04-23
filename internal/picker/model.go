// Package picker renders the ccw worktree picker (bubbletea) and returns the
// user's decision as a pure (Action, Selection) value for the caller to act on.
package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tqer39/ccw-cli/internal/gh"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

// Action is the user's choice returned by Run.
type Action int

// Action values. ActionCancel is the zero value so an uninitialized result
// means "no action taken".
const (
	ActionCancel Action = iota
	ActionResume
	ActionDelete
	ActionNew
	ActionBulkDelete
)

// String returns the lowercase name of the action.
func (a Action) String() string {
	switch a {
	case ActionResume:
		return "resume"
	case ActionDelete:
		return "delete"
	case ActionNew:
		return "new"
	case ActionBulkDelete:
		return "bulk-delete"
	default:
		return "cancel"
	}
}

// Selection identifies the worktree the user picked.
type Selection struct {
	Path        string
	Branch      string
	Status      worktree.Status
	ForceDelete bool
}

// BulkDeletion は複数 worktree をまとめて消す際の宛先。
type BulkDeletion struct {
	Paths []string
	Force bool
}

// Icon maps a worktree.Status to a one-rune glyph (legacy API, kept for back-
// compat). The picker now renders Badge() instead.
func Icon(s worktree.Status) string {
	switch s {
	case worktree.StatusPushed:
		return "✅"
	case worktree.StatusLocalOnly:
		return "⚠"
	case worktree.StatusDirty:
		return "⛔"
	default:
		return "•"
	}
}

// state is an internal state-machine tag.
type state int

const (
	stateList state = iota
	stateMenu
	stateDeleteConfirm
	stateBulkFilter
	stateBulkConfirm
)

// Model is the bubbletea model for the picker.
type Model struct {
	state         state
	infos         []worktree.Info
	list          list.Model
	selIdx        int
	action        Action
	selection     Selection
	width         int
	height        int
	prs           map[string]gh.PRInfo
	prUnavailable bool
	bulkFilter    map[worktree.Status]bool
	bulkTargets   []int
	bulkForce     bool
}

// listItem is a bubbles/list.Item with a tag that lets us distinguish
// real worktree rows from the synthetic menu rows.
type listItem struct {
	title string
	desc  string
	tag   itemTag
	idx   int
	wt    *worktree.Info
	pr    *gh.PRInfo
}

type itemTag int

const (
	tagWorktree itemTag = iota
	tagNew
	tagQuit
	tagDeleteAll
	tagCleanPushed
	tagCustomSelect
)

// Title / Description / FilterValue implement list.Item. For worktree rows
// the custom delegate ignores these, but they still need to be non-empty to
// satisfy bubbles/list internal bookkeeping.
func (i listItem) Title() string {
	if i.wt != nil {
		return fmt.Sprintf("%s (%s)", i.wt.Branch, i.wt.Status)
	}
	return i.title
}

func (i listItem) Description() string {
	if i.wt != nil {
		return i.wt.Path
	}
	return i.desc
}

func (i listItem) FilterValue() string { return i.Title() }

// New constructs a Model from the worktree list.
func New(infos []worktree.Info) Model {
	items := make([]list.Item, 0, len(infos)+5)
	for i := range infos {
		cp := infos[i]
		items = append(items, listItem{tag: tagWorktree, idx: i, wt: &cp})
	}
	items = append(items,
		listItem{title: "🗑️  [delete all]", desc: "Remove all worktrees (confirm required)", tag: tagDeleteAll},
		listItem{title: "🧹  [clean pushed]", desc: "Remove worktrees that are pushed & clean", tag: tagCleanPushed},
		listItem{title: "☑️  [custom select]", desc: "Pick by status (pushed / local-only / dirty)", tag: tagCustomSelect},
		listItem{title: "➕  [new]", desc: "Start a new worktree", tag: tagNew},
		listItem{title: "🚪  [quit]", desc: "Cancel", tag: tagQuit},
	)
	delegate := rowDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "ccw worktrees"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return Model{state: stateList, infos: infos, list: l}
}

// Action returns the action the user chose (valid after the program exits).
func (m Model) Action() Action { return m.action }

// Selection returns the selected worktree (valid after the program exits).
func (m Model) Selection() Selection { return m.selection }

// Bulk returns the bulk-delete descriptor (valid after ActionBulkDelete).
func (m Model) Bulk() BulkDeletion {
	paths := make([]string, 0, len(m.bulkTargets))
	for _, i := range m.bulkTargets {
		paths = append(paths, m.infos[i].Path)
	}
	return BulkDeletion{Paths: paths, Force: m.bulkForce}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if !gh.Available() {
		return nil
	}
	branches := make([]string, 0, len(m.infos))
	for _, w := range m.infos {
		branches = append(branches, w.Branch)
	}
	return fetchPRsCmd(branches)
}

// prFetchedMsg は gh.PRStatus が成功したときに送られる。
type prFetchedMsg struct{ prs map[string]gh.PRInfo }

// prFetchErrMsg は gh の呼び出しが失敗したときに送られる。
type prFetchErrMsg struct{ err error }

// fetchPRsCmd は branches 分の PR を非同期取得する。
func fetchPRsCmd(branches []string) tea.Cmd {
	return func() tea.Msg {
		m, err := gh.PRStatus(branches)
		if err != nil {
			return prFetchErrMsg{err: err}
		}
		return prFetchedMsg{prs: m}
	}
}

// applyPRsToItems は現在の m.prs / m.prUnavailable を delegate と listItem に反映する。
func (m *Model) applyPRsToItems() {
	items := m.list.Items()
	for i, it := range items {
		li, ok := it.(listItem)
		if !ok || li.tag != tagWorktree {
			continue
		}
		if pr, ok := m.prs[li.wt.Branch]; ok {
			cp := pr
			li.pr = &cp
		} else {
			li.pr = nil
		}
		items[i] = li
	}
	m.list.SetItems(items)
	m.list.SetDelegate(rowDelegate{prUnavailable: m.prUnavailable})
}
