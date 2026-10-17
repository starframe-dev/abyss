package browser

import (
	"strings"
	"testing"
)

func TestSplitANSI_DropsOSCAndControlCharacters(t *testing.T) {
	input := "before\x1b]52;c;clipboard\x07after\x1b[2J\x1b[31mtext\x1b[0m\x9b2Jend"
	segs := splitANSI(input)
	var got strings.Builder
	for _, seg := range segs {
		got.WriteString(seg.text)
	}
	if got.String() != "beforeaftertextend" {
		t.Fatalf("unsafe controls survived: %q", got.String())
	}
}

func TestSplitANSI_SeparatesStyledRuns(t *testing.T) {
	segs := splitANSI("\x1b[38;2;220;30;30mAB\x1b[0mCD")
	if len(segs) != 2 {
		t.Fatalf("got %d segments, want 2", len(segs))
	}
	if segs[0].text != "AB" || segs[1].text != "CD" {
		t.Errorf("segment texts = %q, %q", segs[0].text, segs[1].text)
	}
	if len(segs[0].style) != 1 || len(segs[1].style) != 0 {
		t.Errorf("styles not tracked: %v / %v", segs[0].style, segs[1].style)
	}
}

func TestSplitANSI_RejectsUnsupportedAndInvalidSGR(t *testing.T) {
	input := "a\x1b[31mb\x1b[38;2;256;0;0mc\x1b[48;2;1;2;3;4md\x1b[38;2;1;2;3;48;2;4;5me"
	segs := splitANSI(input)
	var text strings.Builder
	for _, seg := range segs {
		text.WriteString(seg.text)
		if len(seg.style) != 0 {
			t.Errorf("invalid SGR unexpectedly retained style: %v", seg.style)
		}
	}
	if text.String() != "abcde" {
		t.Fatalf("text = %q, want abcde", text.String())
	}
}

func TestStripRawC1Controls(t *testing.T) {
	for _, input := range []string{
		"a\x9b31mb",
		"a\x90payload\x9cb",
		"a\x98payload\x1b\\b",
		"a\x9dtitle\x07b",
		"a\x9etext\x9cb",
		"a\x9ftext\x9cb",
	} {
		if got := stripRawC1Controls(input); got != "ab" {
			t.Errorf("stripRawC1Controls(%q) = %q, want ab", input, got)
		}
	}
}

func TestValidRendererSGR(t *testing.T) {
	for _, params := range []string{"", "0", "38;2;0;128;255", "48;2;1;2;3", "38;2;1;2;3;48;2;4;5;6"} {
		if !validRendererSGR(params) {
			t.Errorf("validRendererSGR(%q) = false", params)
		}
	}
	for _, params := range []string{"1", "31", "38;5;1", "38;2;256;0;0", "38;2;-1;0;0", "48;2;1;2", "38;2;1;x;3"} {
		if validRendererSGR(params) {
			t.Errorf("validRendererSGR(%q) = true", params)
		}
	}
}

func TestClipANSI_KeepsEscapeSequencesIntact(t *testing.T) {
	line := "\x1b[38;2;220;30;30mL" + "\x1b[38;2;0;0;0m###" + "\x1b[0m"
	got := clipANSI(line, 0, 2)
	if !strings.Contains(got, "\x1b[38;2;220;30;30m") {
		t.Errorf("foreground sequence was truncated: %q", got)
	}
	if visibleWidth(got) != 2 {
		t.Errorf("visible width = %d, want 2", visibleWidth(got))
	}
	if strings.Contains(stripSGR(got), "[") {
		t.Errorf("bracket leaked as printable text: %q", got)
	}
}

func TestClipANSI_NoDanglingEscape(t *testing.T) {
	line := strings.Repeat("\x1b[0m\x1b[38;2;220;30;30mL", 40)
	out := clipANSI(line, 0, 10)
	if visibleWidth(out) != 10 {
		t.Errorf("visible width = %d, want 10", visibleWidth(out))
	}
	if strings.HasSuffix(out, "\x1b") {
		t.Errorf("output ends with a dangling ESC: %q", out)
	}
	if strings.Contains(stripSGR(out), "[") {
		t.Errorf("escape fragments leaked as text: %q", out)
	}
	if got := strings.Count(out, "\x1b[38;2;220;30;30m"); got != 10 {
		t.Errorf("expected 10 SGR sequences, got %d in %q", got, out)
	}
}

func TestClipANSI_HorizontalScroll(t *testing.T) {
	line := "\x1b[38;2;255;0;0mABCDE\x1b[0mFGHIJ"
	got := clipANSI(line, 3, 4)
	if visibleWidth(got) != 4 {
		t.Errorf("clip window width = %d, want 4 (%q)", visibleWidth(got), got)
	}
	if plain := stripSGR(got); plain != "DEFG" {
		t.Errorf("clipped text = %q, want DEFG", plain)
	}
}

func TestClipANSI_BoundariesAndNonpositiveWidth(t *testing.T) {
	if got := clipANSI("abc", -2, 2); got != "ab" {
		t.Errorf("negative start clip = %q, want ab", got)
	}
	for _, width := range []int{0, -1} {
		if got := clipANSI("abc", 0, width); got != "" {
			t.Errorf("width %d returned %q, want empty", width, got)
		}
	}
	if got := clipANSI("abc", 9, 2); got != "" {
		t.Errorf("out-of-range clip = %q, want empty", got)
	}
}

func TestANSIClipUsesTerminalColumnsForWideRunes(t *testing.T) {
	line := "A猫B"
	if got := visibleWidth(line); got != 4 {
		t.Fatalf("visible width = %d, want 4", got)
	}
	if got := clipANSI(line, 1, 2); got != "猫" {
		t.Fatalf("whole wide glyph clip = %q", got)
	}
	if got := clipANSI(line, 2, 2); got != " B" {
		t.Fatalf("partial wide glyph clip = %q, want spacer and B", got)
	}
	emoji := "👩‍💻"
	if got := visibleWidth(emoji); got != 2 {
		t.Fatalf("emoji grapheme width = %d, want 2", got)
	}
	if got := clipANSI("A"+emoji+"B", 1, 2); got != emoji {
		t.Fatalf("emoji clip = %q, want %q", got, emoji)
	}
}

func TestVisibleWidth(t *testing.T) {
	if got := visibleWidth("\x1b[38;2;255;0;0mABC\x1b[0m"); got != 3 {
		t.Errorf("visibleWidth = %d, want 3", got)
	}
	if got := visibleWidth("plain"); got != 5 {
		t.Errorf("visibleWidth = %d, want 5", got)
	}
}

func stripSGR(s string) string {
	var b strings.Builder
	for _, seg := range splitANSI(s) {
		b.WriteString(seg.text)
	}
	return b.String()
}
