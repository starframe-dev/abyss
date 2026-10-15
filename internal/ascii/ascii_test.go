package ascii

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
)

func TestRenderEmptyAndViewport(t *testing.T) {
	if got := Render(nil, nil, 5); len(got) != 1 || got[0] != "" {
		t.Fatalf("Render(nil) = %#v", got)
	}
	fb := &framebuffer.Framebuffer{Width: cellmetrics.WidthPx * 2, Height: cellmetrics.HeightPx}
	got := Render(fb, nil, 2)
	if len(got) != 1 || got[0] != "  " {
		t.Fatalf("Render blank = %#v", got)
	}
	if got := Render(fb, nil, -1); len(got) != 1 || got[0] != "" {
		t.Fatalf("Render negative viewport = %#v", got)
	}
}

func TestRenderTextAndANSI(t *testing.T) {
	fb := &framebuffer.Framebuffer{Width: cellmetrics.WidthPx * 4, Height: cellmetrics.HeightPx}
	tree := &layout.LayoutBox{TextLines: []layout.TextLine{{X: 0, Y: 0, Text: "Hi"}}, Style: css.Style{Color: css.RGB(1, 2, 3)}}
	got := Render(fb, tree, 4)
	if len(got) != 1 || got[0] != "Hi  " {
		t.Fatalf("Render text = %#v", got)
	}
	ansi := RenderANSI(fb, tree, 4)
	if len(ansi) != 1 || !strings.Contains(ansi[0], "\033[38;2;1;2;3mHi") || !strings.HasSuffix(ansi[0], "\033[0m") {
		t.Fatalf("RenderANSI = %#v", ansi)
	}
	if got := RenderANSI(nil, nil, 3); len(got) != 1 || got[0] != "" {
		t.Fatalf("RenderANSI(nil) = %#v", got)
	}
}

func TestRenderUnicodeAndClipping(t *testing.T) {
	fb := &framebuffer.Framebuffer{Width: cellmetrics.WidthPx * 3, Height: cellmetrics.HeightPx}
	tree := &layout.LayoutBox{TextLines: []layout.TextLine{{Text: "界x"}}}
	got := Render(fb, tree, 1)
	if len(got) != 1 || got[0] != " " {
		t.Fatalf("wide cluster clipping = %#v", got)
	}
	got = Render(fb, &layout.LayoutBox{TextLines: []layout.TextLine{{Text: "e\u0301"}}}, 2)
	if len(got) != 1 || got[0] != "e\u0301 " {
		t.Fatalf("grapheme output = %#v", got)
	}
}
