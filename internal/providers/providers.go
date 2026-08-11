// Package providers is the single source of truth for which database
// drivers GoForge supports: their canonical name (as stored in
// goforge.yaml), accepted aliases, human label, and an example DSN with
// CHANGE_* placeholders so a project's .env can be scaffolded without
// anyone having to look up the driver's connection string syntax.
package providers

import (
	"fmt"
	"strings"
)

// Descriptor describes one supported database provider.
type Descriptor struct {
	// Driver is the canonical name stored in goforge.yaml's database.driver
	// and used internally once resolved (see Resolve).
	Driver string
	// Label is the human-readable name shown in CLI output.
	Label string
	// Aliases are other spellings --driver accepts for this Descriptor.
	Aliases []string
	// ExampleDSN is a template DSN with CHANGE_* placeholders, in the exact
	// syntax this driver expects, so a generated .env is immediately
	// editable without consulting external documentation.
	ExampleDSN string
	// VersionQuery returns the connected server's own version string.
	VersionQuery string
	// CurrentDatabaseQuery returns the name of the connected database.
	CurrentDatabaseQuery string
}

var registry = []Descriptor{
	{
		Driver:               "postgres",
		Label:                "PostgreSQL",
		Aliases:              []string{"postgresql", "pg"},
		ExampleDSN:           "postgres://CHANGE_USER:CHANGE_PASSWORD@CHANGE_HOST:5432/CHANGE_DATABASE?sslmode=disable",
		VersionQuery:         "SELECT version()",
		CurrentDatabaseQuery: "SELECT current_database()",
	},
	{
		Driver: "mariadb",
		Label:  "MariaDB",
		// mysql is accepted as an alias: go-sql-driver/mysql (which backs
		// database/mariadb) speaks the same wire protocol and DSN syntax
		// against a real MySQL server, so there is no need for a second,
		// separately-tested provider implementation.
		Aliases:              []string{"mysql"},
		ExampleDSN:           "CHANGE_USER:CHANGE_PASSWORD@tcp(CHANGE_HOST:3306)/CHANGE_DATABASE?parseTime=true",
		VersionQuery:         "SELECT VERSION()",
		CurrentDatabaseQuery: "SELECT DATABASE()",
	},
}

// List returns every supported provider, in a stable order.
func List() []Descriptor {
	out := make([]Descriptor, len(registry))
	copy(out, registry)
	return out
}

// Names returns the canonical driver names, for help text and error messages.
func Names() []string {
	names := make([]string, len(registry))
	for i, d := range registry {
		names[i] = d.Driver
	}
	return names
}

// Resolve normalizes a --driver/database.driver value (case-insensitively,
// accepting aliases) to its canonical Descriptor.
func Resolve(driver string) (Descriptor, error) {
	key := strings.ToLower(strings.TrimSpace(driver))
	for _, d := range registry {
		if d.Driver == key {
			return d, nil
		}
		for _, alias := range d.Aliases {
			if alias == key {
				return d, nil
			}
		}
	}
	return Descriptor{}, fmt.Errorf("unsupported driver %q (supported: %s)", driver, strings.Join(Names(), ", "))
}
