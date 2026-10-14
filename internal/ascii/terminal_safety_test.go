package ascii

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
)

func TestRenderANSIReplacesTerminalControlsInPageText(t *testing.T) {
	fb := framebuffer.New(12, 2)
	text := "ok\x1b]52;c;payload\x07\x1b[2J\x9b31m\u009c"
	box := &layout.LayoutBox{TextLines: []layout.TextLine{{X: 0, Y: 0, Text: text}}, Style: css.ComputedStyle{Color: css.RGB(1, 2, 3)}}
	line := strings.Join(RenderANSI(fb, box, 12), "")
	if strings.Contains(line, "\x1b]") || strings.Contains(line, "\x1b[2J") || strings.ContainsRune(line, '\x9b') || strings.ContainsRune(line, '\x07') {
		t.Fatalf("page text emitted terminal control: %q", line)
	}
}
