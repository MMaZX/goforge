package migration

import "errors"

var (
	// ErrChecksumMismatch is returned when a previously applied migration's
	// content no longer matches the checksum recorded at apply time.
	ErrChecksumMismatch = errors.New("migration: checksum mismatch, applied migration was modified")
	// ErrDirtyState is returned when the history table contains a migration
	// left in an incomplete (dirty) state by a previous failed run.
	ErrDirtyState = errors.New("migration: dirty migration state, manual intervention required")
	// ErrDuplicateVersion is returned when two migrations declare the same version.
	ErrDuplicateVersion = errors.New("migration: duplicate migration version")
	// ErrLocked is returned when the migration lock could not be acquired.
	ErrLocked = errors.New("migration: could not acquire migration lock")
	// ErrNoMigrations is returned when no migrations are available to run.
	ErrNoMigrations = errors.New("migration: no migrations found")
	// ErrAlreadyApplied is returned when attempting to reapply a migration
	// that is already recorded as applied.
	ErrAlreadyApplied = errors.New("migration: already applied")
	// ErrNotApplied is returned when attempting to roll back a migration
	// that was never applied.
	ErrNotApplied = errors.New("migration: not applied")
	// ErrOutOfOrder is returned when a pending migration has a lower version
	// than one already applied, and out-of-order execution was not allowed.
	ErrOutOfOrder = errors.New("migration: out-of-order migration")
)
