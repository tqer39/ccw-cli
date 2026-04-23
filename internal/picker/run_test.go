package picker

import (
	"bytes"
	"strings"
	"testing"

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
