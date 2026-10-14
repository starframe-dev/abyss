package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

// A wide, short raster image must fill the rows its layout box reserves, not
// collapse to the single row of its alt-text placeholder.
func TestImageCellsUseLaidOutHeight(t *testing.T) {
	saved := ImageProvider
	defer func() { ImageProvider = saved }()
	src := image.NewRGBA(image.Rect(0, 0, 272, 92))
	for y := 0; y < 92; y++ {
		for x := 0; x < 272; x++ {
			src.Set(x, y, color.RGBA{R: 20, G: 20, B: 20, A: 255})
		}
	}
	ImageProvider = func(string) image.Image { return src }

	doc := dom.Parse(`<body><img alt="Google" src="logo.png" width="272" height="92"></body>`)
	box := layout.Build(doc, css.Compute(doc), 200)
	var img *layout.LayoutBox
	var find func(b *layout.LayoutBox)
	find = func(b *layout.LayoutBox) {
		if b.Node != nil && b.Node.TagName == "img" {
			img = b
		}
		for _, c := range b.Children {
			find(c)
		}
	}
	find(box)
	if img == nil {
		t.Fatal("img box not found")
	}
	cells := quantizeImageBox(img)
	if len(cells) < 20 {
		t.Fatalf("image rows = %d, want at least 20 (laid-out height %v px)", len(cells), img.Content.Height)
	}
}
