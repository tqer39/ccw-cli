package superpowers

import (
	"strings"
	"testing"

	"github.com/tqer39/ccw-cli/internal/i18n"
)

func TestPreamble_JA(t *testing.T) {
	got := Preamble(i18n.LangJA)
	if !strings.Contains(got, "sandbox") {
		t.Errorf("ja preamble missing 'sandbox': %q", got)
	}
	if !strings.Contains(got, "順で進めて") {
		t.Errorf("ja preamble not in Japanese: %q", got)
	}
	if !strings.Contains(got, "superpowers:brainstorming") {
		t.Errorf("ja preamble missing brainstorming step: %q", got)
	}
}

func TestPreamble_EN(t *testing.T) {
	got := Preamble(i18n.LangEN)
	if !strings.Contains(got, "sandbox") {
		t.Errorf("en preamble missing 'sandbox': %q", got)
	}
	if !strings.Contains(got, "Proceed in this order") {
		t.Errorf("en preamble not in English: %q", got)
	}
	if !strings.Contains(got, "superpowers:brainstorming") {
		t.Errorf("en preamble missing brainstorming step: %q", got)
	}
}

func TestPreamble_UnknownFallsBackToEN(t *testing.T) {
	if Preamble("xx") != Preamble(i18n.LangEN) {
		t.Error("unknown lang should fall back to EN")
	}
}
