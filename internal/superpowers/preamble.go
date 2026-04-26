// Package superpowers provides the localized preamble that `ccw -s` injects
// as the first user prompt to Claude. Plugin installation is handled out of
// band via .claude/settings.json `enabledPlugins`.
package superpowers

import (
	_ "embed"

	"github.com/tqer39/ccw-cli/internal/i18n"
)

//go:embed preamble_en.txt
var preambleEN string

//go:embed preamble_ja.txt
var preambleJA string

// Preamble returns the preamble text for lang. Unknown values fall back to English.
func Preamble(lang i18n.Lang) string {
	if lang == i18n.LangJA {
		return preambleJA
	}
	return preambleEN
}
