package browser

import (
	"strings"
	"testing"
)

// TestClipANSI_TrueColorImageLine feeds a deterministic renderer-style image
// line through horizontal clipping and checks that every visible cell survives.
//
// Renderer-style color sequences make raw rune count much larger than the
// number of visible columns, which defeated the old rune-based clipping.
func TestClipANSI_TrueColorImageLine(t *testing.T) {
	var line strings.Builder
	for i := 0; i < 64; i++ {
		line.WriteString("\x1b[38;2;120;40;200m#\x1b[0m")
	}
	imgLine := line.String()
	full := visibleWidth(imgLine)

	// The old bug: a 120-rune slice covered only a few visible columns.
	if len([]rune(imgLine)) < 3*full {
		t.Fatalf("dump line unexpectedly compact: %d runes for %d columns", len([]rune(imgLine)), full)
	}

	if got := clipANSI(imgLine, 0, 120); visibleWidth(got) != full {
		t.Errorf("clip at offset 0 lost columns: got %d, want %d", visibleWidth(got), full)
	}
	if got := clipANSI(imgLine, 20, 120); visibleWidth(got) != full-20 {
		t.Errorf("clip at offset 20 = %d, want %d", visibleWidth(got), full-20)
	}
	if got := clipANSI(imgLine, 0, 40); visibleWidth(got) != 40 {
		t.Errorf("clip window = %d, want 40", visibleWidth(got))
	}
	if got := clipANSI(imgLine, 0, 120); strings.Contains(stripSGR(got), "[") {
		t.Errorf("escape fragments leaked as printable text: %q", got)
	}
}
