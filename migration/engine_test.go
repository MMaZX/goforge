package migration

import (
	"context"
	"errors"
	"testing"
)

func newEngine(t *testing.T, transactional bool, migrations ...*fakeMigration) (*Engine, *fakeProvider, *fakeDB) {
	t.Helper()
	provider, db := newFakeProvider(transactional)
	entries := make([]Entry, 0, len(migrations))
	for _, m := range migrations {
		entries = append(entries, entryFor(m))
	}
	e, err := NewEngine(db, provider, entries)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return e, provider, db
}

func TestUpAppliesPendingInOrder(t *testing.T) {
	e, _, _ := newEngine(t, true,
		&fakeMigration{version: 2, name: "b"},
		&fakeMigration{version: 1, name: "a"},
	)
	result, err := e.Up(context.Background(), 0)
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if len(result.Executed) != 2 {
		t.Fatalf("expected 2 executed, got %d", len(result.Executed))
	}
	if result.Executed[0].Version != 1 || result.Executed[1].Version != 2 {
		t.Fatalf("expected ascending order 1,2 got %v", result.Executed)
	}
	if result.Batch != 1 {
		t.Fatalf("expected batch 1, got %d", result.Batch)
	}
}

func TestUpNoPendingReturnsErrNoMigrations(t *testing.T) {
	e, _, _ := newEngine(t, true, &fakeMigration{version: 1, name: "a"})
	if _, err := e.Up(context.Background(), 0); err != nil {
		t.Fatalf("first Up: %v", err)
	}
	if _, err := e.Up(context.Background(), 0); !errors.Is(err, ErrNoMigrations) {
		t.Fatalf("expected ErrNoMigrations, got %v", err)
	}
}

func TestDownRollsBackLastBatch(t *testing.T) {
	e, _, _ := newEngine(t, true, &fakeMigration{version: 1, name: "a"})
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err != nil {
		t.Fatalf("Up: %v", err)
	}
	// second batch
	e2, _, db2 := newEngine(t, true, &fakeMigration{version: 1, name: "a"}, &fakeMigration{version: 2, name: "b"})
	_ = db2
	if _, err := e2.Up(ctx, 1); err != nil { // batch 1: version 1
		t.Fatalf("Up steps 1: %v", err)
	}
	if _, err := e2.Up(ctx, 0); err != nil { // batch 2: version 2
		t.Fatalf("Up rest: %v", err)
	}
	result, err := e2.Rollback(ctx, 0)
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if len(result.Executed) != 1 || result.Executed[0].Version != 2 {
		t.Fatalf("expected rollback of batch 2 (version 2) only, got %v", result.Executed)
	}

	status, err := e2.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	for _, se := range status {
		if se.Version == 1 && !se.Applied {
			t.Fatalf("version 1 should remain applied after rolling back batch 2")
		}
		if se.Version == 2 && se.Applied {
			t.Fatalf("version 2 should have been rolled back")
		}
	}
}

func TestResetRollsBackEverything(t *testing.T) {
	e, _, _ := newEngine(t, true, &fakeMigration{version: 1, name: "a"}, &fakeMigration{version: 2, name: "b"})
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if _, err := e.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	status, err := e.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	for _, se := range status {
		if se.Applied {
			t.Fatalf("expected nothing applied after Reset, got %+v", se)
		}
	}
}

func TestFreshResetsAndReapplies(t *testing.T) {
	e, _, _ := newEngine(t, true, &fakeMigration{version: 1, name: "a"})
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err != nil {
		t.Fatalf("Up: %v", err)
	}
	result, err := e.Fresh(ctx)
	if err != nil {
		t.Fatalf("Fresh: %v", err)
	}
	if len(result.Executed) != 1 || result.Executed[0].Version != 1 {
		t.Fatalf("expected version 1 reapplied, got %v", result.Executed)
	}
}

func TestFailedMigrationTransactionalRollsBackHistory(t *testing.T) {
	e, provider, _ := newEngine(t, true, &fakeMigration{version: 1, name: "a", failUp: true})
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err == nil {
		t.Fatal("expected error from failing migration")
	}
	records, _ := provider.History().List(ctx, nil)
	if len(records) != 0 {
		t.Fatalf("expected no history left after transactional rollback, got %v", records)
	}
}

func TestFailedMigrationNonTransactionalLeavesDirty(t *testing.T) {
	e, provider, _ := newEngine(t, false, &fakeMigration{version: 1, name: "a", failUp: true})
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err == nil {
		t.Fatal("expected error from failing migration")
	}
	records, _ := provider.History().List(ctx, nil)
	if len(records) != 1 || !records[0].Dirty {
		t.Fatalf("expected one dirty record, got %v", records)
	}

	// Any subsequent operation must refuse to proceed until the dirty state
	// is resolved manually.
	_, err := e.Up(ctx, 0)
	if !errors.Is(err, ErrDirtyState) {
		t.Fatalf("expected ErrDirtyState, got %v", err)
	}
}

func TestChecksumMismatchStopsExecution(t *testing.T) {
	m := &fakeMigration{version: 1, name: "a"}
	e, provider, db := newEngine(t, true, m)
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err != nil {
		t.Fatalf("Up: %v", err)
	}

	// Simulate the migration file being modified after it was applied: a
	// new Engine is built with a different checksum for the same version.
	tamperedEntry := Entry{Migration: m, Checksum: "tampered-checksum"}
	e2, err := NewEngine(db, provider, []Entry{tamperedEntry})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	if _, err := e2.Up(ctx, 0); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected ErrChecksumMismatch, got %v", err)
	}
	if err := e2.Validate(ctx); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected Validate to report ErrChecksumMismatch, got %v", err)
	}
}

func TestAlreadyAppliedIsSkippedNotReapplied(t *testing.T) {
	m := &fakeMigration{version: 1, name: "a"}
	e, _, _ := newEngine(t, true, m)
	ctx := context.Background()
	if _, err := e.Up(ctx, 0); err != nil {
		t.Fatalf("first Up: %v", err)
	}
	plan, err := e.Plan(ctx, Up, 0)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plan) != 0 {
		t.Fatalf("expected nothing pending, got %v", plan)
	}
}

func TestPendingMigrationsPlan(t *testing.T) {
	e, _, _ := newEngine(t, true,
		&fakeMigration{version: 1, name: "a"},
		&fakeMigration{version: 2, name: "b"},
	)
	ctx := context.Background()
	if _, err := e.Up(ctx, 1); err != nil {
		t.Fatalf("Up steps=1: %v", err)
	}
	plan, err := e.Plan(ctx, Up, 0)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plan) != 1 || plan[0].Version() != 2 {
		t.Fatalf("expected only version 2 pending, got %v", plan)
	}
}

func TestOutOfOrderMigrationIsStillApplied(t *testing.T) {
	// version 1 arrives later than version 2 was applied - a common
	// scenario when merging branches. GoForge applies it rather than
	// silently skipping it, since Version() ordering is the source of truth
	// and 1 is still pending.
	ctx := context.Background()
	e2, provider, db := newEngine(t, true, &fakeMigration{version: 2, name: "b"})
	if _, err := e2.Up(ctx, 0); err != nil {
		t.Fatalf("Up on e2: %v", err)
	}
	entries := []Entry{
		entryFor(&fakeMigration{version: 1, name: "a"}),
		entryFor(&fakeMigration{version: 2, name: "b"}),
	}
	eCombined, err := NewEngine(db, provider, entries)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	plan, err := eCombined.Plan(ctx, Up, 0)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plan) != 1 || plan[0].Version() != 1 {
		t.Fatalf("expected version 1 pending, got %v", plan)
	}
}

func TestInvalidMigrationDuplicateVersionRejected(t *testing.T) {
	provider, db := newFakeProvider(true)
	entries := []Entry{
		entryFor(&fakeMigration{version: 1, name: "a"}),
		entryFor(&fakeMigration{version: 1, name: "a-dup"}),
	}
	if _, err := NewEngine(db, provider, entries); !errors.Is(err, ErrDuplicateVersion) {
		t.Fatalf("expected ErrDuplicateVersion, got %v", err)
	}
}

func TestConcurrentMigrateIsSerializedByLock(t *testing.T) {
	provider, db := newFakeProvider(true)
	entries := []Entry{entryFor(&fakeMigration{version: 1, name: "a"})}
	e, err := NewEngine(db, provider, entries)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	locker := provider.Locker(db)
	if err := locker.Lock(context.Background()); err != nil {
		t.Fatalf("acquiring lock for test: %v", err)
	}
	defer locker.Unlock(context.Background())

	if _, err := e.Up(context.Background(), 0); !errors.Is(err, ErrLocked) {
		t.Fatalf("expected ErrLocked while another holder has the lock, got %v", err)
	}
}

func TestRequiresTxCapableDBWhenTransactional(t *testing.T) {
	provider, _ := newFakeProvider(true)
	nonTxDB := struct{ DB }{&fakeDB{history: newFakeHistory(), failSQL: map[string]bool{}}}
	_, err := NewEngine(nonTxDB, provider, nil)
	if err == nil {
		t.Fatal("expected error constructing Engine with non-transactional DB for a transactional provider")
	}
}
