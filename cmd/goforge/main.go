// Command goforge is the standalone CLI for the GoForge migration engine.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/MMaZX/goforge/internal/cliutil"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		var cfgErr *cliutil.ConfigError
		if errors.As(err, &cfgErr) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
