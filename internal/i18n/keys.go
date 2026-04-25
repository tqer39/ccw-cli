// Package i18n provides language detection and translated message lookup
// for ccw's user-facing output (TUI, tips, help, CLI messages).
package i18n

// Lang identifies an active language. Only "en" and "ja" are supported.
type Lang string

// Supported language identifiers.
const (
	LangEN Lang = "en"
	LangJA Lang = "ja"
)

// Key is a stable identifier for a translatable message. Values match the
// dot-flattened path inside the locale YAML files.
type Key string
