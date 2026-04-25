package picker

import (
	"strings"
	"testing"

	"github.com/tqer39/ccw-cli/internal/worktree"
)

func TestPRBadge_NoColorLowercase(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cases := map[string]string{
		"OPEN":   "[open]",
		"DRAFT":  "[draft]",
		"MERGED": "[merged]",
		"CLOSED": "[closed]",
	}
	for in, want := range cases {
		got := PRBadge(in)
		if got != want {
			t.Errorf("PRBadge(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPRBadge_ColoredContainsLabel(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// lipgloss v2 の Style.Render は検出したプロファイルに依存せず
	// 常に ANSI エスケープを返す（プロファイル側のフィルタは Writer 層で行われる）。
	// そのためここで profile を強制設定する必要はない。

	for _, state := range []string{"OPEN", "DRAFT", "MERGED", "CLOSED"} {
		got := PRBadge(state)
		if !strings.Contains(got, "["+state+"]") {
			t.Errorf("PRBadge(%q) should contain [%s], got %q", state, state, got)
		}
		if !strings.Contains(got, "\x1b[") {
			t.Errorf("PRBadge(%q) expected ANSI escape when NO_COLOR unset, got %q", state, got)
		}
	}
}

func TestPRBadge_UnknownState(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := PRBadge("WEIRD")
	if got != "[weird]" {
		t.Errorf("PRBadge(WEIRD) = %q, want [weird]", got)
	}
}

func TestResumeBadge_HasSession(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if got := ResumeBadge(true); got != "[RESUME]" {
		t.Errorf("ResumeBadge(true) NO_COLOR = %q, want [RESUME]", got)
	}
	if got := ResumeBadge(false); got != "[NEW]   " {
		t.Errorf("ResumeBadge(false) NO_COLOR = %q, want [NEW]   ", got)
	}
}

func TestResumeBadge_Colored(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// lipgloss v2 はプロファイルを Writer 層で管理するため ColorProfile/SetColorProfile は不要。
	got := ResumeBadge(true)
	if !strings.Contains(got, "RESUME") {
		t.Errorf("ResumeBadge(true) = %q, want substring RESUME", got)
	}
	got = ResumeBadge(false)
	if !strings.Contains(got, "NEW") {
		t.Errorf("ResumeBadge(false) = %q, want substring NEW", got)
	}
}

func TestBadge_PrunableNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := Badge(worktree.StatusPrunable)
	if got != "[prune] " {
		t.Errorf("Badge(prunable) NO_COLOR = %q, want %q", got, "[prune] ")
	}
}
