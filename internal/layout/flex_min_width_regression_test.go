package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// Percentage min-widths of flex items share the row: when their sum fits the
// row, each item gets at least its minimum and the second item stays on screen.
func TestFlexPercentMinWidthKeepsItemsInRow(t *testing.T) {
	doc := dom.Parse(`<body><div style="display:flex"><div style="width:10px;min-width:40%">a</div><div style="width:10px;min-width:40%">b</div></div></body>`)
	box := Build(doc, css.Compute(doc), 800)
	divs := findAllTags(box, "div")
	require.Len(t, divs, 3)
	require.GreaterOrEqual(t, divs[1].Border.Width, 320.0-0.5, "first item keeps its min-width")
	right := divs[2].Border.X + divs[2].Border.Width
	require.LessOrEqual(t, right, 800.0+0.5, "second item must stay inside the row")
}
