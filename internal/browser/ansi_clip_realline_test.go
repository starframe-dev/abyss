package browser

import (
    "os"
    "strings"
    "testing"
)

// TestClipANSI_RealAbyssImageLine feeds the true line produced by bin/abyss
// through the horizontal-scroll clip and checks the picture survives.
//
// A real image line costs ~1300 raw runes for ~64 visible columns, which is
// exactly why the old rune-based slice truncated it after ~6 columns.
func TestClipANSI_RealAbyssImageLine(t *testing.T) {
    raw, err := os.ReadFile("/tmp/abyss-compact.ansi")
    if err != nil {
        t.Skip("no real abyss dump available")
    }
    var imgLine string
    for _, l := range strings.Split(string(raw), "\n") {
        if strings.Contains(l, "\x1b[38;2;") && strings.Contains(stripSGR(l), "###") {
            imgLine = l
            break
        }
    }
    if imgLine == "" {
        t.Fatal("no truecolor image line found in dump")
    }
    full := visibleWidth(imgLine)

    // The old bug: 120 raw runes covered only a sixth of the line.
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
