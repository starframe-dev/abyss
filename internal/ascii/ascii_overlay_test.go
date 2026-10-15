package ascii

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
)

func TestRenderTextUsesTerminalCellWidth(t *testing.T) {
	fb := framebuffer.New(7*cellmetrics.WidthPx, 2*cellmetrics.HeightPx)
	tree := &layout.LayoutBox{TextLines: []layout.TextLine{{Text: "A猫e\u03011️⃣", X: 0, Y: 0}}}
	plain := Render(fb, tree, 7)[0]
	if got := runewidth.StringWidth(plain); got != 7 {
		t.Fatalf("plain display width = %d, want 7 (%q)", got, plain)
	}
	ansi := RenderANSI(fb, tree, 7)[0]
	if !strings.Contains(ansi, "A猫e\u03011️⃣") {
		t.Fatalf("ANSI output lost wide/combining text: %q", ansi)
	}
}

func TestRenderTextDropsBidiControlsButKeepsEmojiJoiner(t *testing.T) {
	fb := framebuffer.New(8*cellmetrics.WidthPx, 2*cellmetrics.HeightPx)
	emoji := "👩‍💻"
	tree := &layout.LayoutBox{TextLines: []layout.TextLine{{Text: "a\u202eB" + emoji, X: 0, Y: 0}}}
	plain := Render(fb, tree, 8)[0]
	if strings.ContainsRune(plain, '\u202e') {
		t.Fatalf("bidi control survived: %q", plain)
	}
	if !strings.Contains(plain, emoji) {
		t.Fatalf("emoji joiner was lost: %q", plain)
	}
}

func TestOverlayAsciiCells_ImgBoxesPopulateGrid(t *testing.T) {
	doc := dom.Parse(`<html><body><img src="https://example.test/x.png" alt="x"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 200)
	if box == nil {
		t.Fatal("layout nil")
	}

	var imgBox *layout.LayoutBox
	var walk func(*layout.LayoutBox)
	walk = func(b *layout.LayoutBox) {
		if b.Node != nil && strings.EqualFold(b.Node.TagName, "img") {
			imgBox = b
		}
		for _, c := range b.Children {
			walk(c)
		}
	}
	walk(box)
	if imgBox == nil {
		t.Fatal("img layout box not found")
	}

	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	grid := asciirender.Quantize(img, asciirender.Options{
		Columns: int(imgBox.Content.Width),
		Rows:    4,
		BGColor: css.RGB(0, 0, 0),
	})
	imgBox.AsciiCells = grid

	fb := framebuffer.New(int(box.Content.Width)+10, int(box.Content.Height)+10)
	cells := makeGrid(fb)
	overlayAsciiCells(cells, box)

	if cells == nil || len(cells) == 0 {
		t.Fatal("grid not initialised")
	}
	sawAscii := false
	for _, row := range cells {
		for _, cell := range row {
			if cell.Char != 0 && cell.Char != ' ' {
				sawAscii = true
			}
		}
	}
	if !sawAscii {
		t.Fatal("expected overlay to inject ASCII characters into grid")
	}
}
