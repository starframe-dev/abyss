package render

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

// A text field is drawn as a pill like the browser search box: its corners are
// rounded to half its height, so a pixel near the corner stays unpainted.
func TestTextFieldCornersAreRounded(t *testing.T) {
	doc := dom.Parse(`<body><input name="q" size="57" style="height:30px"></body>`)
	box := layout.Build(doc, css.Compute(doc), 200)
	var input *layout.LayoutBox
	var find func(b *layout.LayoutBox)
	find = func(b *layout.LayoutBox) {
		if b.Node != nil && b.Node.TagName == "input" {
			input = b
		}
		for _, c := range b.Children {
			find(c)
		}
	}
	find(box)
	if input == nil {
		t.Fatal("input box not found")
	}
	fb := Render(box, 200)
	x := int(input.Border.X) + 4
	y := int(input.Border.Y) + 2
	if got := fb.Get(x, y); got.IsSet {
		t.Fatalf("corner pixel (%d,%d) is painted; field corners must be rounded: %+v", x, y, got)
	}
}
