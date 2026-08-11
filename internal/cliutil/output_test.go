package cliutil

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/MMaZX/goforge/internal/i18n"
)

func TestPrintStatusHumanGolden(t *testing.T) {
	rows := []StatusRow{
		{Version: 1, Name: "create_users", Applied: true, Batch: 1},
		{Version: 2, Name: "add_emails", Applied: false},
		{Version: 3, Name: "index_emails", Applied: true, Batch: 2, Dirty: true},
	}
	// Column layout: %06d_ + name padded to 40 + " " + mark.
	line := func(v int, name, mark string) string {
		return fmt.Sprintf("%06d_", v) + name + strings.Repeat(" ", 40-len(name)) + " " + mark + "\n"
	}

	t.Run("en", func(t *testing.T) {
		i18n.SetLanguage("en")
		t.Cleanup(func() { i18n.SetLanguage("en") })
		var buf bytes.Buffer
		PrintStatusHuman(&buf, rows)
		want := line(1, "create_users", "✓ applied (batch 1)") +
			line(2, "add_emails", "✗ pending") +
			line(3, "index_emails", "✓ applied (batch 2) [DIRTY]")
		if got := buf.String(); got != want {
			t.Errorf("en output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("es", func(t *testing.T) {
		i18n.SetLanguage("es")
		t.Cleanup(func() { i18n.SetLanguage("en") })
		var buf bytes.Buffer
		PrintStatusHuman(&buf, rows)
		want := line(1, "create_users", "✓ aplicada (lote 1)") +
			line(2, "add_emails", "✗ pendiente") +
			line(3, "index_emails", "✓ aplicada (lote 2) [DIRTY]")
		if got := buf.String(); got != want {
			t.Errorf("es output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})
}
