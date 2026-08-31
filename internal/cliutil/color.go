package cliutil

import (
	"os"
	"strings"
)

// colorEnabled determines whether ANSI color codes are emitted. Checks the
// standard NO_COLOR environment variable, TERM=dumb, and that stdout is
// actually a terminal (so piping/redirecting output never leaks ANSI codes).
var colorEnabled = os.Getenv("NO_COLOR") == "" &&
	os.Getenv("TERM") != "dumb" &&
	IsTerminal(os.Stdout)

// SetColorEnabled explicitly enables or disables ANSI color escape sequences.
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// IsColorEnabled returns whether color output is currently enabled.
func IsColorEnabled() bool {
	return colorEnabled
}

// Primitives. Never used directly outside this file — every call site in
// cmd/ goes through the semantic roles below, so a color always carries one
// meaning across the whole CLI.
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBoldRed = "\033[1;31m"
	// ansiAccent is the project's violet accent, reproduced in 24-bit color
	// so it matches the reference palette exactly rather than approximating
	// it with one of the 8 standard ANSI hues.
	ansiAccent = "\033[38;2;196;160;245m"

	ansiBadgeSuccess = "\033[30;42m" // black on green
	ansiBadgeDanger  = "\033[97;41m" // bright white on red
)

func colorize(code, text string) string {
	if !colorEnabled || text == "" {
		return text
	}
	return code + text + ansiReset
}

// Success marks a positive per-item outcome (a passed check, an applied
// migration).
func Success(text string) string { return colorize(ansiGreen, text) }

// Danger marks a negative per-item outcome (a failed check, a dirty
// migration).
func Danger(text string) string { return colorize(ansiRed, text) }

// DangerBold marks an irreversible destructive action — reserved for the
// warning banners and last-chance confirmations of reset/fresh/rollback.
func DangerBold(text string) string { return colorize(ansiBoldRed, text) }

// Warning marks interactive caution: confirmation hints and cancellations.
func Warning(text string) string { return colorize(ansiYellow, text) }

// Accent marks a neutral interface pointer (a plan bullet). It is never a
// status color — success/failure always go through Success/Danger.
func Accent(text string) string { return colorize(ansiAccent, text) }

// Muted marks inactive or secondary text: skipped checks, pending
// migrations.
func Muted(text string) string { return colorize(ansiDim, text) }

// Bold adds structural emphasis (headers) without implying a color meaning.
func Bold(text string) string { return colorize(ansiBold, text) }

func badge(code, text string) string {
	if !colorEnabled || text == "" {
		return text
	}
	return code + " " + text + " " + ansiReset
}

// SuccessBadge renders text as a solid-background pill. Reserved for the
// single headline verdict of a run (e.g. "All checks passed") — individual
// items use Success instead, so a listing never turns into a wall of pills.
func SuccessBadge(text string) string { return badge(ansiBadgeSuccess, text) }

// DangerBadge is the failing counterpart of SuccessBadge, for a run's
// headline verdict.
func DangerBadge(text string) string { return badge(ansiBadgeDanger, text) }

// SuccessBadgeLine wraps a message that carries its own surrounding blank
// lines (as the migrate/rollback/reset "done" strings do) in a success
// badge, without the badge's background bleeding onto those blank lines.
func SuccessBadgeLine(text string) string {
	return "\n" + SuccessBadge(strings.TrimSpace(text)) + "\n"
}
