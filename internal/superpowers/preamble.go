// Package superpowers handles the optional superpowers plugin preamble,
// plugin presence detection, and .gitignore augmentation used by `ccw -s`.
package superpowers

import _ "embed"

//go:embed preamble.txt
var preambleText string

// Preamble returns the preamble string injected as the first claude message
// when `-s` is passed.
func Preamble() string {
	return preambleText
}
