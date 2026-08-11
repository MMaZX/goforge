package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MMaZX/goforge/migration"
)

func TestGenerateProducesValidRegistry(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"000001_create_users.go": "package migrations\n\ntype migration000001 struct{}\n",
		"000002_add_index.go":    "package migrations\n\ntype migration000002 struct{}\n",
	}
	var wantChecksums = map[uint64]string{}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("writing fixture %s: %v", name, err)
		}
	}
	wantChecksums[1] = migration.Checksum([]byte(files["000001_create_users.go"]))
	wantChecksums[2] = migration.Checksum([]byte(files["000002_add_index.go"]))

	if err := Generate(dir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	generated, err := os.ReadFile(filepath.Join(dir, OutputFileName))
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	if _, err := format.Source(generated); err != nil {
		t.Fatalf("generated file is not valid Go source: %v", err)
	}

	src := string(generated)
	if !strings.Contains(src, `r.Register(migration000001{}, "`+wantChecksums[1]+`")`) {
		t.Fatalf("expected registration for migration000001 with checksum %s, got:\n%s", wantChecksums[1], src)
	}
	if !strings.Contains(src, `r.Register(migration000002{}, "`+wantChecksums[2]+`")`) {
		t.Fatalf("expected registration for migration000002 with checksum %s, got:\n%s", wantChecksums[2], src)
	}
}

func TestGenerateRejectsDuplicateVersion(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "000001_a.go", "package migrations\n\ntype migration000001 struct{}\n")
	mustWrite(t, dir, "000001_b.go", "package migrations\n\ntype migration000001 struct{}\n")

	err := Generate(dir)
	if err == nil {
		t.Fatal("expected error for duplicate version")
	}
}

func TestGenerateSkipsItsOwnOutputFile(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "000001_a.go", "package migrations\n\ntype migration000001 struct{}\n")
	if err := Generate(dir); err != nil {
		t.Fatalf("first Generate: %v", err)
	}
	// Regenerating must not treat the previous output as a migration file.
	if err := Generate(dir); err != nil {
		t.Fatalf("second Generate: %v", err)
	}
}

func mustWrite(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}
