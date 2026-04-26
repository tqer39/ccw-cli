package main

import (
	"os"
	"path/filepath"
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

func TestWithPluginDir_Disabled(t *testing.T) {
	in := []string{"--model", "opus"}
	got := withPluginDir(false, in)
	if len(got) != len(in) {
		t.Fatalf("disabled should not modify passthrough, got %v", got)
	}
	for i := range in {
		if got[i] != in[i] {
			t.Errorf("at %d: got %q, want %q", i, got[i], in[i])
		}
	}
}

func TestWithPluginDir_EnabledMiss(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	in := []string{"--model", "opus"}
	got := withPluginDir(true, in)
	if len(got) != len(in) {
		t.Fatalf("miss should not modify passthrough, got %v", got)
	}
}

func TestWithPluginDir_EnabledHit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest", ".claude-plugin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	in := []string{"--model", "opus"}
	got := withPluginDir(true, in)
	if len(got) != len(in)+2 {
		t.Fatalf("expected 2 extra args, got %v", got)
	}
	if got[0] != "--plugin-dir" {
		t.Errorf("first arg should be --plugin-dir, got %q", got[0])
	}
	wantDir := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest")
	if got[1] != wantDir {
		t.Errorf("plugin-dir path: got %q, want %q", got[1], wantDir)
	}
	if got[2] != "--model" || got[3] != "opus" {
		t.Errorf("passthrough not preserved: %v", got)
	}
}
