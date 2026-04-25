// Package tips provides short rotating tip strings shown in the picker footer.
package tips

import "math/rand/v2"

var defaults = []string{
	"worktree 名 = session 名。手で /rename しても ccw は何もしません",
	"claude --from-pr <番号> で PR 連携セッションを直接 resume できます",
	"--clean-all で push 済 worktree を一括削除",
	"ccw -- --model <id> で claude にフラグを素通し",
	"picker の RESUME バッジは ~/.claude/projects/ から判定しています",
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
