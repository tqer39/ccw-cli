// Package cli defines ccw's command-line argument surface.
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
)

// Flags is the parsed representation of ccw's command-line arguments.
type Flags struct {
	Help         bool
	Version      bool
	NewWorktree  bool
	Superpowers  bool
	CleanAll     bool
	StatusFilter string
	Force        bool
	DryRun       bool
	AssumeYes    bool
	Lang         string
	Passthrough  []string
}

// Parse interprets argv (without the program name) and returns Flags.
// Unknown flags, positional args, and removed flags (--update / --uninstall)
// return a non-nil error.
func Parse(argv []string) (Flags, error) {
	pre, post := splitAtDoubleDash(argv)

	fs := pflag.NewFlagSet("ccw", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var f Flags
	fs.BoolVarP(&f.Help, "help", "h", false, "show help")
	fs.BoolVarP(&f.Version, "version", "v", false, "show version")
	fs.BoolVarP(&f.NewWorktree, "new", "n", false, "always start a new worktree")
	fs.BoolVarP(&f.Superpowers, "superpowers", "s", false, "inject superpowers preamble (implies --new)")
	fs.BoolVar(&f.CleanAll, "clean-all", false, "bulk delete worktrees (non-interactive unless --force/--yes)")
	fs.StringVar(&f.StatusFilter, "status", "", "status filter: all | pushed | local-only | dirty (default all)")
	fs.BoolVar(&f.Force, "force", false, "allow --force removal of dirty worktrees")
	fs.BoolVar(&f.DryRun, "dry-run", false, "list targets without deleting")
	fs.BoolVarP(&f.AssumeYes, "yes", "y", false, "skip confirmation prompts (--clean-all, -s plugin install)")
	fs.StringVar(&f.Lang, "lang", "", "force output language: en | ja")

	if err := fs.Parse(pre); err != nil {
		return Flags{}, fmt.Errorf("parse flags: %w", err)
	}
	if args := fs.Args(); len(args) > 0 {
		return Flags{}, fmt.Errorf("unexpected positional arguments: %v (use -- to pass args to claude)", args)
	}
	if f.Superpowers {
		f.NewWorktree = true
	}
	if f.CleanAll {
		if f.StatusFilter == "" {
			f.StatusFilter = "all"
		}
		switch f.StatusFilter {
		case "all", "pushed", "local-only", "dirty":
		default:
			return Flags{}, fmt.Errorf("--status: invalid value %q (want all|pushed|local-only|dirty)", f.StatusFilter)
		}
		if f.StatusFilter == "dirty" && !f.Force {
			return Flags{}, fmt.Errorf("--status=dirty requires --force")
		}
	}
	f.Passthrough = post
	return f, nil
}

// splitAtDoubleDash returns (before, after) around the first bare "--" token.
// If "--" is absent, after is nil and before is argv. If "--" is present,
// after is a non-nil (possibly empty) slice to let callers distinguish it.
func splitAtDoubleDash(argv []string) (before, after []string) {
	for i, a := range argv {
		if a == "--" {
			return argv[:i], append([]string{}, argv[i+1:]...)
		}
	}
	return argv, nil
}
