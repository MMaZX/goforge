// Command goforge is the standalone CLI for the GoForge migration engine.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/i18n"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("app.error"), err)
		var cfgErr *cliutil.ConfigError
		if errors.As(err, &cfgErr) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
