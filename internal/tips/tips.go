// Package tips provides short rotating tip strings shown in the picker footer.
package tips

import "math/rand/v2"

var defaults = []string{
	"Worktree name = session name; renaming with /rename is fine, ccw doesn't track it",
	"claude --from-pr <number> resumes a PR-linked session directly",
	"--clean-all sweeps pushed worktrees in bulk",
	"ccw -- --model <id> passes flags through to claude",
	"The RESUME badge is derived from ~/.claude/projects/",
}

// Defaults returns a copy of the built-in TIPS set.
func Defaults() []string {
	out := make([]string, len(defaults))
	copy(out, defaults)
	return out
}

// PickRandom returns a single tip selected deterministically from seed.
func PickRandom(seed uint64) string {
	return pickFrom(defaults, seed)
}

func pickFrom(set []string, seed uint64) string {
	if len(set) == 0 {
		return ""
	}
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	return set[r.IntN(len(set))]
}
