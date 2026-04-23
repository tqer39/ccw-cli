package claude

import (
	"bytes"
	"strings"
	"testing"
)

func TestEnsureInstalled_ChoiceNpmCallsInstaller(t *testing.T) {
	t.Setenv("PATH", "")
	origNpm := npmInstaller
	origBrew := brewInstaller
	origLookPath := lookPath
	t.Cleanup(func() {
		npmInstaller = origNpm
		brewInstaller = origBrew
		lookPath = origLookPath
	})

	lookPath = func(name string) (string, error) {
		if name == "claude" {
			return "", errFakeMissing
		}
		if name == "npm" {
			return "/fake/npm", nil
		}
		return "", errFakeMissing
	}
	var called bool
	npmInstaller = func() error { called = true; return nil }
	brewInstaller = func() error { t.Fatal("brew should not be called"); return nil }

	var out bytes.Buffer
	if err := EnsureInstalled(strings.NewReader("1\n"), &out, true); err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	if !called {
		t.Error("npmInstaller not called")
	}
}

func TestEnsureInstalled_ChoiceBrewCallsInstaller(t *testing.T) {
	origNpm := npmInstaller
	origBrew := brewInstaller
	origLookPath := lookPath
	t.Cleanup(func() {
		npmInstaller = origNpm
		brewInstaller = origBrew
		lookPath = origLookPath
	})

	lookPath = func(name string) (string, error) {
		if name == "claude" {
			return "", errFakeMissing
		}
		if name == "brew" {
			return "/fake/brew", nil
		}
		return "", errFakeMissing
	}
	var called bool
	brewInstaller = func() error { called = true; return nil }
	npmInstaller = func() error { t.Fatal("npm should not be called"); return nil }

	var out bytes.Buffer
	if err := EnsureInstalled(strings.NewReader("2\n"), &out, true); err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	if !called {
		t.Error("brewInstaller not called")
	}
}

func TestEnsureInstalled_ChoiceQuit(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })
	lookPath = func(string) (string, error) { return "", errFakeMissing }

	var out bytes.Buffer
	err := EnsureInstalled(strings.NewReader("q\n"), &out, true)
	if err == nil {
		t.Fatal("EnsureInstalled quit: want error")
	}
}

func TestEnsureInstalled_Present(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })
	lookPath = func(name string) (string, error) {
		if name == "claude" {
			return "/fake/claude", nil
		}
		return "", errFakeMissing
	}
	var out bytes.Buffer
	if err := EnsureInstalled(strings.NewReader(""), &out, true); err != nil {
		t.Errorf("EnsureInstalled present: %v", err)
	}
}

func TestEnsureInstalled_NonInteractive(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })
	lookPath = func(string) (string, error) { return "", errFakeMissing }

	var out bytes.Buffer
	err := EnsureInstalled(strings.NewReader(""), &out, false)
	if err == nil {
		t.Fatal("EnsureInstalled non-interactive: want error")
	}
}

func TestEnsureInstalled_NpmMissing(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })
	lookPath = func(string) (string, error) {
		return "", errFakeMissing
	}

	var out bytes.Buffer
	err := EnsureInstalled(strings.NewReader("1\n"), &out, true)
	if err == nil {
		t.Fatal("EnsureInstalled npm-missing: want error")
	}
}

var errFakeMissing = fakeMissingErr("not found")

type fakeMissingErr string

func (f fakeMissingErr) Error() string { return string(f) }
