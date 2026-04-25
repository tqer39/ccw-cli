package namegen

import (
	"regexp"
	"testing"
)

func TestGenerate_FormatAndUniqueness(t *testing.T) {
	re := regexp.MustCompile(`^[a-z]+-[a-z]+-[0-9a-f]{4}$`)
	seen := map[string]struct{}{}
	for i := 0; i < 100; i++ {
		got := Generate()
		if !re.MatchString(got) {
			t.Fatalf("Generate() = %q, want match %s", got, re)
		}
		seen[got] = struct{}{}
	}
	if len(seen) < 90 {
		t.Errorf("Generate() collisions too high: %d/100 unique", len(seen))
	}
}

func TestGenerateWithSeed_Deterministic(t *testing.T) {
	a := generateWithSeed(42)
	b := generateWithSeed(42)
	if a != b {
		t.Errorf("generateWithSeed(42): non-deterministic %q vs %q", a, b)
	}
}

func TestGenerate_NoSpacesNoUppercase(t *testing.T) {
	for i := 0; i < 50; i++ {
		got := Generate()
		for _, r := range got {
			if r == ' ' {
				t.Fatalf("Generate() = %q contains space", got)
			}
			if r >= 'A' && r <= 'Z' {
				t.Fatalf("Generate() = %q contains uppercase", got)
			}
		}
	}
}
