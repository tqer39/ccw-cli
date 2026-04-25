package i18n

import "testing"

func TestLoadCatalog_BothLanguagesParse(t *testing.T) {
	en, err := loadCatalog(LangEN)
	if err != nil {
		t.Fatalf("loadCatalog(EN) error: %v", err)
	}
	if got := en["meta.smoke"]; got != "english smoke value" {
		t.Errorf("en[meta.smoke] = %q, want %q", got, "english smoke value")
	}
	ja, err := loadCatalog(LangJA)
	if err != nil {
		t.Fatalf("loadCatalog(JA) error: %v", err)
	}
	if got := ja["meta.smoke"]; got != "日本語スモーク値" {
		t.Errorf("ja[meta.smoke] = %q, want %q", got, "日本語スモーク値")
	}
}

func TestInit_FlagWins(t *testing.T) {
	t.Setenv("LC_ALL", "ja_JP.UTF-8")
	t.Setenv("CCW_LANG", "ja")
	if err := Init("en"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Current() != LangEN {
		t.Errorf("Current() = %q, want en", Current())
	}
}

func TestInit_EnvWinsOverLocale(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("CCW_LANG", "ja")
	if err := Init(""); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Current() != LangJA {
		t.Errorf("Current() = %q, want ja", Current())
	}
}

func TestInit_LocaleFallback(t *testing.T) {
	t.Setenv("LC_ALL", "ja_JP.UTF-8")
	t.Setenv("CCW_LANG", "")
	if err := Init(""); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Current() != LangJA {
		t.Errorf("Current() = %q, want ja", Current())
	}
}

func TestInit_DefaultEN(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
	t.Setenv("CCW_LANG", "")
	if err := Init(""); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Current() != LangEN {
		t.Errorf("Current() = %q, want en", Current())
	}
}

func TestInit_InvalidFlagReturnsError(t *testing.T) {
	if err := Init("fr"); err == nil {
		t.Error("Init(\"fr\") returned nil, want error")
	}
}

func TestInit_InvalidEnvFallsThrough(t *testing.T) {
	t.Setenv("CCW_LANG", "fr")
	t.Setenv("LC_ALL", "ja_JP.UTF-8")
	if err := Init(""); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Current() != LangJA {
		t.Errorf("Current() = %q, want ja (env fr should fall through to locale)", Current())
	}
}

func TestT_ReturnsTranslation(t *testing.T) {
	t.Setenv("CCW_LANG", "")
	if err := Init("ja"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := T("meta.smoke"); got != "日本語スモーク値" {
		t.Errorf("T(meta.smoke) = %q, want 日本語スモーク値", got)
	}
}

func TestT_FormatsArgs(t *testing.T) {
	if err := Init("en"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	currentCatalog["test.fmt"] = "hello %s, count=%d"
	defer delete(currentCatalog, "test.fmt")
	if got := T("test.fmt", "world", 7); got != "hello world, count=7" {
		t.Errorf("T(test.fmt, world, 7) = %q", got)
	}
}

func TestT_UnknownKeyReturnsKey(t *testing.T) {
	if err := Init("en"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := T("does.not.exist"); got != "does.not.exist" {
		t.Errorf("T(missing) = %q, want %q", got, "does.not.exist")
	}
}
