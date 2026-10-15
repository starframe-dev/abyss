// Package ascii converts a framebuffer into ANSI color art.
package ascii

import (
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
)

// Cell represents one terminal cell with a character and colors.
type Cell struct {
	Char         rune
	Extra        string
	Width        int
	Continuation bool
	FG           css.Color
	BG           css.Color
}

// Render converts a layout tree into plain-text lines for the given viewport
// width, using the framebuffer's cell contents.
func Render(fb *framebuffer.Framebuffer, tree *layout.LayoutBox, viewportWidth int) []string {
	viewportWidth = max(0, viewportWidth)
	if fb == nil || fb.Width == 0 {
		return []string{""}
	}

	grid := makeGrid(fb)

	// Overlay text
	overlayText(grid, tree)

	// Plain text for E2E compatibility
	return renderPlain(grid, viewportWidth)
}

// RenderANSI returns ANSI color art lines for terminal display.
func RenderANSI(fb *framebuffer.Framebuffer, tree *layout.LayoutBox, viewportWidth int) []string {
	viewportWidth = max(0, viewportWidth)
	if fb == nil || fb.Width == 0 {
		return []string{""}
	}

	grid := makeGrid(fb)

	overlayText(grid, tree)

	return renderANSI(grid, viewportWidth)
}

func makeGrid(fb *framebuffer.Framebuffer) [][]Cell {
	numRows := (fb.Height + cellmetrics.HeightPx - 1) / cellmetrics.HeightPx
	numCols := fb.Width / cellmetrics.WidthPx
	grid := make([][]Cell, numRows)
	for row := 0; row < numRows; row++ {
		grid[row] = make([]Cell, numCols)
		for col := 0; col < numCols; col++ {
			// One cell is sampled at its centre pixel.
			x := col*cellmetrics.WidthPx + cellmetrics.WidthPx/2
			y := row*cellmetrics.HeightPx + cellmetrics.HeightPx/2
			char, fg, bg := chooseChar(fb.Get(x, y), css.Transparent())
			grid[row][col] = Cell{Char: []rune(char)[0], FG: fg, BG: bg}
		}
	}
	return grid
}

func chooseChar(top, bottom css.Color) (string, css.Color, css.Color) {
	if !top.IsSet && !bottom.IsSet {
		return " ", css.Transparent(), css.Transparent()
	}
	// Use background color for non-text cells, foreground for text overlay
	fg := css.Transparent()
	if top.IsSet {
		bg := top
		return " ", fg, bg
	}
	return " ", fg, bottom
}

func colorsEqual(a, b css.Color) bool {
	return a.IsSet == b.IsSet && a.R == b.R && a.G == b.G && a.B == b.B
}

func renderANSI(grid [][]Cell, maxWidth int) []string {
	var lines []string
	for _, row := range grid {
		var b strings.Builder
		lastFG := css.Transparent()
		lastBG := css.Transparent()
		visibleCount := 0

		for _, cell := range row {
			if visibleCount >= maxWidth {
				break
			}
			if cell.Continuation {
				continue
			}

			if writeANSIColor(&b, cell.FG, cell.BG, lastFG, lastBG) {
				lastFG = cell.FG
				lastBG = cell.BG
			}

			// Determine the visible character
			var ch rune
			if cell.Char == 0 {
				ch = ' '
			} else {
				ch = cell.Char
			}
			if isPrintable(ch) {
				cellWidth := cell.Width
				if cellWidth <= 0 {
					cellWidth = max(1, runewidth.RuneWidth(ch))
				}
				if visibleCount+cellWidth > maxWidth {
					b.WriteByte(' ')
					visibleCount++
					continue
				}
				b.WriteRune(ch)
				writeSafeClusterTail(&b, cell.Extra)
				visibleCount += cellWidth
			} else {
				b.WriteRune('�')
				visibleCount++
			}
		}

		// Pad with spaces to maxWidth if needed
		for visibleCount < maxWidth {
			b.WriteRune(' ')
			visibleCount++
		}

		// Reset at end of line
		b.WriteString("\033[0m")

		lines = append(lines, b.String())
	}
	return lines
}

func renderPlain(grid [][]Cell, maxWidth int) []string {
	var lines []string
	for _, row := range grid {
		var b strings.Builder
		visibleWidth := 0
		for _, cell := range row {
			if cell.Continuation {
				continue
			}
			ch := cell.Char
			if ch == 0 {
				ch = ' '
			} else if !isPrintable(ch) {
				ch = '�'
			}
			width := cell.Width
			if width <= 0 {
				width = max(1, runewidth.RuneWidth(ch))
			}
			if visibleWidth+width > maxWidth {
				b.WriteByte(' ')
				visibleWidth++
				break
			}
			b.WriteRune(ch)
			writeSafeClusterTail(&b, cell.Extra)
			visibleWidth += width
		}
		lines = append(lines, b.String())
	}
	return lines
}

func writeSafeClusterTail(b *strings.Builder, tail string) {
	for _, r := range tail {
		if unicode.IsPrint(r) || r == 0x200c || r == 0x200d || (r >= 0xfe00 && r <= 0xfe0f) || (r >= 0xe0100 && r <= 0xe01ef) {
			b.WriteRune(r)
		}
	}
}

func writeANSIColor(b *strings.Builder, fg, bg, lastFG, lastBG css.Color) bool {
	if colorsEqual(fg, lastFG) && colorsEqual(bg, lastBG) {
		return false
	}
	if !fg.IsSet && !bg.IsSet {
		return false
	}
	b.WriteString("\033[0m\033[")
	if fg.IsSet {
		writeRGBParams(b, "38;2;", fg)
	}
	if bg.IsSet {
		if fg.IsSet {
			b.WriteByte(';')
		}
		writeRGBParams(b, "48;2;", bg)
	}
	b.WriteByte('m')
	return true
}

func writeRGBParams(b *strings.Builder, prefix string, c css.Color) {
	b.WriteString(prefix)
	writeByteDecimal(b, c.R)
	b.WriteByte(';')
	writeByteDecimal(b, c.G)
	b.WriteByte(';')
	writeByteDecimal(b, c.B)
}

func writeByteDecimal(b *strings.Builder, value uint8) {
	if value >= 100 {
		b.WriteByte('0' + value/100)
		value %= 100
		b.WriteByte('0' + value/10)
		b.WriteByte('0' + value%10)
		return
	}
	if value >= 10 {
		b.WriteByte('0' + value/10)
		b.WriteByte('0' + value%10)
		return
	}
	b.WriteByte('0' + value)
}

func isPrintable(r rune) bool {
	return r != 0x1b && r != 0x7f && !(r >= 0x80 && r <= 0x9f) && unicode.IsPrint(r)
}

func overlayText(grid [][]Cell, tree *layout.LayoutBox) {
	if tree == nil {
		return
	}

	overlayAsciiCells(grid, tree)

	var textLines []layout.TextLine
	var textColors []css.Color
	collectTextLines(tree, &textLines, &textColors)

	for i, tl := range textLines {
		color := textColors[i]
		if !color.IsSet {
			color = css.RGB(30, 30, 30)
		}

		row := int(tl.Y) / cellmetrics.HeightPx
		if row < 0 || row >= len(grid) {
			continue
		}

		col := int(tl.X) / cellmetrics.WidthPx
		if col < 0 {
			col = 0
		}

		graphemes := uniseg.NewGraphemes(tl.Text)
		for graphemes.Next() {
			cluster := graphemes.Str()
			cellWidth := graphemes.Width()
			if cellWidth <= 0 {
				continue
			}
			if col+cellWidth > len(grid[row]) {
				break
			}
			clusterRunes := []rune(cluster)
			// Keep original background from grid so text doesn't appear on black bg.
			origBG := grid[row][col].BG
			grid[row][col] = Cell{Char: clusterRunes[0], Extra: string(clusterRunes[1:]), Width: cellWidth, FG: color, BG: origBG}
			if cellWidth == 2 {
				continuationBG := grid[row][col+1].BG
				grid[row][col+1] = Cell{Continuation: true, FG: color, BG: continuationBG}
			}
			col += cellWidth
		}
	}
}

// overlayAsciiCells copies the pre-quantized ASCII grids from image boxes
// into the terminal grid, overwriting the placeholders generated by the
// framebuffer collapse step.
func overlayAsciiCells(grid [][]Cell, tree *layout.LayoutBox) {
	if tree == nil {
		return
	}
	for _, child := range tree.Children {
		overlayAsciiCells(grid, child)
	}
	if len(tree.AsciiCells) == 0 {
		return
	}
	startCol := int(tree.Border.X / cellmetrics.WidthPx)
	startRow := int(tree.Border.Y / cellmetrics.HeightPx)
	rows := tree.AsciiCells
	cols := 0
	if len(rows) > 0 {
		cols = len(rows[0])
	}
	for r, row := range rows {
		gridRow := startRow + r
		if gridRow < 0 || gridRow >= len(grid) {
			continue
		}
		for c, cell := range row {
			gridCol := startCol + c
			if gridCol < 0 || gridCol >= len(grid[gridRow]) {
				continue
			}
			grid[gridRow][gridCol] = Cell{
				Char: cell.Char,
				FG:   cell.FG,
				BG:   cell.BG,
			}
			_ = asciirender.Cell{} // keep import live for tree-shaking across builds
		}
	}
	_ = cols
}

func collectTextLines(box *layout.LayoutBox, lines *[]layout.TextLine, colors *[]css.Color) {
	for _, tl := range box.TextLines {
		*lines = append(*lines, tl)
		*colors = append(*colors, box.Style.Color)
	}
	for _, child := range box.Children {
		collectTextLines(child, lines, colors)
	}
}
