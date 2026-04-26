package superpowers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePluginDir_AllMiss(t *testing.T) {
	home := t.TempDir()
	got, ok := resolvePluginDirIn(home)
	if ok {
		t.Fatalf("expected miss, got %q", got)
	}
}

func TestResolvePluginDir_WellKnownHit(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest", ".claude-plugin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := resolvePluginDirIn(home)
	if !ok {
		t.Fatal("expected hit, got miss")
	}
	want := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolvePluginDir_GlobHit(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "plugins", "cache", "third-party", "superpowers", "latest", ".claude-plugin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := resolvePluginDirIn(home)
	if !ok {
		t.Fatal("expected hit, got miss")
	}
	want := filepath.Join(home, ".claude", "plugins", "cache", "third-party", "superpowers", "latest")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolvePluginDir_InstalledJSONHit(t *testing.T) {
	home := t.TempDir()
	cacheDir := filepath.Join(home, ".claude", "plugins", "cache", "my-marketplace", "superpowers", "v2", ".claude-plugin")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	pluginsDir := filepath.Join(home, ".claude", "plugins")
	jsonBody := `{
		"version": 2,
		"plugins": {
			"superpowers@my-marketplace": [
				{"scope": "user", "installPath": "", "version": "v2"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(jsonBody), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := resolvePluginDirIn(home)
	if !ok {
		t.Fatal("expected hit, got miss")
	}
	want := filepath.Join(home, ".claude", "plugins", "cache", "my-marketplace", "superpowers", "v2")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolvePluginDir_InstalledJSONExplicitInstallPath(t *testing.T) {
	home := t.TempDir()
	customDir := filepath.Join(home, "custom", "abs", "superpowers-checkout", ".claude-plugin")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customDir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	pluginsDir := filepath.Join(home, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(customDir)
	jsonBody := `{
		"version": 2,
		"plugins": {
			"superpowers@local": [
				{"scope": "project", "installPath": "` + parent + `", "version": "dev"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(jsonBody), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := resolvePluginDirIn(home)
	if !ok {
		t.Fatal("expected hit, got miss")
	}
	if got != parent {
		t.Errorf("got %q, want %q", got, parent)
	}
}
