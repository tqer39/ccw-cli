package picker

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestRunFallback_Resume(t *testing.T) {
	infos := []worktree.Info{
		{Path: "/a/.claude/worktrees/x", Branch: "x", Status: worktree.StatusPushed},
		{Path: "/a/.claude/worktrees/y", Branch: "y", Status: worktree.StatusLocalOnly},
	}
	var out bytes.Buffer
	a, s, err := runFallback(infos, strings.NewReader("2\n"), &out)
	if err != nil {
		t.Fatalf("runFallback: %v", err)
	}
	if a != ActionResume || s.Branch != "y" {
		t.Errorf("got (%s, %+v), want (resume, y)", a, s)
	}
}

func TestRunFallback_New(t *testing.T) {
	infos := []worktree.Info{{Path: "/a", Branch: "x", Status: worktree.StatusPushed}}
	var out bytes.Buffer
	a, _, err := runFallback(infos, strings.NewReader("n\n"), &out)
	if err != nil {
		t.Fatalf("runFallback: %v", err)
	}
	if a != ActionNew {
		t.Errorf("got %s, want new", a)
	}
}

func TestRunFallback_QuitByEOF(t *testing.T) {
	infos := []worktree.Info{{Path: "/a", Branch: "x", Status: worktree.StatusPushed}}
	var out bytes.Buffer
	a, _, err := runFallback(infos, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("runFallback EOF: %v", err)
	}
	if a != ActionCancel {
		t.Errorf("got %s, want cancel", a)
	}
}

func TestRunFallback_InvalidNumber(t *testing.T) {
	infos := []worktree.Info{{Path: "/a", Branch: "x", Status: worktree.StatusPushed}}
	var out bytes.Buffer
	_, _, err := runFallback(infos, strings.NewReader("99\n"), &out)
	if err == nil {
		t.Fatal("want error for invalid number")
	}
}

func teatestModel(t *testing.T, infos []worktree.Info) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(
		t, New(infos),
		teatest.WithInitialTermSize(80, 24),
	)
}

func TestTUI_ResumeFirstWorktree(t *testing.T) {
	infos := []worktree.Info{
		{Path: "/a/.claude/worktrees/feat", Branch: "feat", Status: worktree.StatusPushed},
	}
	tm := teatestModel(t, infos)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	final, ok := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if !ok {
		t.Fatal("FinalModel is not Model")
	}
	if final.Action() != ActionResume {
		t.Errorf("Action = %s, want resume", final.Action())
	}
	if final.Selection().Branch != "feat" {
		t.Errorf("Selection.Branch = %q, want feat", final.Selection().Branch)
	}
}

func TestTUI_DeleteDirtyForces(t *testing.T) {
	infos := []worktree.Info{
		{Path: "/a/.claude/worktrees/feat", Branch: "feat", Status: worktree.StatusDirty},
	}
	tm := teatestModel(t, infos)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	final, ok := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if !ok {
		t.Fatal("FinalModel is not Model")
	}
	if final.Action() != ActionDelete {
		t.Errorf("Action = %s, want delete", final.Action())
	}
	if !final.Selection().ForceDelete {
		t.Error("ForceDelete = false, want true for dirty")
	}
}

func TestTUI_QCancels(t *testing.T) {
	infos := []worktree.Info{
		{Path: "/a/.claude/worktrees/feat", Branch: "feat", Status: worktree.StatusPushed},
	}
	tm := teatestModel(t, infos)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	final, ok := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if !ok {
		t.Fatal("FinalModel is not Model")
	}
	if final.Action() != ActionCancel {
		t.Errorf("Action = %s, want cancel", final.Action())
	}
}
