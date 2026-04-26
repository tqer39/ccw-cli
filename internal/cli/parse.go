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
	List         bool
	TargetDir    string
	JSON         bool
	NoPR         bool
	NoSession    bool
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
	fs.BoolVarP(&f.List, "list", "L", false, "non-interactive list of ccw worktrees (text by default)")
	fs.StringVarP(&f.TargetDir, "dir", "d", "", "target directory for --list (defaults to cwd)")
	fs.BoolVar(&f.JSON, "json", false, "use JSON output for --list")
	fs.BoolVar(&f.NoPR, "no-pr", false, "skip gh PR lookup for --list")
	fs.BoolVar(&f.NoSession, "no-session", false, "skip session log lookup for --list")

	if err := fs.Parse(pre); err != nil {
		return Flags{}, fmt.Errorf("parse flags: %w", err)
	}
	if args := fs.Args(); len(args) > 0 {
		return Flags{}, fmt.Errorf("unexpected positional arguments: %v (use -- to pass args to claude)", args)
	}
	if f.Superpowers {
		f.NewWorktree = true
	}
	if err := validateCleanAll(&f); err != nil {
		return Flags{}, err
	}
	if err := validateList(f, post); err != nil {
		return Flags{}, err
	}
	f.Passthrough = post
	return f, nil
}

func validateCleanAll(f *Flags) error {
	if !f.CleanAll {
		return nil
	}
	if f.StatusFilter == "" {
		f.StatusFilter = "all"
	}
	switch f.StatusFilter {
	case "all", "pushed", "local-only", "dirty":
	default:
		return fmt.Errorf("--status: invalid value %q (want all|pushed|local-only|dirty)", f.StatusFilter)
	}
	if f.StatusFilter == "dirty" && !f.Force {
		return fmt.Errorf("--status=dirty requires --force")
	}
	return nil
}

func validateList(f Flags, post []string) error {
	if f.TargetDir != "" && !f.List {
		return fmt.Errorf("--dir/-d requires --list/-L")
	}
	if !f.List {
		return nil
	}
	switch {
	case f.NewWorktree:
		return fmt.Errorf("--list cannot be combined with --new")
	case f.Superpowers:
		return fmt.Errorf("--list cannot be combined with --superpowers")
	case f.CleanAll:
		return fmt.Errorf("--list cannot be combined with --clean-all")
	case post != nil:
		return fmt.Errorf("--list does not accept passthrough args after --")
	}
	return nil
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
