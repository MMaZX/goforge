package cliutil

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestPromptConfirmationSingleSuccess(t *testing.T) {
	in := strings.NewReader("si\n")
	var out bytes.Buffer

	prompts := []string{"Confirm: "}
	ok, err := PromptConfirmation(in, &out, prompts, "si")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected true for exact 'si' match")
	}
	if out.String() != "Confirm: " {
		t.Errorf("got prompt output %q, want %q", out.String(), "Confirm: ")
	}
}

func TestPromptConfirmationStrictLowercaseRejectsUppercase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase SI", "SI\n", "si"},
		{"mixed case Si", "Si\n", "si"},
		{"leading space", " si\n", "si"},
		{"trailing space", "si \n", "si"},
		{"uppercase YES", "YES\n", "yes"},
		{"mixed case Yes", "Yes\n", "yes"},
		{"arbitrary text", "no\n", "si"},
		{"empty enter", "\n", "si"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			var out bytes.Buffer
			ok, err := PromptConfirmation(in, &out, []string{"Prompt: "}, tt.expected)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Errorf("input %q must be rejected for expected word %q", tt.input, tt.expected)
			}
		})
	}
}

func TestPromptConfirmationDoublePrompts(t *testing.T) {
	t.Run("both match", func(t *testing.T) {
		in := strings.NewReader("si\nsi\n")
		var out bytes.Buffer
		prompts := []string{"Step 1: ", "Step 2: "}
		ok, err := PromptConfirmation(in, &out, prompts, "si")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatalf("expected true when both match")
		}
		if out.String() != "Step 1: Step 2: " {
			t.Errorf("got %q, want %q", out.String(), "Step 1: Step 2: ")
		}
	})

	t.Run("step 1 fails", func(t *testing.T) {
		in := strings.NewReader("no\nsi\n")
		var out bytes.Buffer
		prompts := []string{"Step 1: ", "Step 2: "}
		ok, err := PromptConfirmation(in, &out, prompts, "si")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatalf("expected false when step 1 fails")
		}
		// Step 2 should not even be prompted
		if out.String() != "Step 1: " {
			t.Errorf("got %q, want only step 1", out.String())
		}
	})

	t.Run("step 2 fails", func(t *testing.T) {
		in := strings.NewReader("si\nSI\n") // Step 2 is uppercase
		var out bytes.Buffer
		prompts := []string{"Step 1: ", "Step 2: "}
		ok, err := PromptConfirmation(in, &out, prompts, "si")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatalf("expected false when step 2 is uppercase")
		}
		if out.String() != "Step 1: Step 2: " {
			t.Errorf("got %q, want both prompted", out.String())
		}
	})

	t.Run("unexpected EOF", func(t *testing.T) {
		in := strings.NewReader("")
		var out bytes.Buffer
		ok, err := PromptConfirmation(in, &out, []string{"Step 1: "}, "si")
		if err != io.EOF {
			t.Fatalf("expected io.EOF, got %v", err)
		}
		if ok {
			t.Fatalf("expected false on EOF")
		}
	})
}

func TestColorHelpers(t *testing.T) {
	SetColorEnabled(true)
	defer SetColorEnabled(false)

	if !strings.Contains(Danger("fail"), "\033[31m") {
		t.Errorf("Danger did not contain ANSI code")
	}
	if !strings.Contains(DangerBold("fail"), "\033[1;31m") {
		t.Errorf("DangerBold did not contain ANSI code")
	}
	if !strings.Contains(Warning("caution"), "\033[33m") {
		t.Errorf("Warning did not contain ANSI code")
	}
	if !strings.Contains(Success("ok"), "\033[32m") {
		t.Errorf("Success did not contain ANSI code")
	}
	if !strings.Contains(Muted("secondary"), "\033[2m") {
		t.Errorf("Muted did not contain ANSI code")
	}
	if !strings.Contains(Accent(">"), "\033[38;2;196;160;245m") {
		t.Errorf("Accent did not contain the exact true-color accent code")
	}

	if got := SuccessBadge("All checks passed."); !strings.Contains(got, "\033[30;42m") || !strings.Contains(got, " All checks passed. ") {
		t.Errorf("SuccessBadge did not render a padded green pill, got %q", got)
	}
	if got := DangerBadge("Some checks failed."); !strings.Contains(got, "\033[97;41m") || !strings.Contains(got, " Some checks failed. ") {
		t.Errorf("DangerBadge did not render a padded red pill, got %q", got)
	}
	if got := SuccessBadgeLine("\n3 migrations applied successfully.\n"); got != "\n"+SuccessBadge("3 migrations applied successfully.")+"\n" {
		t.Errorf("SuccessBadgeLine did not trim surrounding blank lines before badging, got %q", got)
	}

	SetColorEnabled(false)
	if Danger("fail") != "fail" {
		t.Errorf("Danger should not colorize when disabled, got %q", Danger("fail"))
	}
	if SuccessBadge("ok") != "ok" {
		t.Errorf("SuccessBadge should not add padding/background when disabled, got %q", SuccessBadge("ok"))
	}
}
