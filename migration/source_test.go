package migration

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestLoadSQLMigrationsPairsUpAndDown(t *testing.T) {
	fsys := fstest.MapFS{
		"000001_create_users.up.sql":   {Data: []byte("CREATE TABLE users (id INT);")},
		"000001_create_users.down.sql": {Data: []byte("DROP TABLE users;")},
		"000002_create_posts.up.sql":   {Data: []byte("CREATE TABLE posts (id INT);")},
		"000002_create_posts.down.sql": {Data: []byte("DROP TABLE posts;")},
	}
	entries, err := LoadSQLMigrations(fsys)
	if err != nil {
		t.Fatalf("LoadSQLMigrations: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Version() != 1 || entries[1].Version() != 2 {
		t.Fatalf("expected ascending order, got %v", entries)
	}
	if entries[0].Checksum == "" {
		t.Fatal("expected non-empty checksum")
	}
}

func TestLoadSQLMigrationsMissingDownFails(t *testing.T) {
	fsys := fstest.MapFS{
		"000001_create_users.up.sql": {Data: []byte("CREATE TABLE users (id INT);")},
	}
	if _, err := LoadSQLMigrations(fsys); err == nil {
		t.Fatal("expected error for missing .down.sql")
	}
}

func TestLoadSQLMigrationsExecutesStatementsInOrder(t *testing.T) {
	fsys := fstest.MapFS{
		"000001_x.up.sql":   {Data: []byte("CREATE TABLE a (id INT);\nCREATE TABLE b (id INT);")},
		"000001_x.down.sql": {Data: []byte("DROP TABLE b;\nDROP TABLE a;")},
	}
	entries, err := LoadSQLMigrations(fsys)
	if err != nil {
		t.Fatalf("LoadSQLMigrations: %v", err)
	}
	var executed []string
	db := &recordingDB{exec: func(query string) { executed = append(executed, query) }}
	if err := entries[0].Up(context.Background(), db); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if len(executed) != 2 || executed[0] != "CREATE TABLE a (id INT)" || executed[1] != "CREATE TABLE b (id INT)" {
		t.Fatalf("unexpected statements executed: %v", executed)
	}
}
