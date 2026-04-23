// Command ccw launches Claude Code in an isolated git worktree.
//
// Phase 2 status: -n / -s の直起動パスが機能。無印起動 (picker) は
// Phase 3 で実装予定で、それまでは "未実装" を出して exit 1。
// bash 版 bin/ccw は picker が必要なユースケース向けに温存されている。
package main

import (
	"fmt"
	"os"

	"github.com/tqer39/ccw-cli/internal/claude"
	"github.com/tqer39/ccw-cli/internal/cli"
	"github.com/tqer39/ccw-cli/internal/gitx"
	"github.com/tqer39/ccw-cli/internal/superpowers"
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

	os.Exit(run(flags))
}

func run(flags cli.Flags) int {
	mainRepo, err := resolveMainRepo()
	if err != nil {
		return 1
	}

	ui.EnsureTool("git", "Install from https://git-scm.com/downloads")

	interactive := ui.IsInteractive(os.Stdin) && ui.IsInteractive(os.Stdout)

	if err := claude.EnsureInstalled(os.Stdin, os.Stderr, interactive); err != nil {
		ui.Error("%v", err)
		return 1
	}

	if err := os.Chdir(mainRepo); err != nil {
		ui.Error("cd to main repo: %v", err)
		return 1
	}
	_ = gitx.SetOriginHead(mainRepo)

	preamble, err := maybeSuperpowers(flags.Superpowers, mainRepo, interactive)
	if err != nil {
		ui.Error("%v", err)
		return 1
	}

	if flags.NewWorktree {
		code, err := claude.LaunchNew(mainRepo, preamble, flags.Passthrough)
		if err != nil {
			ui.Error("%v", err)
			return 1
		}
		return code
	}

	ui.Error("Phase 2: picker 未実装。`-n` を指定するか bash 版 bin/ccw を使用してください。")
	return 1
}

func resolveMainRepo() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		ui.Error("getwd: %v", err)
		return "", fmt.Errorf("getwd: %w", err)
	}
	if err := gitx.RequireRepo(cwd); err != nil {
		ui.Warn("ccw must be run inside a git repository.")
		ui.Info("  current directory: %s", cwd)
		ui.Info("  hint: cd into an existing repo, or run `git init` to create one.")
		return "", fmt.Errorf("require repo: %w", err)
	}
	mainRepo, err := gitx.ResolveMainRepo(cwd)
	if err != nil {
		ui.Error("failed to resolve main repository root: %v", err)
		return "", fmt.Errorf("resolve main repo: %w", err)
	}
	return mainRepo, nil
}

func maybeSuperpowers(enabled bool, mainRepo string, interactive bool) (string, error) {
	if !enabled {
		return "", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve HOME: %w", err)
	}
	if err := superpowers.EnsureInstalled(os.Stdin, os.Stderr, home, interactive); err != nil {
		return "", fmt.Errorf("superpowers install: %w", err)
	}
	if err := superpowers.EnsureGitignore(os.Stdin, os.Stderr, mainRepo, interactive); err != nil {
		return "", fmt.Errorf("superpowers gitignore: %w", err)
	}
	return superpowers.Preamble(), nil
}
