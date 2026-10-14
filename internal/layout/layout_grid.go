// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
	"math"

	"github.com/starframe-dev/abyss/internal/css"
)

func layoutGrid(box *LayoutBox, contentWidth float64) float64 {
	style := box.Style
	gap := style.Gap.Resolve(contentWidth)
	columns := style.GridTemplateColumns
	if len(columns) == 0 {
		columns = []css.Length{css.Px(contentWidth)}
	}
	rows := style.GridTemplateRows
	if len(rows) == 0 {
		rows = []css.Length{css.Px(100)}
	}
	rows = resolveGridFractionalRows(rows, style.Height.Resolve(contentWidth), gap)
	colWidths := resolveGridColumnWidths(columns, contentWidth, gap)
	rowHeights := placeGridChildren(box, columns, rows, colWidths, gap, contentWidth)
	return sumRowHeights(rowHeights, gap)
}

func resolveGridColumnWidths(columns []css.Length, contentWidth, gap float64) []float64 {
	widths := make([]float64, len(columns))
	var fixed, totalWeight float64
	for i, col := range columns {
		switch {
		case col.IsPercent:
			widths[i] = col.Resolve(contentWidth)
			fixed += widths[i]
		case col.IsFraction:
			weight := col.Fraction
			if !finiteNonNegative(weight) {
				weight = 0
			}
			widths[i] = -weight
			totalWeight += weight
		case col.IsAuto:
			widths[i] = 0
		default:
			widths[i] = col.Px
			fixed += widths[i]
		}
	}
	remaining := max(0, contentWidth-fixed-gap*float64(len(columns)-1))
	if totalWeight > 0 {
		for i := range widths {
			if widths[i] < 0 {
				widths[i] = remaining * -widths[i] / totalWeight
			}
		}
	}
	return widths
}

func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func resolveGridFractionalRows(rows []css.Length, availableHeight, gap float64) []css.Length {
	var fixedHeight, totalWeight float64
	for _, row := range rows {
		if row.IsFraction {
			if finiteNonNegative(row.Fraction) {
				totalWeight += row.Fraction
			}
		} else {
			fixedHeight += row.Resolve(availableHeight)
		}
	}
	if totalWeight <= 0 {
		resolved := append([]css.Length(nil), rows...)
		for i, row := range rows {
			if row.IsPercent {
				resolved[i] = css.Px(row.Resolve(availableHeight))
			}
		}
		return resolved
	}
	remaining := max(0, availableHeight-fixedHeight-gap*float64(len(rows)-1))
	resolved := make([]css.Length, len(rows))
	for i, row := range rows {
		if row.IsFraction && finiteNonNegative(row.Fraction) {
			resolved[i] = css.Px(remaining * row.Fraction / totalWeight)
		} else if row.IsPercent {
			resolved[i] = css.Px(row.Resolve(availableHeight))
		} else {
			resolved[i] = row
		}
	}
	return resolved
}

func placeGridChildren(box *LayoutBox, columns, rows []css.Length, colWidths []float64, gap, contentWidth float64) []float64 {
	numCols := len(columns)
	rowHeights := make([]float64, 0, (len(box.Children)+numCols-1)/numCols)
	for i, child := range box.Children {
		col, row := i%numCols, i/numCols
		for len(rowHeights) <= row {
			rowHeights = append(rowHeights, rows[len(rowHeights)%len(rows)].Resolve(contentWidth))
		}
		x := gridOffsetX(box.Content.X, colWidths, gap, col)
		y := gridOffsetY(box.Content.Y, rowHeights, gap, row)
		origWidth := child.Style.Width
		child.Style.Width = css.Px(colWidths[col])
		layoutBox(child, colWidths[col], x, y, false)
		child.Style.Width = origWidth
		if child.Margin.Height > rowHeights[row] {
			rowHeights[row] = child.Margin.Height
		}
	}
	return rowHeights
}

func gridOffsetX(base float64, colWidths []float64, gap float64, col int) float64 {
	x := base
	for c := 0; c < col; c++ {
		x += colWidths[c] + gap
	}
	return x
}

func gridOffsetY(base float64, rowHeights []float64, gap float64, row int) float64 {
	y := base
	for r := 0; r < row; r++ {
		y += rowHeights[r] + gap
	}
	return y
}

func sumRowHeights(rowHeights []float64, gap float64) float64 {
	var total float64
	for _, height := range rowHeights {
		total += height + gap
	}
	if len(rowHeights) > 0 {
		total -= gap
	}
	return total
}
