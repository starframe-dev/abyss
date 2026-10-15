package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// An absolutely positioned child inside a cell must not move or shrink the
// sibling cell, which keeps its own column.
func TestAbsoluteChildInCellDoesNotMoveSiblingCell(t *testing.T) {
	doc := dom.Parse(`<body><table><tr><td><div style="position:relative"><span style="position:absolute;right:5px;top:4px">x</span>aaaa</div></td><td>bbbb</td></tr></table></body>`)
	box := Build(doc, css.Compute(doc), 200)
	var cells []*LayoutBox
	var walk func(b *LayoutBox)
	walk = func(b *LayoutBox) {
		if b.Node != nil && b.Node.Type == dom.ElementNode && b.Node.TagName == "td" {
			cells = append(cells, b)
		}
		for _, c := range b.Children {
			walk(c)
		}
	}
	walk(box)
	require.Len(t, cells, 2)
	require.Greater(t, cells[1].Border.X, cells[0].Border.X, "absolute child pushed the second cell")
}
