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
