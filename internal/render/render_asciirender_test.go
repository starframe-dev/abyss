package render

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

func newImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestRender_ImgPopulatesAsciiCellsAndFramebuffer(t *testing.T) {
	const src = "https://example.test/red.png"

	prev := ImageProvider
	ImageProvider = func(s string) image.Image {
		if s != src {
			return nil
		}
		return newImage(40, 20, color.RGBA{200, 30, 30, 255})
	}
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="` + src + `" alt="red"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	if box == nil {
		t.Fatal("layout box nil")
	}
	fb := Render(box, 120)
	if fb == nil {
		t.Fatal("framebuffer nil")
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
	if len(imgBox.AsciiCells) == 0 {
		t.Fatalf("AsciiCells empty for img box; expected grid to be populated")
	}
	if got := len(imgBox.AsciiCells[0]); got == 0 {
		t.Fatalf("AsciiCells row has zero columns")
	}
	// Regression: layoutImage seeds TextLines with the "[image] alt" placeholder.
	// Once cells exist the placeholder must be cleared, otherwise the text overlay
	// paints over the picture and erases its left columns.
	if len(imgBox.TextLines) != 0 {
		t.Errorf("TextLines must be cleared when AsciiCells are rendered, got %d lines", len(imgBox.TextLines))
	}
}

func TestRender_ImgKeepsPlaceholderTextWithoutProvider(t *testing.T) {
	prev := ImageProvider
	ImageProvider = nil
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="https://example.test/missing.png" alt="alt text"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	Render(box, 120)

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
	if len(imgBox.TextLines) == 0 {
		t.Fatal("placeholder TextLines must survive when no image is decoded")
	}
}

func TestRender_ImgWithoutProviderFallsBackToNeutralPanel(t *testing.T) {
	prev := ImageProvider
	ImageProvider = nil
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="https://example.test/x.png" alt="x"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	fb := Render(box, 120)
	if fb == nil {
		t.Fatal("framebuffer nil")
	}

	// Placeholder neutral panel color #e2e8f0 → some pixel must equal it.
	want := css.RGB(226, 232, 240)
	found := false
fbloop:
	for _, px := range fb.Pixels {
		if px == want {
			found = true
			break fbloop
		}
	}
	if !found {
		t.Fatalf("expected neutral placeholder pixels in framebuffer")
	}
}

func TestRender_ImgProviderReturnsNilFallsBack(t *testing.T) {
	prev := ImageProvider
	ImageProvider = func(string) image.Image { return nil }
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="https://example.test/x.png" alt="x"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	fb := Render(box, 120)
	if fb == nil {
		t.Fatal("framebuffer nil")
	}
	if box == nil {
		t.Fatal("layout box nil")
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
	if len(imgBox.AsciiCells) != 0 {
		t.Fatalf("AsciiCells should be empty when provider returns nil, got %d rows", len(imgBox.AsciiCells))
	}
}

func TestQuantizeImageBox_RespectsCachedCells(t *testing.T) {
	const src = "https://example.test/cached.png"

	prev := ImageProvider
	ImageProvider = func(s string) image.Image {
		return newImage(20, 20, color.RGBA{0, 200, 0, 255})
	}
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="` + src + `" alt="x"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)

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

	first := quantizeImageBox(imgBox)
	if first == nil {
		t.Fatal("first quantize returned nil")
	}
	second := quantizeImageBox(imgBox)
	if len(second) != len(first) || len(second[0]) != len(first[0]) {
		t.Fatalf("cached quantize differs from initial: %dx%d vs %dx%d",
			len(second), len(second[0]), len(first), len(first[0]))
	}
	for r := range first {
		for c := range first[r] {
			if first[r][c].Char != second[r][c].Char {
				t.Fatalf("cached grid differs at [%d][%d]: %q vs %q", r, c,
					first[r][c].Char, second[r][c].Char)
			}
		}
	}
}

func TestRender_ImgWritesForegroundPixels(t *testing.T) {
	const src = "https://example.test/blue.png"

	prev := ImageProvider
	ImageProvider = func(s string) image.Image {
		return newImage(20, 20, color.RGBA{10, 20, 250, 255})
	}
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="` + src + `" alt="x"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	fb := Render(box, 120)

	blueFound := false
	for _, px := range fb.Pixels {
		if px.R == 10 && px.G == 20 && px.B == 250 {
			blueFound = true
			break
		}
	}
	if !blueFound {
		t.Fatal("expected foreground blue pixels in framebuffer")
	}
}

func TestRender_ImgDimensionsMatchContent(t *testing.T) {
	const src = "https://example.test/dim.png"

	prev := ImageProvider
	ImageProvider = func(s string) image.Image {
		return newImage(60, 30, color.RGBA{120, 120, 120, 255})
	}
	defer func() { ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="` + src + `" alt="x"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	Render(box, 120)

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

	if len(imgBox.AsciiCells) == 0 || len(imgBox.AsciiCells[0]) == 0 {
		t.Fatal("AsciiCells empty")
	}
	cols := len(imgBox.AsciiCells[0])
	expectedCols := int(imgBox.Content.Width)
	if expectedCols < 1 {
		expectedCols = 1
	}
	if cols != expectedCols {
		t.Fatalf("AsciiCells columns %d, want %d", cols, expectedCols)
	}
	_ = asciirender.Palette
}
