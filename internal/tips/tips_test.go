package tips

import (
	"os"
	"strings"
	"testing"

	"github.com/tqer39/ccw-cli/internal/i18n"
)

func TestMain(m *testing.M) {
	if err := i18n.Init("en"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestPickRandom_FromDefaultSet(t *testing.T) {
	got := PickRandom(42)
	if got == "" {
		t.Fatal("PickRandom(42) = empty string")
	}
	found := false
	for _, c := range Defaults() {
		if got == c {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("PickRandom(42) = %q, not in Defaults()", got)
	}
}

func TestPickRandom_Deterministic(t *testing.T) {
	a, b := PickRandom(7), PickRandom(7)
	if a != b {
		t.Errorf("PickRandom(7) is non-deterministic: %q != %q", a, b)
	}
}

func TestPickFrom_Empty(t *testing.T) {
	if got := pickFrom(nil, 1); got != "" {
		t.Errorf("pickFrom(nil) = %q, want empty", got)
	}
	if got := pickFrom([]string{}, 1); got != "" {
		t.Errorf("pickFrom([]) = %q, want empty", got)
	}
}

func TestDefaults_NonEmpty(t *testing.T) {
	d := Defaults()
	if len(d) == 0 {
		t.Error("Defaults() empty")
	}
	for _, s := range d {
		if strings.TrimSpace(s) == "" {
			t.Errorf("empty TIPS string in defaults")
		}
	}
}
