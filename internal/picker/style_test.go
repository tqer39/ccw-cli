package picker

import (
	"strings"
	"testing"
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
