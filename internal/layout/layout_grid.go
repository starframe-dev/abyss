// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
    "github.com/starframe/abyss/internal/css"
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

    colWidths := resolveGridColumnWidths(columns, contentWidth, gap)
    rowHeights := placeGridChildren(box, columns, rows, colWidths, gap, contentWidth)
    return sumRowHeights(rowHeights, gap)
}

func resolveGridColumnWidths(columns []css.Length, contentWidth, gap float64) []float64 {
    var colWidths []float64
    var totalFixedWidth, totalFr float64
    for _, col := range columns {
        if col.IsPercent {
            w := col.Resolve(contentWidth)
            colWidths = append(colWidths, w)
            totalFixedWidth += w
        } else if !col.IsAuto {
            w := col.Px
            colWidths = append(colWidths, w)
            totalFixedWidth += w
        } else {
            // fr unit: treat as 1fr for simplicity
            colWidths = append(colWidths, 0)
            totalFr += col.Px
        }
    }

    // Distribute remaining width to fr columns
    remainingWidth := contentWidth - totalFixedWidth - gap*float64(len(columns)-1)
    if remainingWidth < 0 {
        remainingWidth = 0
    }
    for i, col := range columns {
        if col.IsAuto && totalFr > 0 {
            colWidths[i] = remainingWidth * col.Px / totalFr
        }
    }
    return colWidths
}

func placeGridChildren(box *LayoutBox, columns, rows []css.Length, colWidths []float64, gap, contentWidth float64) []float64 {
    numCols := len(columns)
    var rowHeights []float64
    for i, child := range box.Children {
        col := i % numCols
        row := i / numCols

        // Ensure row height exists
        for len(rowHeights) <= row {
            rowHeights = append(rowHeights, rows[len(rowHeights)%len(rows)].Resolve(contentWidth))
        }

        x := gridOffsetX(box.Content.X, colWidths, gap, col)
        y := gridOffsetY(box.Content.Y, rowHeights, gap, row)

        origWidth := child.Style.Width
        child.Style.Width = css.Px(colWidths[col])
        layoutBox(child, colWidths[col], x, y, false)
        child.Style.Width = origWidth

        // Update row height if child is taller
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
    var totalHeight float64
    for _, h := range rowHeights {
        totalHeight += h + gap
    }
    if len(rowHeights) > 0 {
        totalHeight -= gap
    }
    return totalHeight
}
