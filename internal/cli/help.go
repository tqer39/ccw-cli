package cli

import (
	"fmt"
	"io"

	"github.com/tqer39/ccw-cli/internal/i18n"
)

// PrintHelp writes the localized usage string to w.
func PrintHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, i18n.T(i18n.KeyHelpUsage))
}
