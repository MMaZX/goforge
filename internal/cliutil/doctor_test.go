package cliutil

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MMaZX/goforge/internal/i18n"
)

func doctorReport() DoctorReport {
	return DoctorReport{
		Healthy: false,
		Checks: []DoctorCheck{
			{Name: "Config", OK: true, Detail: "goforge.yaml (driver: PostgreSQL, migrations: ./migrations)"},
			{Name: ".env", OK: true, Detail: "found at .env"},
			{Name: "Migrations directory", OK: true, Detail: "./migrations (2 found)"},
			{Name: "Database connection", OK: false, Detail: "connection refused"},
			{Name: "Migration history", Skipped: true, Detail: "database connection failed"},
			{Name: "Locking", Skipped: true, Detail: "database connection failed"},
		},
	}
}

func TestPrintDoctorHumanGolden(t *testing.T) {
	t.Run("en", func(t *testing.T) {
		i18n.SetLanguage("en")
		t.Cleanup(func() { i18n.SetLanguage("en") })
		var buf bytes.Buffer
		PrintDoctorHuman(&buf, doctorReport())
		want := "GoForge doctor\n\n" +
			"✓ Config\n    goforge.yaml (driver: PostgreSQL, migrations: ./migrations)\n" +
			"✓ .env\n    found at .env\n" +
			"✓ Migrations directory\n    ./migrations (2 found)\n" +
			"✗ Database connection\n    connection refused\n" +
			"… Migration history\n    database connection failed\n" +
			"… Locking\n    database connection failed\n\n" +
			"Some checks failed.\n"
		if got := buf.String(); got != want {
			t.Errorf("en output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("es", func(t *testing.T) {
		i18n.SetLanguage("es")
		t.Cleanup(func() { i18n.SetLanguage("en") })
		var buf bytes.Buffer
		PrintDoctorHuman(&buf, doctorReport())
		want := "GoForge doctor\n\n" +
			"✓ Configuración\n    goforge.yaml (driver: PostgreSQL, migrations: ./migrations)\n" +
			"✓ .env\n    found at .env\n" +
			"✓ Directorio de migraciones\n    ./migrations (2 found)\n" +
			"✗ Conexión a la base de datos\n    connection refused\n" +
			"… Historial de migraciones\n    database connection failed\n" +
			"… Bloqueo (lock)\n    database connection failed\n\n" +
			"Algunas verificaciones fallaron.\n"
		if got := buf.String(); got != want {
			t.Errorf("es output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})
}

func TestPrintDoctorHumanHealthy(t *testing.T) {
	report := DoctorReport{Healthy: true, Checks: []DoctorCheck{{Name: "Config", OK: true}}}

	i18n.SetLanguage("es")
	t.Cleanup(func() { i18n.SetLanguage("en") })
	var buf bytes.Buffer
	PrintDoctorHuman(&buf, report)
	if !strings.Contains(buf.String(), "Todas las verificaciones pasaron.") {
		t.Errorf("expected Spanish healthy message, got:\n%s", buf.String())
	}
}

// TestDoctorJSONNotTranslated guards the design rule: Name/Detail feed the
// --json report and must stay in English no matter the active language, so
// the machine contract is stable. Only the human renderer translates.
func TestDoctorJSONNotTranslated(t *testing.T) {
	report := doctorReport()

	en, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	i18n.SetLanguage("es")
	t.Cleanup(func() { i18n.SetLanguage("en") })
	es, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(en, es) {
		t.Errorf("JSON report changed with language:\nen: %s\nes: %s", en, es)
	}
	if strings.Contains(string(es), "Conexión a la base de datos") {
		t.Errorf("JSON report must keep English check names: %s", es)
	}
}
