package render

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

// An empty text field must paint its surface so it is visible on the page.
func TestEmptyInputPaintsSurface(t *testing.T) {
	doc := dom.Parse(`<body><input name="q" size="57"></body>`)
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
	x := int(input.Border.X) + 1
	y := int(input.Border.Y) + int(input.Border.Height)/2
	got := fb.Get(x, y)
	if got == (css.Color{}) || got.IsSet == false {
		t.Fatalf("pixel (%d,%d) not painted: %+v", x, y, got)
	}
}
