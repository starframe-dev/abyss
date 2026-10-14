package render

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

// An empty textarea must paint its surface like a text input.
func TestEmptyTextareaPaintsSurface(t *testing.T) {
	doc := dom.Parse(`<body><textarea name="q" rows="1"></textarea></body>`)
	box := layout.Build(doc, css.Compute(doc), 200)
	var ta *layout.LayoutBox
	var find func(b *layout.LayoutBox)
	find = func(b *layout.LayoutBox) {
		if b.Node != nil && b.Node.TagName == "textarea" {
			ta = b
		}
		for _, c := range b.Children {
			find(c)
		}
	}
	find(box)
	if ta == nil {
		t.Fatal("textarea box not found")
	}
	fb := Render(box, 200)
	x := int(ta.Border.X) + 1
	y := int(ta.Border.Y) + int(ta.Border.Height)/2
	if got := fb.Get(x, y); got != css.RGB(241, 243, 244) {
		t.Fatalf("pixel (%d,%d) = %v, want #f1f3f4", x, y, got)
	}
}
