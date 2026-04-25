package i18n

import "testing"

func TestDetectFromEnv(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want Lang
	}{
		{"empty defaults to EN", map[string]string{}, LangEN},
		{"LANG ja_JP.UTF-8", map[string]string{"LANG": "ja_JP.UTF-8"}, LangJA},
		{"LANG en_US.UTF-8", map[string]string{"LANG": "en_US.UTF-8"}, LangEN},
		{"LANG ja", map[string]string{"LANG": "ja"}, LangJA},
		{"LANG fr_FR.UTF-8 falls back to EN", map[string]string{"LANG": "fr_FR.UTF-8"}, LangEN},
		{"LC_MESSAGES wins over LANG", map[string]string{"LC_MESSAGES": "ja_JP", "LANG": "en_US"}, LangJA},
		{"LC_ALL wins over LC_MESSAGES", map[string]string{"LC_ALL": "ja_JP", "LC_MESSAGES": "en_US", "LANG": "en_US"}, LangJA},
		{"empty LC_ALL falls through", map[string]string{"LC_ALL": "", "LANG": "ja_JP"}, LangJA},
		{"modifier suffix stripped", map[string]string{"LANG": "ja_JP.UTF-8@bidi"}, LangJA},
		{"C locale defaults to EN", map[string]string{"LANG": "C"}, LangEN},
		{"POSIX locale defaults to EN", map[string]string{"LANG": "POSIX"}, LangEN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lookup := func(k string) string { return tc.env[k] }
			if got := detectFromEnv(lookup); got != tc.want {
				t.Errorf("detectFromEnv(%v) = %q, want %q", tc.env, got, tc.want)
			}
		})
	}
}
