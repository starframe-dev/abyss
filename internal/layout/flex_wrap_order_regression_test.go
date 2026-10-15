package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com's footer is a row with flex-wrap: wrap, so the link groups stack
// onto a second line when the row is too narrow.
func TestFlexWrapMovesItemsToNextLine(t *testing.T) {
	doc := dom.Parse(`<body><div style="display:flex;flex-wrap:wrap"><div style="width:600px">a</div><div style="width:600px">b</div></div></body>`)
	box := Build(doc, css.Compute(doc), 800)
	divs := findAllTags(box, "div")
	require.Len(t, divs, 3)
	a, b := divs[1], divs[2]
	require.Equal(t, a.Border.X, b.Border.X, "wrapped item must start at the left edge")
	require.Greater(t, b.Border.Y, a.Border.Y, "wrapped item must sit on a lower line")
}

// CSS order reorders flex items; the DOM order stays the same.
func TestFlexOrderReordersItems(t *testing.T) {
	doc := dom.Parse(`<body><div style="display:flex"><div style="width:100px;order:2">x</div><div style="width:100px;order:1">y</div></div></body>`)
	box := Build(doc, css.Compute(doc), 800)
	divs := findAllTags(box, "div")
	require.Len(t, divs, 3)
	first, second := divs[1], divs[2]
	require.Equal(t, "x", first.Node.Children[0].Text, "DOM order must stay")
	require.Less(t, second.Border.X, first.Border.X, "order:1 must come before order:2")
}
