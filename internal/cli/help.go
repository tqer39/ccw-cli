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

List mode (non-interactive):
  -L, --list           Print ccw worktrees and exit (text table)
  -d, --dir <path>     Target directory for --list (defaults to cwd)
      --json           Emit JSON instead of the text table
      --no-pr          Skip gh PR lookup
      --no-session     Skip session log lookup

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
  2  system error (git failure, etc.)
  *  passthrough from ` + "`claude`" + `

Repository: https://github.com/tqer39/ccw-cli
`

// PrintHelp writes the usage string to w.
func PrintHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, usage)
}
