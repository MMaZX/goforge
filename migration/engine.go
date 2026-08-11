package migration

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// Engine runs migrations against a database using a Provider for locking,
// transactional behavior and history storage. The CLI and the MCP server
// both operate exclusively through this type, so they share identical
// guarantees.
type Engine struct {
	db        DB
	provider  Provider
	entries   []Entry
	byVersion map[uint64]Entry
}

// NewEngine builds an Engine from a sorted, de-duplicated migration set.
func NewEngine(db DB, provider Provider, entries []Entry) (*Engine, error) {
	if db == nil {
		return nil, fmt.Errorf("migration: db is required")
	}
	if provider == nil {
		return nil, fmt.Errorf("migration: provider is required")
	}
	if provider.SupportsTransactionalDDL() {
		if _, ok := db.(TxCapableDB); !ok {
			return nil, fmt.Errorf("migration: provider %q requires a TxCapableDB", provider.Name())
		}
	}

	byVersion := make(map[uint64]Entry, len(entries))
	for _, e := range entries {
		if _, dup := byVersion[e.Version()]; dup {
			return nil, fmt.Errorf("%w: %d", ErrDuplicateVersion, e.Version())
		}
		byVersion[e.Version()] = e
	}
	sorted := append([]Entry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Version() < sorted[j].Version() })

	return &Engine{db: db, provider: provider, entries: sorted, byVersion: byVersion}, nil
}

// StatusEntry describes one migration's applied/pending state.
type StatusEntry struct {
	Version   uint64
	Name      string
	Applied   bool
	AppliedAt time.Time
	Batch     int
	Dirty     bool
}

// Result reports what an Up/Rollback/Reset/Fresh call executed.
type Result struct {
	Executed []Record
	Batch    int
}

// Status lists every known migration alongside its applied state.
func (e *Engine) Status(ctx context.Context) ([]StatusEntry, error) {
	records, err := e.loadHistory(ctx)
	if err != nil {
		return nil, err
	}
	byVersion := make(map[uint64]Record, len(records))
	for _, r := range records {
		byVersion[r.Version] = r
	}

	out := make([]StatusEntry, 0, len(e.entries))
	for _, entry := range e.entries {
		se := StatusEntry{Version: entry.Version(), Name: entry.Name()}
		if r, ok := byVersion[entry.Version()]; ok {
			se.Applied = true
			se.AppliedAt = r.AppliedAt
			se.Batch = r.Batch
			se.Dirty = r.Dirty
		}
		out = append(out, se)
	}
	return out, nil
}

// Plan reports which migrations Up or Rollback would execute, without
// modifying the database.
func (e *Engine) Plan(ctx context.Context, dir Direction, steps int) ([]Entry, error) {
	records, err := e.loadHistoryChecked(ctx)
	if err != nil {
		return nil, err
	}
	if dir == Down {
		return e.computeDownPlan(records, steps), nil
	}
	return e.computeUpPlan(records, steps), nil
}

// Validate verifies checksums of applied migrations and reports dirty state,
// without touching the database beyond reading the history table.
func (e *Engine) Validate(ctx context.Context) error {
	records, err := e.loadHistoryChecked(ctx)
	if err != nil {
		return err
	}
	for _, r := range records {
		if _, ok := e.byVersion[r.Version]; !ok {
			return fmt.Errorf("migration: applied version %d (%s) has no corresponding migration file", r.Version, r.Name)
		}
	}
	return nil
}

// Up applies pending migrations in ascending version order. steps <= 0
// applies all of them.
func (e *Engine) Up(ctx context.Context, steps int) (Result, error) {
	var result Result
	err := e.withLock(ctx, func(ctx context.Context) error {
		records, err := e.loadHistoryChecked(ctx)
		if err != nil {
			return err
		}
		pending := e.computeUpPlan(records, steps)
		if len(pending) == 0 {
			return ErrNoMigrations
		}
		batch := maxBatch(records) + 1

		for _, entry := range pending {
			rec, err := e.apply(ctx, entry, batch)
			if err != nil {
				return err
			}
			result.Executed = append(result.Executed, rec)
		}
		result.Batch = batch
		return nil
	})
	return result, err
}

// Rollback reverts the most recently applied batch, or the last N applied
// migrations overall when steps > 0.
func (e *Engine) Rollback(ctx context.Context, steps int) (Result, error) {
	return e.rollbackWith(ctx, func(records []Record) []Entry {
		return e.computeDownPlan(records, steps)
	})
}

// Reset reverts every applied migration, newest first.
func (e *Engine) Reset(ctx context.Context) (Result, error) {
	return e.rollbackWith(ctx, e.computeAllAppliedDescending)
}

// Fresh reverts every applied migration and then re-applies all of them.
func (e *Engine) Fresh(ctx context.Context) (Result, error) {
	if _, err := e.Reset(ctx); err != nil {
		return Result{}, err
	}
	return e.Up(ctx, 0)
}

func (e *Engine) rollbackWith(ctx context.Context, plan func([]Record) []Entry) (Result, error) {
	var result Result
	err := e.withLock(ctx, func(ctx context.Context) error {
		records, err := e.loadHistoryChecked(ctx)
		if err != nil {
			return err
		}
		entries := plan(records)
		if len(entries) == 0 {
			return ErrNoMigrations
		}
		for _, entry := range entries {
			if err := e.revert(ctx, entry); err != nil {
				return err
			}
			result.Executed = append(result.Executed, Record{Version: entry.Version(), Name: entry.Name()})
		}
		return nil
	})
	return result, err
}

func (e *Engine) withLock(ctx context.Context, fn func(ctx context.Context) error) error {
	locker := e.provider.Locker(e.db)
	if err := locker.Lock(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrLocked, err)
	}
	defer locker.Unlock(ctx)
	return fn(ctx)
}

func (e *Engine) loadHistory(ctx context.Context) ([]Record, error) {
	if err := e.provider.History().EnsureTable(ctx, e.db); err != nil {
		return nil, fmt.Errorf("migration: ensuring history table: %w", err)
	}
	records, err := e.provider.History().List(ctx, e.db)
	if err != nil {
		return nil, fmt.Errorf("migration: listing history: %w", err)
	}
	return records, nil
}

// loadHistoryChecked loads history and enforces the invariants that must
// hold before any plan is computed or executed: no dirty migration left
// over from a previous failed run, and no applied migration modified since
// it ran.
func (e *Engine) loadHistoryChecked(ctx context.Context) ([]Record, error) {
	records, err := e.loadHistory(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		if r.Dirty {
			return nil, fmt.Errorf("%w: version %d (%s)", ErrDirtyState, r.Version, r.Name)
		}
	}
	for _, r := range records {
		entry, ok := e.byVersion[r.Version]
		if !ok {
			continue
		}
		if entry.Checksum != r.Checksum {
			return nil, fmt.Errorf("%w: version %d (%s)", ErrChecksumMismatch, r.Version, r.Name)
		}
	}
	return records, nil
}

func (e *Engine) computeUpPlan(records []Record, steps int) []Entry {
	applied := make(map[uint64]struct{}, len(records))
	for _, r := range records {
		applied[r.Version] = struct{}{}
	}
	var pending []Entry
	for _, entry := range e.entries {
		if _, ok := applied[entry.Version()]; !ok {
			pending = append(pending, entry)
		}
	}
	if steps > 0 && steps < len(pending) {
		pending = pending[:steps]
	}
	return pending
}

func (e *Engine) computeDownPlan(records []Record, steps int) []Entry {
	last := maxBatch(records)
	var candidates []Record
	for _, r := range records {
		if r.Batch == last {
			candidates = append(candidates, r)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Version > candidates[j].Version })
	if steps > 0 && steps < len(candidates) {
		candidates = candidates[:steps]
	}
	return e.entriesFor(candidates)
}

func (e *Engine) computeAllAppliedDescending(records []Record) []Entry {
	sorted := append([]Record(nil), records...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Version > sorted[j].Version })
	return e.entriesFor(sorted)
}

func (e *Engine) entriesFor(records []Record) []Entry {
	out := make([]Entry, 0, len(records))
	for _, r := range records {
		if entry, ok := e.byVersion[r.Version]; ok {
			out = append(out, entry)
		}
	}
	return out
}

func maxBatch(records []Record) int {
	max := 0
	for _, r := range records {
		if r.Batch > max {
			max = r.Batch
		}
	}
	return max
}

func (e *Engine) apply(ctx context.Context, entry Entry, batch int) (Record, error) {
	start := time.Now()
	hist := e.provider.History()

	if e.provider.SupportsTransactionalDDL() {
		txdb := e.db.(TxCapableDB)
		tx, err := txdb.BeginTx(ctx, nil)
		if err != nil {
			return Record{}, fmt.Errorf("migration: beginning transaction for %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		if err := hist.Begin(ctx, tx, entry.Version(), entry.Name(), entry.Checksum, batch); err != nil {
			_ = tx.Rollback()
			return Record{}, fmt.Errorf("migration: recording start of %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		if err := entry.Up(ctx, tx); err != nil {
			_ = tx.Rollback()
			return Record{}, fmt.Errorf("migration: applying %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		duration := time.Since(start)
		if err := hist.Complete(ctx, tx, entry.Version(), duration); err != nil {
			_ = tx.Rollback()
			return Record{}, fmt.Errorf("migration: recording completion of %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return Record{}, fmt.Errorf("migration: committing %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		return Record{Version: entry.Version(), Name: entry.Name(), Checksum: entry.Checksum, AppliedAt: start, ExecutionTime: duration, Batch: batch}, nil
	}

	if err := hist.Begin(ctx, e.db, entry.Version(), entry.Name(), entry.Checksum, batch); err != nil {
		return Record{}, fmt.Errorf("migration: recording start of %d (%s): %w", entry.Version(), entry.Name(), err)
	}
	if err := entry.Up(ctx, e.db); err != nil {
		return Record{}, fmt.Errorf("migration: applying %d (%s), left dirty for manual review: %w", entry.Version(), entry.Name(), err)
	}
	duration := time.Since(start)
	if err := hist.Complete(ctx, e.db, entry.Version(), duration); err != nil {
		return Record{}, fmt.Errorf("migration: recording completion of %d (%s): %w", entry.Version(), entry.Name(), err)
	}
	return Record{Version: entry.Version(), Name: entry.Name(), Checksum: entry.Checksum, AppliedAt: start, ExecutionTime: duration, Batch: batch}, nil
}

func (e *Engine) revert(ctx context.Context, entry Entry) error {
	hist := e.provider.History()

	if e.provider.SupportsTransactionalDDL() {
		txdb := e.db.(TxCapableDB)
		tx, err := txdb.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("migration: beginning transaction for %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		if err := entry.Down(ctx, tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration: reverting %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		if err := hist.Remove(ctx, tx, entry.Version()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration: removing history for %d (%s): %w", entry.Version(), entry.Name(), err)
		}
		return tx.Commit()
	}

	if err := entry.Down(ctx, e.db); err != nil {
		return fmt.Errorf("migration: reverting %d (%s): %w", entry.Version(), entry.Name(), err)
	}
	return hist.Remove(ctx, e.db, entry.Version())
}
