package migration

import "context"

// Migration is a single, versioned schema change. Implementations are
// provided either as Go code (registered explicitly via a Registry) or
// generated from a pair of .up.sql / .down.sql files.
type Migration interface {
	Version() uint64
	Name() string
	Up(ctx context.Context, db DB) error
	Down(ctx context.Context, db DB) error
}

// Direction identifies which side of a migration is being executed.
type Direction int

const (
	Up Direction = iota
	Down
)

func (d Direction) String() string {
	if d == Down {
		return "down"
	}
	return "up"
}
