package migration

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestRegistryRejectsDuplicateVersion(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(&fakeMigration{version: 1, name: "a"}, "sum-a"); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := r.Register(&fakeMigration{version: 1, name: "a-again"}, "sum-b")
	if !errors.Is(err, ErrDuplicateVersion) {
		t.Fatalf("expected ErrDuplicateVersion, got %v", err)
	}
}

func TestChecksumDetectsChange(t *testing.T) {
	a := Checksum([]byte("CREATE TABLE users (id INT);"))
	b := Checksum([]byte("CREATE TABLE users (id BIGINT);"))
	if a == b {
		t.Fatal("expected different checksums for different content")
	}
	if a != Checksum([]byte("CREATE TABLE users (id INT);")) {
		t.Fatal("expected checksum to be deterministic")
	}
}

func TestLoadMergesSQLAndGoMigrationsRejectsOverlap(t *testing.T) {
	fsys := fstest.MapFS{
		"000001_create_users.up.sql":   {Data: []byte("CREATE TABLE users (id INT);")},
		"000001_create_users.down.sql": {Data: []byte("DROP TABLE users;")},
	}
	reg := NewRegistry()
	if err := reg.Register(&fakeMigration{version: 2, name: "seed"}, "sum"); err != nil {
		t.Fatalf("Register: %v", err)
	}
	entries, err := Load(fsys, reg)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 2 || entries[0].Version() != 1 || entries[1].Version() != 2 {
		t.Fatalf("unexpected merged entries: %v", entries)
	}

	regOverlap := NewRegistry()
	if err := regOverlap.Register(&fakeMigration{version: 1, name: "collides"}, "sum"); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := Load(fsys, regOverlap); !errors.Is(err, ErrDuplicateVersion) {
		t.Fatalf("expected ErrDuplicateVersion for SQL/Go version collision, got %v", err)
	}
}
