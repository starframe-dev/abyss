package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// Cells of one table row sit side by side in columns, not stacked as blocks.
func TestTableRowCellsAreSideBySide(t *testing.T) {
	doc := dom.Parse(`<body><table><tr><td>aaaa</td><td>bbbb</td></tr></table></body>`)
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
	require.Greater(t, cells[1].Border.X, cells[0].Border.X, "second cell must start to the right of the first")
	require.LessOrEqual(t, cells[0].Border.Width+cells[1].Border.Width, 200.5, "cells overflow the table width")
}
