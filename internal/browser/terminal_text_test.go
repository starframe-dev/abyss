package browser

import (
	"strings"
	"testing"
)

func TestSafeTerminalTextReplacesNonPrintingSeparators(t *testing.T) {
	got := safeTerminalText("one\u2028two")
	if strings.ContainsRune(got, '\u2028') {
		t.Fatalf("line separator survived: %q", got)
	}
}

func TestSafeTerminalTextRemovesBidiFormatting(t *testing.T) {
	got := safeTerminalText("safe\u202ehidden")
	if strings.ContainsRune(got, '\u202e') {
		t.Fatalf("bidi formatting survived: %q", got)
	}
	if got := safeTerminalText("👩‍💻"); got != "👩‍💻" {
		t.Fatalf("emoji joiner changed: %q", got)
	}
}

func TestSafeTerminalTextRemovesControls(t *testing.T) {
	input := "url\x1b]52;c;clip\x07\x1b[2J\x9b31m\u0085"
	got := safeTerminalText(input)
	if strings.Contains(got, "\x1b") || strings.ContainsRune(got, '\x07') || strings.ContainsRune(got, '\x9b') || strings.ContainsRune(got, '\u0085') {
		t.Fatalf("control survived sanitization: %q", got)
	}
}
