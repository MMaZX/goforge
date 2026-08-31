package cliutil

import (
	"fmt"
	"io"
	"strings"

	"github.com/MMaZX/goforge/internal/i18n"
)

// DoctorCheck is one diagnostic check `goforge doctor` ran.
type DoctorCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Skipped bool   `json:"skipped,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// DoctorReport is the full result of `goforge doctor`.
type DoctorReport struct {
	Healthy bool          `json:"healthy"`
	Checks  []DoctorCheck `json:"checks"`
}

// doctorCheckKeys maps the stable English check names (which double as the
// JSON "name" field) to their i18n keys. Unknown names pass through
// untranslated, so adding a check never breaks human output.
var doctorCheckKeys = map[string]string{
	"Config":               "doctor.check.config",
	".env":                 "doctor.check.env",
	"Migrations directory": "doctor.check.migrations_dir",
	"Database connection":  "doctor.check.db_connection",
	"Migration history":    "doctor.check.history",
	"Locking":              "doctor.check.locking",
}

// PrintDoctorHuman renders a DoctorReport as a checklist for humans, in the
// active language. Name and Detail stay in English inside the struct so the
// --json report remains a stable machine contract; only the fixed
// boilerplate and the known check names are translated here, at print time.
func PrintDoctorHuman(w io.Writer, report DoctorReport) {
	fmt.Fprintln(w, i18n.T("doctor.header"))
	fmt.Fprintln(w)
	for _, c := range report.Checks {
		mark := Success("✓")
		switch {
		case c.Skipped:
			mark = Muted("…")
		case !c.OK:
			mark = Danger("✗")
		}
		name := c.Name
		if key, ok := doctorCheckKeys[c.Name]; ok {
			name = i18n.T(key)
		}
		fmt.Fprintf(w, "%s %s\n", mark, name)
		if c.Detail != "" {
			for _, line := range strings.Split(c.Detail, "\n") {
				fmt.Fprintf(w, "    %s\n", line)
			}
		}
	}
	fmt.Fprintln(w)
	if report.Healthy {
		fmt.Fprintln(w, SuccessBadge(i18n.T("doctor.all_passed")))
	} else {
		fmt.Fprintln(w, DangerBadge(i18n.T("doctor.some_failed")))
	}
}
