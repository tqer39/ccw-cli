package cli

import (
	"fmt"
	"io"
)

const usage = `Usage: ccw [options] [-- <claude-args>...]

Options:
  -n, --new            Always start a new worktree (skip picker)
  -s, --superpowers    Inject superpowers preamble (implies -n)
  -v, --version        Show version
  -h, --help           Show this help

Bulk delete:
      --clean-all        Bulk delete mode
      --status=<filter>  all | pushed | local-only | dirty (default: all)
      --force            Delete dirty worktrees with --force
      --dry-run          Print targets and exit
  -y, --yes              Skip confirmation prompts (--clean-all, -s plugin install)

Arguments after ` + "`--`" + ` are forwarded to ` + "`claude`" + ` verbatim.

Environment:
  NO_COLOR=1           Disable colored output
  CCW_DEBUG=1          Verbose debug logging

Exit codes:
  0  success
  1  user error / cancellation
  *  passthrough from ` + "`claude`" + `

Repository: https://github.com/tqer39/ccw-cli
`

// PrintHelp writes the usage string to w.
func PrintHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, usage)
}
