package migration

import (
	"fmt"
	"io/fs"
	"sort"
)

// Load combines SQL migrations discovered in fsys with Go migrations from
// reg into a single, version-sorted list. reg may be nil if the project has
// no Go migrations.
func Load(fsys fs.FS, reg *Registry) ([]Entry, error) {
	sqlEntries, err := LoadSQLMigrations(fsys)
	if err != nil {
		return nil, err
	}

	seen := make(map[uint64]struct{}, len(sqlEntries))
	all := make([]Entry, 0, len(sqlEntries))
	for _, e := range sqlEntries {
		seen[e.Version()] = struct{}{}
		all = append(all, e)
	}

	if reg != nil {
		for _, e := range reg.All() {
			if _, dup := seen[e.Version()]; dup {
				return nil, fmt.Errorf("%w: %d", ErrDuplicateVersion, e.Version())
			}
			seen[e.Version()] = struct{}{}
			all = append(all, e)
		}
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Version() < all[j].Version() })
	return all, nil
}
