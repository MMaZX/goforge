//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/MMaZX/goforge/migration"
)

// testMigration is a Migration backed by real SQL, used to exercise the
// engine against an actual database instead of the in-process fakes used by
// the migration package's unit tests.
type testMigration struct {
	version        uint64
	name           string
	upSQL, downSQL string
	failUp         bool
}

func (m testMigration) Version() uint64 { return m.version }
func (m testMigration) Name() string    { return m.name }

func (m testMigration) Up(ctx context.Context, db migration.DB) error {
	if m.failUp {
		return fmt.Errorf("simulated failure applying %d", m.version)
	}
	_, err := db.ExecContext(ctx, m.upSQL)
	return err
}

func (m testMigration) Down(ctx context.Context, db migration.DB) error {
	_, err := db.ExecContext(ctx, m.downSQL)
	return err
}

func entry(m testMigration) migration.Entry {
	return migration.Entry{Migration: m, Checksum: migration.Checksum([]byte(m.upSQL + "|" + m.downSQL))}
}

// resetHistory drops the history table so each subtest starts from a clean
// slate on the shared container.
func resetHistory(ctx context.Context, t *testing.T, db migration.DB) {
	t.Helper()
	if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS goforge_migrations"); err != nil {
		t.Fatalf("resetting history table: %v", err)
	}
}

func runSuite(t *testing.T, h harness) {
	ctx := context.Background()
	dsn, terminate, err := h.start(ctx)
	if err != nil {
		t.Fatalf("%s: starting container: %v", h.name, err)
	}
	t.Cleanup(terminate)

	newEngine := func(t *testing.T, entries ...migration.Entry) (*migration.Engine, provider, migration.DB) {
		t.Helper()
		p, err := h.newProvider(ctx, dsn)
		if err != nil {
			t.Fatalf("%s: connecting provider: %v", h.name, err)
		}
		t.Cleanup(func() { p.Close() })
		db := p.DB()
		resetHistory(ctx, t, db)
		e, err := migration.NewEngine(db, p, entries)
		if err != nil {
			t.Fatalf("%s: NewEngine: %v", h.name, err)
		}
		return e, p, db
	}

	t.Run(h.name+"/MigrationUp", func(t *testing.T) {
		e, _, db := newEngine(t, entry(testMigration{
			version: 1, name: "create_widgets_up",
			upSQL:   "CREATE TABLE widgets_up (id INT PRIMARY KEY)",
			downSQL: "DROP TABLE widgets_up",
		}))
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_up")

		result, err := e.Up(ctx, 0)
		if err != nil {
			t.Fatalf("Up: %v", err)
		}
		if len(result.Executed) != 1 {
			t.Fatalf("expected 1 executed, got %d", len(result.Executed))
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO widgets_up (id) VALUES (1)"); err != nil {
			t.Fatalf("table from migration not usable: %v", err)
		}
	})

	t.Run(h.name+"/MigrationDown", func(t *testing.T) {
		e, _, db := newEngine(t, entry(testMigration{
			version: 1, name: "create_widgets_down",
			upSQL:   "CREATE TABLE widgets_down (id INT PRIMARY KEY)",
			downSQL: "DROP TABLE widgets_down",
		}))
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_down")

		if _, err := e.Up(ctx, 0); err != nil {
			t.Fatalf("Up: %v", err)
		}
		if _, err := e.Reset(ctx); err != nil {
			t.Fatalf("Reset (down): %v", err)
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO widgets_down (id) VALUES (1)"); err == nil {
			t.Fatal("expected widgets_down to no longer exist after Down")
		}
	})

	t.Run(h.name+"/Rollback", func(t *testing.T) {
		e, _, db := newEngine(t,
			entry(testMigration{version: 1, name: "a", upSQL: "CREATE TABLE widgets_rb_a (id INT)", downSQL: "DROP TABLE widgets_rb_a"}),
			entry(testMigration{version: 2, name: "b", upSQL: "CREATE TABLE widgets_rb_b (id INT)", downSQL: "DROP TABLE widgets_rb_b"}),
		)
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_rb_a")
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_rb_b")

		if _, err := e.Up(ctx, 1); err != nil {
			t.Fatalf("Up batch 1: %v", err)
		}
		if _, err := e.Up(ctx, 0); err != nil {
			t.Fatalf("Up batch 2: %v", err)
		}
		result, err := e.Rollback(ctx, 0)
		if err != nil {
			t.Fatalf("Rollback: %v", err)
		}
		if len(result.Executed) != 1 || result.Executed[0].Version != 2 {
			t.Fatalf("expected rollback of version 2 only, got %v", result.Executed)
		}
		if _, err := db.ExecContext(ctx, "SELECT 1 FROM widgets_rb_a LIMIT 1"); err != nil {
			t.Fatalf("widgets_rb_a should still exist: %v", err)
		}
	})

	t.Run(h.name+"/FailedMigration", func(t *testing.T) {
		e, _, _ := newEngine(t, entry(testMigration{version: 1, name: "boom", failUp: true, downSQL: "SELECT 1"}))
		if _, err := e.Up(ctx, 0); err == nil {
			t.Fatal("expected error from failing migration")
		}
	})

	t.Run(h.name+"/TransactionRollback", func(t *testing.T) {
		if !h.transactional {
			t.Skip("provider does not support transactional DDL")
		}
		e, p, db := newEngine(t, entry(testMigration{
			version: 1, name: "half",
			upSQL:   "CREATE TABLE widgets_txr (id INT); SELECT 1/0", // second statement fails
			downSQL: "DROP TABLE widgets_txr",
		}))
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_txr")

		if _, err := e.Up(ctx, 0); err == nil {
			t.Fatal("expected error from failing statement")
		}
		records, err := p.History().List(ctx, db)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(records) != 0 {
			t.Fatalf("expected history rolled back, got %v", records)
		}
		if _, err := db.ExecContext(ctx, "SELECT 1 FROM widgets_txr LIMIT 1"); err == nil {
			t.Fatal("expected widgets_txr to not exist, DDL should have rolled back with the transaction")
		}
	})

	t.Run(h.name+"/ChecksumMismatch", func(t *testing.T) {
		m := testMigration{version: 1, name: "a", upSQL: "CREATE TABLE widgets_cksum (id INT)", downSQL: "DROP TABLE widgets_cksum"}
		e, p, db := newEngine(t, entry(m))
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_cksum")

		if _, err := e.Up(ctx, 0); err != nil {
			t.Fatalf("Up: %v", err)
		}
		tampered := migration.Entry{Migration: m, Checksum: "tampered"}
		e2, err := migration.NewEngine(db, p, []migration.Entry{tampered})
		if err != nil {
			t.Fatalf("NewEngine: %v", err)
		}
		if _, err := e2.Up(ctx, 0); !errors.Is(err, migration.ErrChecksumMismatch) {
			t.Fatalf("expected ErrChecksumMismatch, got %v", err)
		}
	})

	t.Run(h.name+"/AlreadyApplied", func(t *testing.T) {
		e, _, _ := newEngine(t, entry(testMigration{version: 1, name: "a", upSQL: "CREATE TABLE widgets_dup (id INT)", downSQL: "DROP TABLE widgets_dup"}))
		if _, err := e.Up(context.Background(), 0); err != nil {
			t.Fatalf("Up: %v", err)
		}
		if _, err := e.Up(ctx, 0); !errors.Is(err, migration.ErrNoMigrations) {
			t.Fatalf("expected ErrNoMigrations on second Up, got %v", err)
		}
	})

	t.Run(h.name+"/PendingMigrations", func(t *testing.T) {
		e, _, _ := newEngine(t,
			entry(testMigration{version: 1, name: "a", upSQL: "CREATE TABLE widgets_pend_a (id INT)", downSQL: "DROP TABLE widgets_pend_a"}),
			entry(testMigration{version: 2, name: "b", upSQL: "CREATE TABLE widgets_pend_b (id INT)", downSQL: "DROP TABLE widgets_pend_b"}),
		)
		if _, err := e.Up(ctx, 1); err != nil {
			t.Fatalf("Up: %v", err)
		}
		plan, err := e.Plan(ctx, migration.Up, 0)
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}
		if len(plan) != 1 || plan[0].Version() != 2 {
			t.Fatalf("expected version 2 pending, got %v", plan)
		}
	})

	t.Run(h.name+"/ConcurrentMigrate", func(t *testing.T) {
		e1, p, db := newEngine(t, entry(testMigration{version: 1, name: "a", upSQL: "CREATE TABLE widgets_conc (id INT)", downSQL: "DROP TABLE widgets_conc"}))
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_conc")

		locker := p.Locker(db)
		if err := locker.Lock(ctx); err != nil {
			t.Fatalf("acquiring lock for test: %v", err)
		}
		defer locker.Unlock(ctx)

		lockCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if _, err := e1.Up(lockCtx, 0); !errors.Is(err, migration.ErrLocked) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected lock contention error, got %v", err)
		}
	})

	t.Run(h.name+"/DatabaseConnectionFailure", func(t *testing.T) {
		_, err := h.newProvider(ctx, "invalid-dsn-that-does-not-resolve")
		if err == nil {
			t.Fatal("expected connection failure for invalid DSN")
		}
	})

	t.Run(h.name+"/DirtyMigrationState", func(t *testing.T) {
		if h.transactional {
			t.Skip("transactional providers never leave a dirty row behind")
		}
		e, _, db := newEngine(t, entry(testMigration{version: 1, name: "boom", failUp: true, downSQL: "SELECT 1"}))
		if _, err := e.Up(ctx, 0); err == nil {
			t.Fatal("expected error from failing migration")
		}
		if _, err := e.Up(ctx, 0); !errors.Is(err, migration.ErrDirtyState) {
			t.Fatalf("expected ErrDirtyState, got %v", err)
		}
		db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_boom")
	})

	t.Run(h.name+"/OutOfOrder", func(t *testing.T) {
		e, p, db := newEngine(t, entry(testMigration{version: 2, name: "b", upSQL: "CREATE TABLE widgets_ooo_b (id INT)", downSQL: "DROP TABLE widgets_ooo_b"}))
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_ooo_a")
		defer db.ExecContext(ctx, "DROP TABLE IF EXISTS widgets_ooo_b")

		if _, err := e.Up(ctx, 0); err != nil {
			t.Fatalf("Up: %v", err)
		}
		entries := []migration.Entry{
			entry(testMigration{version: 1, name: "a", upSQL: "CREATE TABLE widgets_ooo_a (id INT)", downSQL: "DROP TABLE widgets_ooo_a"}),
			entry(testMigration{version: 2, name: "b", upSQL: "CREATE TABLE widgets_ooo_b (id INT)", downSQL: "DROP TABLE widgets_ooo_b"}),
		}
		e2, err := migration.NewEngine(db, p, entries)
		if err != nil {
			t.Fatalf("NewEngine: %v", err)
		}
		result, err := e2.Up(ctx, 0)
		if err != nil {
			t.Fatalf("Up: %v", err)
		}
		if len(result.Executed) != 1 || result.Executed[0].Version != 1 {
			t.Fatalf("expected version 1 applied out of order, got %v", result.Executed)
		}
	})

	t.Run(h.name+"/InvalidMigrationDuplicateVersion", func(t *testing.T) {
		p, err := h.newProvider(ctx, dsn)
		if err != nil {
			t.Fatalf("connecting provider: %v", err)
		}
		defer p.Close()
		entries := []migration.Entry{
			entry(testMigration{version: 1, name: "a", upSQL: "SELECT 1", downSQL: "SELECT 1"}),
			entry(testMigration{version: 1, name: "a-dup", upSQL: "SELECT 1", downSQL: "SELECT 1"}),
		}
		if _, err := migration.NewEngine(p.DB(), p, entries); !errors.Is(err, migration.ErrDuplicateVersion) {
			t.Fatalf("expected ErrDuplicateVersion, got %v", err)
		}
	})
}
