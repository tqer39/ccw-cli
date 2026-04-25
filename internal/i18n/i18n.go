package i18n

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed locales/en.yaml locales/ja.yaml
var localesFS embed.FS

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
