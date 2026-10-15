package layout

import (
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// layoutTable places the cells of each table row side by side. Columns take
// their percentage width of the table; the remaining width is shared equally
// between auto cells. Rows stack vertically using the tallest cell.
// Row and section boxes are not laid out; only cells get geometry.
func layoutTable(box *LayoutBox, contentWidth float64) float64 {
	var y float64
	for _, row := range tableRows(box) {
		cells := rowCells(row)
		if len(cells) == 0 {
			continue
		}
		widths := tableColumnWidths(cells, contentWidth)
		var x, rowHeight float64
		for i, cell := range cells {
			orig := cell.Style.Width
			cell.Style.Width = css.Px(widths[i])
			layoutBox(cell, widths[i], box.Content.X+x, box.Content.Y+y, false)
			cell.Style.Width = orig
			rowHeight = max(rowHeight, cell.Margin.Height)
			x += widths[i]
		}
		y += rowHeight
	}
	return y
}

// tableRows returns the tr elements of a table, looking through thead, tbody
// and tfoot sections.
func tableRows(box *LayoutBox) []*LayoutBox {
	var rows []*LayoutBox
	var walk func(b *LayoutBox)
	walk = func(b *LayoutBox) {
		if b.Node != nil && b.Node.Type == dom.ElementNode && b.Node.TagName == "tr" {
			rows = append(rows, b)
			return
		}
		for _, c := range b.Children {
			walk(c)
		}
	}
	walk(box)
	return rows
}

// rowCells returns the td and th children of a row in document order.
func rowCells(row *LayoutBox) []*LayoutBox {
	var cells []*LayoutBox
	for _, c := range row.Children {
		if c.Node != nil && c.Node.Type == dom.ElementNode && (c.Node.TagName == "td" || c.Node.TagName == "th") {
			cells = append(cells, c)
		}
	}
	return cells
}

// tableColumnWidths resolves each cell width against the table width.
func tableColumnWidths(cells []*LayoutBox, tableWidth float64) []float64 {
	widths := make([]float64, len(cells))
	var fixed float64
	var autoCount int
	for i, cell := range cells {
		if cell.Style.Width.IsAuto {
			autoCount++
			continue
		}
		widths[i] = cell.Style.Width.Resolve(tableWidth)
		fixed += widths[i]
	}
	if autoCount > 0 {
		share := max(0, tableWidth-fixed) / float64(autoCount)
		for i, cell := range cells {
			if cell.Style.Width.IsAuto {
				widths[i] = share
			}
		}
	}
	return widths
}
