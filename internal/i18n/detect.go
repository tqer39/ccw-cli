package i18n

import "strings"

// detectFromEnv resolves a Lang from POSIX locale env vars, in priority order
// LC_ALL > LC_MESSAGES > LANG. Returns LangEN for any non-Japanese tag.
// lookup abstracts os.Getenv so tests can inject a fake environment.
func detectFromEnv(lookup func(string) string) Lang {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		raw := lookup(key)
		if raw == "" {
			continue
		}
		if primaryTag(raw) == "ja" {
			return LangJA
		}
		return LangEN
	}
	return LangEN
}

// primaryTag normalizes a POSIX locale string ("ja_JP.UTF-8@modifier") to its
// primary language tag ("ja"). Empty string and "C"/"POSIX" return "".
func primaryTag(raw string) string {
	s := raw
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '_'); i >= 0 {
		s = s[:i]
	}
	s = strings.ToLower(s)
	if s == "c" || s == "posix" {
		return ""
	}
	return s
}
