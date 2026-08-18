package cliutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// IsTerminal checks if the given reader is a character device (interactive terminal).
func IsTerminal(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// PromptStep represents a prompt question to display to the user.
type PromptStep struct {
	Message string
}

// PromptConfirmation runs one or more consecutive confirmation prompts against in/out.
// Every step must be answered with the exact expectedWord (strictly case-sensitive,
// e.g. "si" or "yes" in lowercase, no spaces).
//
// Returns:
// - (true, nil): User successfully confirmed all steps.
// - (false, nil): User entered something else (cancellation).
// - (false, err): I/O or scanner error (e.g. unexpected EOF).
func PromptConfirmation(in io.Reader, out io.Writer, prompts []string, expectedWord string) (bool, error) {
	scanner := bufio.NewScanner(in)
	for _, promptText := range prompts {
		fmt.Fprint(out, promptText)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, err
			}
			return false, io.EOF
		}
		line := scanner.Text()
		// Strip carriage return and line feed without stripping leading/trailing whitespace,
		// strictly enforcing exact match (e.g. no " si ", no "SI", only "si").
		line = strings.TrimRight(line, "\r\n")
		if line != expectedWord {
			return false, nil
		}
	}
	return true, nil
}
