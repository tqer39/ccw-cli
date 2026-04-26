package main

import (
	"strings"
	"testing"
)

func TestMaybePreamble_Disabled(t *testing.T) {
	if got := maybePreamble(false); got != "" {
		t.Errorf("disabled should return empty, got %q", got)
	}
}

func TestMaybePreamble_Enabled(t *testing.T) {
	got := maybePreamble(true)
	if got == "" {
		t.Fatal("enabled should return non-empty preamble")
	}
	if !strings.Contains(got, "superpowers:brainstorming") {
		t.Errorf("preamble missing brainstorming step: %q", got)
	}
}
