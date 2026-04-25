// Package namegen generates short slug names like "quick-falcon-7bd2"
// to use as both worktree directory name and Claude Code session name.
package namegen

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

var adjectives = []string{
	"quick", "lazy", "happy", "brave", "calm", "eager", "fancy", "glad",
	"jolly", "kind", "lively", "merry", "nice", "polite", "quiet", "silly",
	"witty", "zany", "bright", "clever", "daring", "fierce", "gentle", "mighty",
	"nimble", "proud", "rapid", "shiny", "sturdy", "tame",
}

var nouns = []string{
	"falcon", "otter", "lion", "tiger", "wolf", "panda", "eagle", "shark",
	"crane", "fox", "raven", "owl", "lynx", "bison", "moose", "hawk",
	"orca", "puma", "yak", "ibex", "robin", "swan", "gecko", "mantis",
	"koala", "badger", "heron", "jaguar", "lemur", "mole",
}

// counter ensures unique seeds when Generate() is called multiple times
// within the same nanosecond (e.g. in tight loops during tests).
var counter atomic.Uint64

// Generate returns a slug like "quick-falcon-7bd2".
// Combines time.Now().UnixNano() with a monotonic counter to guarantee
// unique seeds even when called in rapid succession.
func Generate() string {
	seq := counter.Add(1)
	return generateWithSeed(uint64(time.Now().UnixNano()) ^ (seq * 0x9E3779B97F4A7C15))
}

func generateWithSeed(seed uint64) string {
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	adj := adjectives[r.IntN(len(adjectives))]
	noun := nouns[r.IntN(len(nouns))]
	suffix := fmt.Sprintf("%04x", r.IntN(0x10000))
	return fmt.Sprintf("%s-%s-%s", adj, noun, suffix)
}
