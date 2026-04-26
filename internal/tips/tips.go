// Package tips provides short rotating tip strings shown in the picker footer.
package tips

import (
	"math/rand/v2"

	"github.com/tqer39/ccw-cli/internal/i18n"
)

var keys = []i18n.Key{
	i18n.KeyTipRename,
	i18n.KeyTipFromPR,
	i18n.KeyTipCleanAll,
	i18n.KeyTipPassthrough,
	i18n.KeyTipResumeBadge,
}

// Defaults returns the current language's tip strings.
func Defaults() []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = i18n.T(k)
	}
	return out
}

// PickRandom returns a single tip selected deterministically from seed.
func PickRandom(seed uint64) string {
	return pickFrom(Defaults(), seed)
}

func pickFrom(set []string, seed uint64) string {
	if len(set) == 0 {
		return ""
	}
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	return set[r.IntN(len(set))]
}
