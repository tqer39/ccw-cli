// Command ccw launches Claude Code in an isolated git worktree.
//
// Phase 1 status: only -h / -v are functional. Other flag combinations
// print a "not implemented" message and exit 1. Continue to use the
// bash implementation at bin/ccw for day-to-day work.
package main

import (
	"fmt"
	"os"

	"github.com/tqer39/ccw-cli/internal/cli"
	"github.com/tqer39/ccw-cli/internal/ui"
	"github.com/tqer39/ccw-cli/internal/version"
)

func main() {
	ui.InitColor()

	flags, err := cli.Parse(os.Args[1:])
	if err != nil {
		ui.Error("%v", err)
		cli.PrintHelp(os.Stderr)
		os.Exit(2)
	}

	if flags.Help {
		cli.PrintHelp(os.Stdout)
		return
	}
	if flags.Version {
		fmt.Println(version.String())
		return
	}

	ui.Error("Phase 1 スケルトンのため、-n / -s / picker は未実装です。bash 版 bin/ccw を使用してください。")
	os.Exit(1)
}
