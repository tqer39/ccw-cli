package i18n

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed locales/en.yaml locales/ja.yaml
var localesFS embed.FS

var (
	currentLang    = LangEN
	currentCatalog map[string]string
)

// Init resolves the active language from (in priority order):
//
//  1. flagLang (e.g. from --lang); empty means "skip this source".
//  2. CCW_LANG env var.
//  3. POSIX locale (LC_ALL > LC_MESSAGES > LANG).
//  4. Default LangEN.
//
// flagLang is validated strictly — an unknown value returns an error.
// CCW_LANG values that aren't en/ja silently fall through to the locale.
// Init must be called once at startup before any T() lookup.
func Init(flagLang string) error {
	lang, err := resolveLang(flagLang, os.Getenv)
	if err != nil {
		return err
	}
	cat, err := loadCatalog(lang)
	if err != nil {
		return err
	}
	currentLang = lang
	currentCatalog = cat
	return nil
}

// Current returns the language resolved by the most recent Init call.
func Current() Lang { return currentLang }

// T returns the translation for key in the current language, formatted with
// fmt.Sprintf semantics. If the key is unknown, the key string itself is
// returned (fail-soft) so missing translations are visible but harmless.
func T(key Key, args ...any) string {
	tmpl, ok := currentCatalog[string(key)]
	if !ok {
		return string(key)
	}
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}

func resolveLang(flagLang string, lookup func(string) string) (Lang, error) {
	if flagLang != "" {
		l, ok := parseLang(flagLang)
		if !ok {
			return "", fmt.Errorf("invalid --lang value: %q (expected 'en' or 'ja')", flagLang)
		}
		return l, nil
	}
	if env := lookup("CCW_LANG"); env != "" {
		if l, ok := parseLang(env); ok {
			return l, nil
		}
	}
	return detectFromEnv(lookup), nil
}

func parseLang(s string) (Lang, bool) {
	switch strings.ToLower(s) {
	case "en":
		return LangEN, true
	case "ja":
		return LangJA, true
	}
	return "", false
}

// loadCatalog parses the embedded YAML for lang into a flat map keyed by
// the dot-joined path ("picker.action.menu" etc.). Returns an error if the
// file is missing or malformed.
func loadCatalog(lang Lang) (map[string]string, error) {
	path := "locales/" + string(lang) + ".yaml"
	data, err := localesFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]string)
	flatten("", raw, out)
	return out, nil
}

// flatten walks a nested map[string]any and writes leaf string values into
// dst keyed by the dot-joined path. Non-string scalars cause a panic at
// startup, which surfaces YAML mistakes immediately during tests.
func flatten(prefix string, node map[string]any, dst map[string]string) {
	for k, v := range node {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch t := v.(type) {
		case string:
			dst[key] = t
		case map[string]any:
			flatten(key, t, dst)
		case map[any]any:
			converted := make(map[string]any, len(t))
			for kk, vv := range t {
				converted[fmt.Sprint(kk)] = vv
			}
			flatten(key, converted, dst)
		default:
			panic(fmt.Sprintf("i18n: unsupported YAML value at %q: %T", key, v))
		}
	}
}
