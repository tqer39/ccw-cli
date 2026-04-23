package superpowers

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectInstalled_NotInstalled(t *testing.T) {
	home := t.TempDir()
	ok, err := DetectInstalled(home)
	if err != nil {
		t.Fatalf("DetectInstalled: %v", err)
	}
	if ok {
		t.Error("DetectInstalled empty home = true, want false")
	}
}

func TestDetectInstalled_PresentAfterMkdir(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, ".claude", "plugins", "cache", "some-plugin", "superpowers")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	ok, err := DetectInstalled(home)
	if err != nil {
		t.Fatalf("DetectInstalled: %v", err)
	}
	if !ok {
		t.Error("DetectInstalled after mkdir = false, want true")
	}
}

func TestDetectInstalled_IgnoresFile(t *testing.T) {
	home := t.TempDir()
	base := filepath.Join(home, ".claude", "plugins", "cache", "p")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "superpowers"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := DetectInstalled(home)
	if err != nil {
		t.Fatalf("DetectInstalled: %v", err)
	}
	if ok {
		t.Error("DetectInstalled matched file (want directory only)")
	}
}

func TestEnsureInstalled_Present(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "plugins", "cache", "any", "superpowers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := EnsureInstalled(strings.NewReader(""), &out, home, true); err != nil {
		t.Errorf("EnsureInstalled present = %v, want nil", err)
	}
	if out.Len() != 0 {
		t.Errorf("no output expected when present, got %q", out.String())
	}
}

func TestEnsureInstalled_MissingNonInteractive(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	err := EnsureInstalled(strings.NewReader(""), &out, home, false)
	if err == nil {
		t.Fatal("EnsureInstalled missing non-interactive: want error")
	}
}

func TestEnsureInstalled_UserCancels(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	err := EnsureInstalled(strings.NewReader("n\n"), &out, home, true)
	if err == nil {
		t.Fatal("EnsureInstalled user-cancel: want error")
	}
}

func TestEnsureInstalled_InstallSucceeds(t *testing.T) {
	home := t.TempDir()
	var called bool
	orig := installRunner
	installRunner = func() error { called = true; return nil }
	t.Cleanup(func() { installRunner = orig })

	var out bytes.Buffer
	if err := EnsureInstalled(strings.NewReader("y\n"), &out, home, true); err != nil {
		t.Fatalf("EnsureInstalled success: %v", err)
	}
	if !called {
		t.Error("installRunner not called")
	}
}

func TestEnsureInstalled_InstallFails(t *testing.T) {
	home := t.TempDir()
	orig := installRunner
	installRunner = func() error { return errFake }
	t.Cleanup(func() { installRunner = orig })

	var out bytes.Buffer
	err := EnsureInstalled(strings.NewReader("y\n"), &out, home, true)
	if err == nil {
		t.Fatal("EnsureInstalled install-fail: want error")
	}
}

var errFake = fakeErr("boom")

type fakeErr string

func (f fakeErr) Error() string { return string(f) }
