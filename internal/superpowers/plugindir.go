package superpowers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolvePluginDir returns the on-disk directory of the superpowers plugin
// suitable for `claude --plugin-dir <path>`. Returns ("", false) if no
// candidate location can be confirmed.
//
// Resolution cascade:
//
//  1. Read ~/.claude/plugins/installed_plugins.json, find the first key
//     matching `superpowers@<marketplace>`, and use its explicit installPath
//     when set, otherwise derive
//     ~/.claude/plugins/cache/<marketplace>/superpowers/<version>.
//  2. Try the well-known path
//     ~/.claude/plugins/cache/claude-plugins-official/superpowers/latest.
//  3. Glob ~/.claude/plugins/cache/*/superpowers/latest/.claude-plugin/plugin.json
//     and return the alphabetically first hit.
//
// Each candidate is validated by checking that
// <candidate>/.claude-plugin/plugin.json exists.
func ResolvePluginDir() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", false
	}
	return resolvePluginDirIn(home)
}

func resolvePluginDirIn(home string) (string, bool) {
	if p, ok := lookupFromInstalledJSON(home); ok {
		return p, true
	}
	if p, ok := lookupWellKnown(home); ok {
		return p, true
	}
	if p, ok := lookupViaGlob(home); ok {
		return p, true
	}
	return "", false
}

func validateCandidate(dir string) bool {
	if dir == "" {
		return false
	}
	manifest := filepath.Join(dir, ".claude-plugin", "plugin.json")
	st, err := os.Stat(manifest)
	if err != nil || st.IsDir() {
		return false
	}
	return true
}

type installedPluginEntry struct {
	Scope       string `json:"scope"`
	InstallPath string `json:"installPath"`
	Version     string `json:"version"`
}

type installedPluginsFile struct {
	Plugins map[string][]installedPluginEntry `json:"plugins"`
}

func lookupFromInstalledJSON(home string) (string, bool) {
	path := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var doc installedPluginsFile
	if err := json.Unmarshal(data, &doc); err != nil {
		return "", false
	}
	for key, entries := range doc.Plugins {
		marketplace, ok := superpowersMarketplace(key)
		if !ok {
			continue
		}
		if len(entries) == 0 {
			continue
		}
		entry := entries[0]
		if entry.InstallPath != "" {
			if validateCandidate(entry.InstallPath) {
				return entry.InstallPath, true
			}
			continue
		}
		version := entry.Version
		if version == "" {
			version = "latest"
		}
		dir := filepath.Join(home, ".claude", "plugins", "cache", marketplace, "superpowers", version)
		if validateCandidate(dir) {
			return dir, true
		}
	}
	return "", false
}

func superpowersMarketplace(key string) (string, bool) {
	const prefix = "superpowers@"
	if !strings.HasPrefix(key, prefix) {
		return "", false
	}
	mp := strings.TrimPrefix(key, prefix)
	if mp == "" {
		return "", false
	}
	return mp, true
}

func lookupWellKnown(home string) (string, bool) {
	dir := filepath.Join(home, ".claude", "plugins", "cache", "claude-plugins-official", "superpowers", "latest")
	if validateCandidate(dir) {
		return dir, true
	}
	return "", false
}

func lookupViaGlob(home string) (string, bool) {
	pattern := filepath.Join(home, ".claude", "plugins", "cache", "*", "superpowers", "latest", ".claude-plugin", "plugin.json")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", false
	}
	sort.Strings(matches)
	for _, m := range matches {
		dir := filepath.Dir(filepath.Dir(m))
		if validateCandidate(dir) {
			return dir, true
		}
	}
	return "", false
}
