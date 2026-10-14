// Package ascii converts a framebuffer into ANSI color art.
package ascii

import (
    "strings"

    "github.com/starframe/abyss/internal/asciirender"
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/framebuffer"
    "github.com/starframe/abyss/internal/layout"
)

// Cell represents one terminal cell with a character and colors.
type Cell struct {
    Char rune
    FG   css.Color
    BG   css.Color
}

// Render converts a layout tree into plain-text lines for the given viewport
// width, using the framebuffer's cell contents.
func Render(fb *framebuffer.Framebuffer, tree *layout.LayoutBox, viewportWidth int) []string {
    if fb == nil || fb.Width == 0 {
        return []string{""}
    }

    numRows := fb.Height / 2
    if fb.Height%2 != 0 {
        numRows++
    }

    grid := makeGrid(fb, numRows)

    // Overlay text
    overlayText(grid, tree)

    // Plain text for E2E compatibility
    return renderPlain(grid, viewportWidth)
}

// RenderANSI returns ANSI color art lines for terminal display.
func RenderANSI(fb *framebuffer.Framebuffer, tree *layout.LayoutBox, viewportWidth int) []string {
    if fb == nil || fb.Width == 0 {
        return []string{""}
    }

    numRows := fb.Height / 2
    if fb.Height%2 != 0 {
        numRows++
    }

    grid := makeGrid(fb, numRows)

    overlayText(grid, tree)

    return renderANSI(grid, viewportWidth)
}

func makeGrid(fb *framebuffer.Framebuffer, numRows int) [][]Cell {
    grid := make([][]Cell, numRows)
    for row := 0; row < numRows; row++ {
        grid[row] = make([]Cell, fb.Width)
        for col := 0; col < fb.Width; col++ {
            top := fb.Get(col, row*2)
            bottom := fb.Get(col, row*2+1)
            char, fg, bg := chooseChar(top, bottom)
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

            esc := ansiColor(cell.FG, cell.BG, lastFG, lastBG)
            if esc != "" {
                b.WriteString(esc)
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
            b.WriteRune(ch)
            visibleCount++
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
        for _, cell := range row {
            if cell.Char == 0 {
                b.WriteRune(' ')
            } else {
                b.WriteRune(cell.Char)
            }
        }
        s := b.String()
        // Clip by rune count, not byte count (multi-byte chars like █ are 3 bytes)
        runes := []rune(s)
        if len(runes) > maxWidth {
            runes = runes[:maxWidth]
        }
        lines = append(lines, string(runes))
    }
    return lines
}

func ansiColor(fg, bg, lastFG, lastBG css.Color) string {
    // Only emit if color actually changed
    if colorsEqual(fg, lastFG) && colorsEqual(bg, lastBG) {
        return ""
    }
    // Always reset first, then set both fg and bg together
    var parts []string
    if fg.IsSet {
        parts = append(parts, "38;2;"+itoa(int(fg.R))+";"+itoa(int(fg.G))+";"+itoa(int(fg.B)))
    }
    if bg.IsSet {
        parts = append(parts, "48;2;"+itoa(int(bg.R))+";"+itoa(int(bg.G))+";"+itoa(int(bg.B)))
    }
    if len(parts) == 0 {
        return ""
    }
    return "\033[0m\033[" + strings.Join(parts, ";") + "m"
}

func itoa(n int) string {
    if n == 0 {
        return "0"
    }
    var buf [20]byte
    i := len(buf)
    for n > 0 {
        i--
        buf[i] = byte('0' + n%10)
        n /= 10
    }
    return string(buf[i:])
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

        row := int(tl.Y) / 2
        if row < 0 || row >= len(grid) {
            continue
        }

        col := int(tl.X)
        if col < 0 {
            col = 0
        }

        for j, ch := range tl.Text {
            c := col + j
            if c >= len(grid[row]) {
                break
            }
            // Keep original background from grid so text doesn't appear on black bg
            origBG := grid[row][c].BG
            grid[row][c] = Cell{Char: rune(ch), FG: color, BG: origBG}
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
    startCol := int(tree.Border.X)
    startRow := int(tree.Border.Y / 2)
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