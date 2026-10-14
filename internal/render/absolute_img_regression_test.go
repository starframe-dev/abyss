package render

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

// An absolutely positioned icon whose pixels are unavailable must not paint a
// gray panel over the text field it sits on (google.com search box).
func TestAbsoluteImagePlaceholderDoesNotCoverInput(t *testing.T) {
	doc := dom.Parse(`<body><div style="position:relative;width:600px"><input name="q" size="57" style="height:30px;padding-right:38px"><img src="/tia.png" style="position:absolute;right:5px;top:4px"></div></body>`)
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
	if got != css.RGB(241, 243, 244) {
		t.Fatalf("pixel (%d,%d) = %v, want input surface #f1f3f4 (absolute icon covered it)", x, y, got)
	}
}
