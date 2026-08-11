package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
)

// The following in-memory Provider/DB/Locker fakes let the engine's
// orchestration logic (planning, batching, locking, checksum/dirty checks,
// and transactional rollback) be unit tested without a real database.

type fakeDB struct {
	mu      sync.Mutex
	history *fakeHistory
	failSQL map[string]bool // query text -> simulate ExecContext failure
}

func (f *fakeDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failSQL[query] {
		return nil, fmt.Errorf("fakeDB: simulated failure for %q", query)
	}
	return driverResult{}, nil
}

func (f *fakeDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, fmt.Errorf("fakeDB: QueryContext not supported")
}

func (f *fakeDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (f *fakeDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	return &fakeTx{db: f, history: f.history, stagedSet: map[uint64]Record{}, stagedDelete: map[uint64]bool{}}, nil
}

type driverResult struct{}

func (driverResult) LastInsertId() (int64, error) { return 0, nil }
func (driverResult) RowsAffected() (int64, error) { return 1, nil }

// fakeTx applies statements directly against the underlying fakeDB (no real
// row-level isolation), but stages history writes so that Rollback discards
// them and Commit applies them atomically, exactly what real transactional
// providers guarantee.
type fakeTx struct {
	db      *fakeDB
	history *fakeHistory

	stagedSet    map[uint64]Record
	stagedDelete map[uint64]bool
	done         bool
}

func (t *fakeTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.db.ExecContext(ctx, query, args...)
}

func (t *fakeTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.db.QueryContext(ctx, query, args...)
}

func (t *fakeTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.db.QueryRowContext(ctx, query, args...)
}

func (t *fakeTx) Commit() error {
	t.done = true
	t.history.mu.Lock()
	defer t.history.mu.Unlock()
	for v, r := range t.stagedSet {
		t.history.records[v] = r
	}
	for v := range t.stagedDelete {
		delete(t.history.records, v)
	}
	return nil
}

func (t *fakeTx) Rollback() error {
	t.done = true
	return nil
}

// fakeHistory stores Records in memory. When called with a *fakeTx it
// stages the change instead of writing straight through, so a rolled-back
// transaction leaves no trace.
type fakeHistory struct {
	mu      sync.Mutex
	records map[uint64]Record
}

func newFakeHistory() *fakeHistory {
	return &fakeHistory{records: make(map[uint64]Record)}
}

func (h *fakeHistory) EnsureTable(ctx context.Context, db DB) error { return nil }

func (h *fakeHistory) List(ctx context.Context, db DB) ([]Record, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Record, 0, len(h.records))
	for _, r := range h.records {
		out = append(out, r)
	}
	return out, nil
}

func (h *fakeHistory) Begin(ctx context.Context, db DB, version uint64, name, checksum string, batch int) error {
	rec := Record{Version: version, Name: name, Checksum: checksum, Batch: batch, AppliedAt: time.Now(), Dirty: true}
	if tx, ok := db.(*fakeTx); ok {
		tx.stagedSet[version] = rec
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records[version] = rec
	return nil
}

func (h *fakeHistory) Complete(ctx context.Context, db DB, version uint64, executionTime time.Duration) error {
	if tx, ok := db.(*fakeTx); ok {
		r, ok := tx.stagedSet[version]
		if !ok {
			return fmt.Errorf("fakeHistory: no staged record for version %d", version)
		}
		r.Dirty = false
		r.ExecutionTime = executionTime
		tx.stagedSet[version] = r
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.records[version]
	if !ok {
		return fmt.Errorf("fakeHistory: no record for version %d", version)
	}
	r.Dirty = false
	r.ExecutionTime = executionTime
	h.records[version] = r
	return nil
}

func (h *fakeHistory) Remove(ctx context.Context, db DB, version uint64) error {
	if tx, ok := db.(*fakeTx); ok {
		tx.stagedDelete[version] = true
		delete(tx.stagedSet, version)
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.records, version)
	return nil
}

// fakeLocker simulates a session-scoped lock: only one holder at a time,
// matching pg_advisory_lock / GET_LOCK semantics closely enough for testing
// that the engine serializes migrate runs.
type fakeLocker struct {
	mu     *sync.Mutex
	locked *bool
}

func (l *fakeLocker) Lock(ctx context.Context) error {
	if !l.mu.TryLock() {
		return errors.New("fakeLocker: already locked")
	}
	*l.locked = true
	return nil
}

func (l *fakeLocker) Unlock(ctx context.Context) error {
	*l.locked = false
	l.mu.Unlock()
	return nil
}

// fakeProvider wires the fakes above behind the Provider interface.
// transactional=false exercises the "left dirty on failure" path, mirroring
// how MariaDB DDL cannot be wrapped in a rollback-able transaction.
type fakeProvider struct {
	transactional bool
	history       *fakeHistory
	lockMu        sync.Mutex
	locked        bool
}

func newFakeProvider(transactional bool) (*fakeProvider, *fakeDB) {
	history := newFakeHistory()
	p := &fakeProvider{transactional: transactional, history: history}
	db := &fakeDB{history: history, failSQL: map[string]bool{}}
	return p, db
}

func (p *fakeProvider) Name() string                   { return "fake" }
func (p *fakeProvider) SupportsTransactionalDDL() bool { return p.transactional }
func (p *fakeProvider) History() HistoryStore          { return p.history }
func (p *fakeProvider) Locker(db DB) Locker {
	return &fakeLocker{mu: &p.lockMu, locked: &p.locked}
}

// fakeMigration is a Migration whose Up/Down behavior is controlled by the
// test via failUp/failDown.
type fakeMigration struct {
	version          uint64
	name             string
	failUp, failDown bool
}

func (m *fakeMigration) Version() uint64 { return m.version }
func (m *fakeMigration) Name() string    { return m.name }

func (m *fakeMigration) Up(ctx context.Context, db DB) error {
	if m.failUp {
		return fmt.Errorf("simulated failure applying %d", m.version)
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf("-- up %d", m.version))
	return err
}

func (m *fakeMigration) Down(ctx context.Context, db DB) error {
	if m.failDown {
		return fmt.Errorf("simulated failure reverting %d", m.version)
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf("-- down %d", m.version))
	return err
}

// recordingDB records each ExecContext query for tests that only care about
// what SQL was run, not full history/locking behavior.
type recordingDB struct {
	exec func(query string)
}

func (r *recordingDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if r.exec != nil {
		r.exec(query)
	}
	return driverResult{}, nil
}

func (r *recordingDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, fmt.Errorf("recordingDB: QueryContext not supported")
}

func (r *recordingDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func entryFor(m *fakeMigration) Entry {
	return Entry{Migration: m, Checksum: Checksum([]byte(fmt.Sprintf("%d:%s", m.version, m.name)))}
}
