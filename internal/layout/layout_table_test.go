package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

func TestTableRowsAndCells(t *testing.T) {
	tr := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "tr"}}
	td := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "td"}}
	th := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "th"}}
	text := &LayoutBox{Node: &dom.Node{Type: dom.TextNode, TagName: "tr"}}
	tr.Children = []*LayoutBox{td, text, th, {Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}}}
	section := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "tbody"}, Children: []*LayoutBox{tr}}
	root := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "table"}, Children: []*LayoutBox{section}}
	require.Equal(t, []*LayoutBox{tr}, tableRows(root))
	require.Equal(t, []*LayoutBox{td, th}, rowCells(tr))
	require.Empty(t, rowCells(text))
}

func TestTableColumnWidths(t *testing.T) {
	cells := []*LayoutBox{
		{Style: css.ComputedStyle{Width: css.Px(30)}},
		{Style: css.ComputedStyle{Width: css.Auto()}},
		{Style: css.ComputedStyle{Width: css.Auto()}},
	}
	require.Equal(t, []float64{30, 35, 35}, tableColumnWidths(cells, 100))
	require.Equal(t, []float64{30, 0, 0}, tableColumnWidths(cells, 20)[:1])
	auto := []*LayoutBox{{Style: css.ComputedStyle{Width: css.Auto()}}, {Style: css.ComputedStyle{Width: css.Auto()}}}
	require.Equal(t, []float64{0, 0}, tableColumnWidths(auto, -5))
}

func TestLayoutTableRows(t *testing.T) {
	style := css.Default()
	style.Width = css.Auto()
	row := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "tr"}, Children: []*LayoutBox{
		{Node: &dom.Node{Type: dom.ElementNode, TagName: "td"}, Style: style},
		{Node: &dom.Node{Type: dom.ElementNode, TagName: "td"}, Style: style},
	}}
	table := &LayoutBox{Node: &dom.Node{Type: dom.ElementNode, TagName: "table"}, Style: css.Default(), Content: Rect{X: 5, Y: 7}, Children: []*LayoutBox{{Node: &dom.Node{Type: dom.ElementNode, TagName: "tbody"}, Children: []*LayoutBox{row}}}}
	require.Equal(t, 0.0, layoutTable(table, 100))
	require.Equal(t, 5.0, row.Children[1].Content.X)
	require.Equal(t, css.Auto(), row.Children[0].Style.Width)
}
