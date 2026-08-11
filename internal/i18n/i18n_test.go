package i18n

import "testing"

// TestCatalogParity is the CI regression guard: every key in the English
// source of truth must have a Spanish translation, and Spanish must not
// invent keys that English lacks.
func TestCatalogParity(t *testing.T) {
	for key := range english {
		if _, ok := spanish[key]; !ok {
			t.Errorf("key %q missing in spanish catalog", key)
		}
	}
	for key := range spanish {
		if _, ok := english[key]; !ok {
			t.Errorf("key %q in spanish catalog has no english counterpart", key)
		}
	}
}

func TestFallbackToEnglish(t *testing.T) {
	// A Spanish catalog entry deleted on purpose must fall back to English.
	delete(spanish, "validate.ok")
	t.Cleanup(func() { spanish["validate.ok"] = "Las migraciones son válidas." })

	SetLanguage("es")
	if got := T("validate.ok"); got != english["validate.ok"] {
		t.Errorf("expected English fallback, got %q", got)
	}
	SetLanguage("en")
}

func TestTnPluralSelection(t *testing.T) {
	SetLanguage("en")
	if got := Tn("migrate.applied", 1); got != "\n1 migration applied successfully.\n" {
		t.Errorf("singular: got %q", got)
	}
	if got := Tn("migrate.applied", 3); got != "\n3 migrations applied successfully.\n" {
		t.Errorf("plural: got %q", got)
	}

	SetLanguage("es")
	if got := Tn("migrate.applied", 1); got != "\n1 migración aplicada correctamente.\n" {
		t.Errorf("singular es: got %q", got)
	}
	SetLanguage("en")
}

func TestUnknownKeyIsVisible(t *testing.T) {
	if got := T("no.such.key"); got != "‹no.such.key›" {
		t.Errorf("got %q", got)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"es":          "es",
		"ES":          "es",
		"es_PE.UTF-8": "es",
		"es-MX":       "es",
		"en":          "en",
		"en_US.UTF-8": "en",
		"fr_FR":       "en", // unsupported → English
		"":            "en",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolvePriority(t *testing.T) {
	// Clear the environment so only the variables under test participate.
	t.Setenv("GOFORGE_LANG", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "")

	if got := Resolve("es", "en"); got != "es" {
		t.Errorf("explicit flag must win: got %q", got)
	}
	t.Setenv("GOFORGE_LANG", "es")
	if got := Resolve("", "en"); got != "es" {
		t.Errorf("GOFORGE_LANG beats yaml: got %q", got)
	}
	t.Setenv("GOFORGE_LANG", "")
	if got := Resolve("", "es"); got != "es" {
		t.Errorf("yaml beats system locale: got %q", got)
	}
	t.Setenv("LC_ALL", "es_PE.UTF-8")
	if got := Resolve("", ""); got != "es" {
		t.Errorf("LC_ALL: got %q", got)
	}
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "es_ES.UTF-8")
	if got := Resolve("", ""); got != "es" {
		t.Errorf("LANG: got %q", got)
	}
	t.Setenv("LANG", "")
	if got := Resolve("", ""); got != "en" {
		t.Errorf("default must be en: got %q", got)
	}
}
