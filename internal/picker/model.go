// Package picker renders the ccw worktree picker (bubbletea) and returns the
// user's decision as a pure (Action, Selection) value for the caller to act on.
package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

// Icon maps a worktree.Status to a one-rune glyph displayed in the list.
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
)

// Model is the bubbletea model for the picker.
type Model struct {
	state     state
	infos     []worktree.Info
	list      list.Model
	selIdx    int
	action    Action
	selection Selection
	width     int
	height    int
}

// listItem is a bubbles/list.Item with a tag that lets us distinguish
// real worktree rows from the synthetic [new] / [quit] rows.
type listItem struct {
	title string
	desc  string
	tag   itemTag
	idx   int
}

type itemTag int

const (
	tagWorktree itemTag = iota
	tagNew
	tagQuit
)

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }

// New constructs a Model from the worktree list.
func New(infos []worktree.Info) Model {
	items := make([]list.Item, 0, len(infos)+2)
	for i, w := range infos {
		items = append(items, listItem{
			title: fmt.Sprintf("%s  %s  (%s)", Icon(w.Status), w.Branch, w.Status),
			desc:  w.Path,
			tag:   tagWorktree,
			idx:   i,
		})
	}
	items = append(items,
		listItem{title: "➕  [new]", desc: "Start a new worktree", tag: tagNew},
		listItem{title: "🚪  [quit]", desc: "Cancel", tag: tagQuit},
	)
	delegate := list.NewDefaultDelegate()
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

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }
