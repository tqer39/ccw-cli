// Package cli defines ccw's command-line argument surface.
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
)

// Flags is the parsed representation of ccw's command-line arguments.
type Flags struct {
	Help        bool
	Version     bool
	NewWorktree bool
	Superpowers bool
	Passthrough []string
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

	if err := fs.Parse(pre); err != nil {
		return Flags{}, fmt.Errorf("parse flags: %w", err)
	}
	if args := fs.Args(); len(args) > 0 {
		return Flags{}, fmt.Errorf("unexpected positional arguments: %v (use -- to pass args to claude)", args)
	}
	if f.Superpowers {
		f.NewWorktree = true
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
