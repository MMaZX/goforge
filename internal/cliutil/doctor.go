package cliutil

import (
	"fmt"
	"io"
	"strings"
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

// PrintDoctorHuman renders a DoctorReport as a checklist for humans.
func PrintDoctorHuman(w io.Writer, report DoctorReport) {
	fmt.Fprintln(w, "GoForge doctor")
	fmt.Fprintln(w)
	for _, c := range report.Checks {
		mark := "✓"
		switch {
		case c.Skipped:
			mark = "…"
		case !c.OK:
			mark = "✗"
		}
		fmt.Fprintf(w, "%s %s\n", mark, c.Name)
		if c.Detail != "" {
			for _, line := range strings.Split(c.Detail, "\n") {
				fmt.Fprintf(w, "    %s\n", line)
			}
		}
	}
	fmt.Fprintln(w)
	if report.Healthy {
		fmt.Fprintln(w, "All checks passed.")
	} else {
		fmt.Fprintln(w, "Some checks failed.")
	}
}
