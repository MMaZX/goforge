// Package i18n provides GoForge's bilingual (English/Spanish) message
// catalogs for human-facing CLI output.
//
// Scope rules (see plan_traduccion.go):
//   - Only text GoForge prints itself is translated. Cobra's help
//     templates, --json output, engine errors, and migration names are
//     intentionally left in English.
//   - The language is resolved once at startup (Rule A): --lang flag >
//     GOFORGE_LANG > language: in goforge.yaml > LC_ALL/LANG > "en".
package i18n

import (
	"fmt"
	"os"
	"strings"
)

// Catalog maps message keys to translated format strings.
type Catalog map[string]string

// active is the catalog used by T/Tn. It defaults to English and is
// replaced once at startup via SetLanguage, before any command prints
// output. The CLI is single-threaded at that point, so no locking is
// needed.
var active = english

// T returns the translation of key, formatted with args (fmt.Sprintf
// semantics). Missing keys render as "‹key›" so an untranslated string
// is visible in output instead of failing silently.
func T(key string, args ...any) string {
	msg, ok := active[key]
	if !ok {
		return "‹" + key + "›"
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

// Tn is T with plural selection: it uses key+"_one" when n == 1 and
// key+"_other" otherwise. n is prepended to args, so the format string
// can reference the count as its first verb.
func Tn(key string, n int, args ...any) string {
	if n == 1 {
		return T(key+"_one", append([]any{n}, args...)...)
	}
	return T(key+"_other", append([]any{n}, args...)...)
}

// SetLanguage activates the catalog for lang. Non-English catalogs are
// merged over English, so a key missing its translation falls back to
// the English text instead of breaking output. Anything that is not
// recognized as Spanish resolves to English.
func SetLanguage(lang string) {
	merged := make(Catalog, len(english))
	for k, v := range english {
		merged[k] = v
	}
	if Normalize(lang) == "es" {
		for k, v := range spanish {
			merged[k] = v
		}
	}
	active = merged
}

// Normalize reduces a locale string ("es_PE.UTF-8", "en-US", "es") to
// "es" or "en".
func Normalize(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if i := strings.IndexAny(locale, "_-.@"); i >= 0 {
		locale = locale[:i]
	}
	if locale == "es" {
		return "es"
	}
	return "en"
}

// Resolve picks the display language following Rule A of the i18n plan:
//
//	explicit (--lang) > GOFORGE_LANG > fromConfig (goforge.yaml) >
//	LC_ALL > LANG > "en"
//
// fromConfig may be empty when no config was loaded (e.g. goforge init,
// goforge version, or a broken goforge.yaml).
func Resolve(explicit, fromConfig string) string {
	for _, candidate := range []string{
		explicit,
		os.Getenv("GOFORGE_LANG"),
		fromConfig,
		os.Getenv("LC_ALL"),
		os.Getenv("LANG"),
	} {
		if strings.TrimSpace(candidate) != "" {
			return Normalize(candidate)
		}
	}
	return "en"
}
