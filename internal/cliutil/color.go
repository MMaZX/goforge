package cliutil

import (
	"os"
)

// colorEnabled determines whether ANSI color codes are emitted.
// Initialized checking the standard NO_COLOR environment variable and TERM=dumb.
var colorEnabled = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

// SetColorEnabled explicitly enables or disables ANSI color escape sequences.
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// IsColorEnabled returns whether color output is currently enabled.
func IsColorEnabled() bool {
	return colorEnabled
}

const (
	ansiReset      = "\033[0m"
	ansiBold       = "\033[1m"
	ansiDim        = "\033[2m"
	ansiRed        = "\033[31m"
	ansiGreen      = "\033[32m"
	ansiYellow     = "\033[33m"
	ansiBlue       = "\033[34m"
	ansiMagenta    = "\033[35m"
	ansiCyan       = "\033[36m"
	ansiBoldRed    = "\033[1;31m"
	ansiBoldYellow = "\033[1;33m"
	ansiBoldGreen  = "\033[1;32m"
)

func colorize(code, text string) string {
	if !colorEnabled || text == "" {
		return text
	}
	return code + text + ansiReset
}

// Red wraps text in ANSI red.
func Red(text string) string { return colorize(ansiRed, text) }

// BoldRed wraps text in ANSI bold red.
func BoldRed(text string) string { return colorize(ansiBoldRed, text) }

// Green wraps text in ANSI green.
func Green(text string) string { return colorize(ansiGreen, text) }

// BoldGreen wraps text in ANSI bold green.
func BoldGreen(text string) string { return colorize(ansiBoldGreen, text) }

// Yellow wraps text in ANSI yellow.
func Yellow(text string) string { return colorize(ansiYellow, text) }

// BoldYellow wraps text in ANSI bold yellow.
func BoldYellow(text string) string { return colorize(ansiBoldYellow, text) }

// Blue wraps text in ANSI blue.
func Blue(text string) string { return colorize(ansiBlue, text) }

// Cyan wraps text in ANSI cyan.
func Cyan(text string) string { return colorize(ansiCyan, text) }

// Bold wraps text in ANSI bold.
func Bold(text string) string { return colorize(ansiBold, text) }

// Dim wraps text in ANSI dim / gray.
func Dim(text string) string { return colorize(ansiDim, text) }
