// Command ccw launches Claude Code in an isolated git worktree.
//
// Phase 3 status: -h / -v / -n / -s and the picker are at parity with the
// bash implementation. The bash `bin/ccw` is kept as a transitional fallback.
package main

import (
	"fmt"
	"os"

	"github.com/tqer39/ccw-cli/internal/claude"
	"github.com/tqer39/ccw-cli/internal/cli"
	"github.com/tqer39/ccw-cli/internal/gitx"
	"github.com/tqer39/ccw-cli/internal/picker"
	"github.com/tqer39/ccw-cli/internal/superpowers"
	"github.com/tqer39/ccw-cli/internal/ui"
	"github.com/tqer39/ccw-cli/internal/version"
	"github.com/tqer39/ccw-cli/internal/worktree"
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

	if flags.CleanAll {
		return runCleanAll(mainRepo, flags, interactive)
	}

	if err := claude.EnsureInstalled(os.Stdin, os.Stderr, interactive); err != nil {
		ui.Error("%v", err)
		return 1
	}

	if err := os.Chdir(mainRepo); err != nil {
		ui.Error("cd to main repo: %v", err)
		return 1
	}
	_ = gitx.SetOriginHead(mainRepo)

	preamble, err := maybeSuperpowers(flags.Superpowers, interactive)
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

	return runPicker(mainRepo, flags.Passthrough, interactive)
}

func runPicker(mainRepo string, passthrough []string, interactive bool) int {
	for {
		action, sel, bulk, err := picker.Run(mainRepo, interactive, os.Stdin, os.Stderr)
		if err != nil {
			ui.Error("%v", err)
			return 1
		}
		switch action {
		case picker.ActionCancel:
			return 0
		case picker.ActionNew:
			code, err := claude.LaunchNew(mainRepo, "", passthrough)
			if err != nil {
				ui.Error("%v", err)
				return 1
			}
			return code
		case picker.ActionResume:
			code, err := claude.Resume(sel.Path, passthrough)
			if err != nil {
				ui.Error("%v", err)
				return 1
			}
			return code
		case picker.ActionDelete:
			if err := worktree.Remove(mainRepo, sel.Path, sel.ForceDelete); err != nil {
				ui.Error("%v", err)
				return 1
			}
			ui.Success("Removed %s", sel.Path)
		case picker.ActionBulkDelete:
			if code := applyBulkDelete(mainRepo, bulk); code != 0 {
				return code
			}
		}
	}
}

func applyBulkDelete(mainRepo string, bulk picker.BulkDeletion) int {
	errs := 0
	for _, p := range bulk.Paths {
		if err := worktree.Remove(mainRepo, p, bulk.Force); err != nil {
			ui.Error("remove %s: %v", p, err)
			errs++
			continue
		}
		ui.Success("Removed %s", p)
	}
	if errs > 0 {
		return 1
	}
	return 0
}

func runCleanAll(mainRepo string, flags cli.Flags, interactive bool) int {
	infos, err := worktree.List(mainRepo)
	if err != nil {
		ui.Error("list worktrees: %v", err)
		return 1
	}
	if len(infos) == 0 {
		ui.Info("no ccw worktrees to clean.")
		return 0
	}

	filter := statusFilterMap(flags.StatusFilter)
	targets := picker.SelectByStatus(infos, filter)

	if !flags.Force && picker.HasDirty(infos, targets) {
		targets = picker.DropDirty(infos, targets)
		ui.Warn("skipping dirty worktrees (use --force to include)")
	}

	if len(targets) == 0 {
		ui.Info("no worktrees matched the filter.")
		return 0
	}

	if flags.DryRun {
		ui.Info("would remove %d worktree(s):", len(targets))
		for _, i := range targets {
			w := infos[i]
			ui.Info("  %s  (%s)  %s", w.Branch, w.Status, w.Path)
		}
		return 0
	}

	if !flags.AssumeYes {
		if !interactive {
			ui.Error("--clean-all in non-interactive mode requires -y/--yes")
			return 1
		}
		if !confirmCleanAll(infos, targets) {
			ui.Info("aborted.")
			return 0
		}
	}

	bulk := picker.BulkDeletion{
		Paths: make([]string, 0, len(targets)),
		Force: flags.Force,
	}
	for _, i := range targets {
		bulk.Paths = append(bulk.Paths, infos[i].Path)
	}
	return applyBulkDelete(mainRepo, bulk)
}

func statusFilterMap(s string) map[worktree.Status]bool {
	switch s {
	case "pushed":
		return map[worktree.Status]bool{worktree.StatusPushed: true}
	case "local-only":
		return map[worktree.Status]bool{worktree.StatusLocalOnly: true}
	case "dirty":
		return map[worktree.Status]bool{worktree.StatusDirty: true}
	default:
		return nil
	}
}

func confirmCleanAll(infos []worktree.Info, targets []int) bool {
	ui.Info("will remove %d worktree(s):", len(targets))
	for _, i := range targets {
		w := infos[i]
		ui.Info("  %s  (%s)  %s", w.Branch, w.Status, w.Path)
	}
	ok, err := ui.PromptYN(os.Stdin, os.Stderr, "proceed?")
	if err != nil {
		ui.Error("%v", err)
		return false
	}
	return ok
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

func maybeSuperpowers(enabled bool, interactive bool) (string, error) {
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
	return superpowers.Preamble(), nil
}
