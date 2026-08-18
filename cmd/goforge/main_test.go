package main

import (
	"bytes"
	"os"
	"testing"
)

// runCLI executes the root command with args and returns captured stdout,
// mimicking cobra's Execute wiring without main's exit handling.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runCLIWithIn(t, bytes.NewBuffer(nil), args...)
}

func runCLIWithIn(t *testing.T, in *bytes.Buffer, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	root.SetArgs(args)
	var out bytes.Buffer
	root.SetIn(in)
	root.SetOut(&out)
	root.SetErr(&out)
	err := root.Execute()
	return out.String(), err
}

// inTempDir runs fn with the working directory set to a fresh temp dir.
func inTempDir(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	fn(dir)
}

func TestInitHumanLocalized(t *testing.T) {
	inTempDir(t, func(_ string) {
		out, err := runCLI(t, "init", "--lang", "es")
		if err != nil {
			t.Fatalf("init: %v", err)
		}
		want := "Se crearon goforge.yaml, ./migrations y ./.env\n" +
			"Driver: PostgreSQL\n\n" +
			"Edita ./.env y reemplaza CHANGE_USER, CHANGE_PASSWORD, CHANGE_HOST y CHANGE_DATABASE con tus credenciales reales:\n" +
			"  DATABASE_URL=postgres://CHANGE_USER:CHANGE_PASSWORD@CHANGE_HOST:5432/CHANGE_DATABASE?sslmode=disable\n"
		if out != want {
			t.Errorf("es init output mismatch:\ngot:\n%s\nwant:\n%s", out, want)
		}

		// Second run: pure business error must be translated too.
		_, err = runCLI(t, "init", "--lang", "es")
		if err == nil || err.Error() != "goforge.yaml ya existe" {
			t.Errorf("expected Spanish 'already exists' error, got %v", err)
		}
	})
}

func TestInitJSONLanguageIndependent(t *testing.T) {
	inTempDir(t, func(_ string) {
		outEn, err := runCLI(t, "init", "--json", "--lang", "en")
		if err != nil {
			t.Fatal(err)
		}
		removeInitArtifacts(t)

		outEs, err := runCLI(t, "init", "--json", "--lang", "es")
		if err != nil {
			t.Fatal(err)
		}
		if outEn != outEs {
			t.Errorf("JSON output must be byte-identical regardless of language:\nen: %q\nes: %q", outEn, outEs)
		}
	})
}

func removeInitArtifacts(t *testing.T) {
	t.Helper()
	for _, p := range []string{"goforge.yaml", ".env"} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	if err := os.RemoveAll("migrations"); err != nil {
		t.Fatal(err)
	}
}

func TestResetAndFreshRequireYesInNonInteractive(t *testing.T) {
	inTempDir(t, func(_ string) {
		_, err := runCLI(t, "init", "--lang", "es")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		// migrate:reset without --yes in non-TTY (bytes.Buffer)
		_, err = runCLI(t, "migrate:reset", "--lang", "es")
		if err == nil {
			t.Fatal("expected error without --yes")
		}
		if err.Error() != "migrate:reset revierte todas las migraciones aplicadas; vuelve a ejecutar con --yes para confirmar" {
			t.Errorf("unexpected error message: %v", err)
		}

		// migrate:fresh without --yes in non-TTY (bytes.Buffer)
		_, err = runCLI(t, "migrate:fresh", "--lang", "es")
		if err == nil {
			t.Fatal("expected error without --yes")
		}
		if err.Error() != "migrate:fresh revierte y vuelve a aplicar todas las migraciones; vuelve a ejecutar con --yes para confirmar" {
			t.Errorf("unexpected error message: %v", err)
		}

		// migrate without --yes in non-TTY (bytes.Buffer)
		_, err = runCLI(t, "migrate", "--lang", "es")
		if err == nil {
			t.Fatal("expected error without --yes")
		}
		if err.Error() != "se requiere el flag --yes para ejecutar en entornos no interactivos" {
			t.Errorf("unexpected error message for migrate: %v", err)
		}

		// migrate:rollback without --yes in non-TTY (bytes.Buffer)
		_, err = runCLI(t, "migrate:rollback", "--lang", "es")
		if err == nil {
			t.Fatal("expected error without --yes")
		}
		if err.Error() != "se requiere el flag --yes para ejecutar en entornos no interactivos" {
			t.Errorf("unexpected error message for rollback: %v", err)
		}

		// English counterpart
		_, err = runCLI(t, "migrate", "--lang", "en")
		if err == nil {
			t.Fatal("expected error without --yes in EN")
		}
		if err.Error() != "the --yes flag is required in non-interactive environments" {
			t.Errorf("unexpected error message for migrate in EN: %v", err)
		}
	})
}

