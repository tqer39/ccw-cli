package listmode

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderText_Header(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderText(&Output{Version: 1, Worktrees: []WorktreeEntry{}}, &buf); err != nil {
		t.Fatal(err)
	}
	first := strings.SplitN(buf.String(), "\n", 2)[0]
	for _, col := range []string{"NAME", "STATUS", "AHEAD/BEHIND", "PR", "SESSION", "BRANCH"} {
		if !strings.Contains(first, col) {
			t.Errorf("header missing %q\nheader: %s", col, first)
		}
	}
}

func TestRenderText_PushedRow(t *testing.T) {
	out := &Output{
		Version: 1,
		Worktrees: []WorktreeEntry{{
			Name:    "ccw-x",
			Branch:  "worktree-ccw-x",
			Status:  "pushed",
			Ahead:   0,
			Behind:  0,
			PR:      &PRInfo{Number: 42, State: "OPEN"},
			Session: SessionInfo{Exists: true},
		}},
	}
	var buf bytes.Buffer
	_ = RenderText(out, &buf)
	got := buf.String()
	for _, want := range []string{"ccw-x", "pushed", "0/0", "#42 OPEN", "RESUME", "worktree-ccw-x"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\nfull:\n%s", want, got)
		}
	}
}

func TestRenderText_NoSessionRendersNEW(t *testing.T) {
	out := &Output{Worktrees: []WorktreeEntry{{Name: "x", Status: "pushed", Session: SessionInfo{Exists: false}}}}
	var buf bytes.Buffer
	_ = RenderText(out, &buf)
	if !strings.Contains(buf.String(), "NEW") {
		t.Errorf("expected NEW, got: %s", buf.String())
	}
}

func TestRenderText_NoPRRendersDash(t *testing.T) {
	out := &Output{Worktrees: []WorktreeEntry{{Name: "x", Status: "pushed", PR: nil}}}
	var buf bytes.Buffer
	_ = RenderText(out, &buf)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("not enough lines: %s", buf.String())
	}
	if !strings.Contains(lines[1], " - ") {
		t.Errorf("expected '-' in row, got: %s", lines[1])
	}
}

func TestRenderText_PrunableShowsDashAhead(t *testing.T) {
	out := &Output{Worktrees: []WorktreeEntry{{Name: "p", Status: "prunable"}}}
	var buf bytes.Buffer
	_ = RenderText(out, &buf)
	if !strings.Contains(buf.String(), "prunable") {
		t.Errorf("missing prunable: %s", buf.String())
	}
	if strings.Contains(buf.String(), "0/0") {
		t.Errorf("prunable row should not show 0/0: %s", buf.String())
	}
}

func TestRenderText_NoANSIEscapes(t *testing.T) {
	out := &Output{Worktrees: []WorktreeEntry{{Name: "x", Status: "dirty"}}}
	var buf bytes.Buffer
	_ = RenderText(out, &buf)
	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("ANSI escape leaked: %q", buf.String())
	}
}
