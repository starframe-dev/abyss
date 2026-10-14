// Package render draws a layout tree into a pixel framebuffer.
package render

import (
	"image"
	"sort"

	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
)

// ImageProvider returns a decoded image for the given src URL, or nil if
// the image is not available. Render uses it to convert <img> boxes into
// ASCII art with original colors. Callers wire this up at startup.
var ImageProvider func(src string) image.Image

// Render draws a layout tree into a framebuffer.
func Render(tree *layout.LayoutBox, viewportWidth int) *framebuffer.Framebuffer {
	if tree == nil {
		return framebuffer.New(1, 1)
	}

	// Collect all boxes with their z-index
	var boxes []*layout.LayoutBox
	collectBoxes(tree, &boxes)

	// Calculate offset to ensure nothing is negative
	minX, minY := calculateOffset(boxes)

	// Shift tree coordinates
	shiftTree(tree, -minX, -minY)

	// Calculate required framebuffer height
	height := calculateHeight(tree)
	fb := framebuffer.New(viewportWidth, height)

	// Use body background color, default to white
	bgColor := css.RGB(255, 255, 255)
	if tree != nil && tree.Style.BackgroundColor.IsSet {
		bgColor = tree.Style.BackgroundColor
	}
	fb.Clear(bgColor)

	// Sort by z-index (lower first). Preserve DOM preorder for equal z-indexes
	// so descendants paint over their ancestors according to normal paint order.
	sort.SliceStable(boxes, func(i, j int) bool {
		return boxes[i].Style.ZIndex < boxes[j].Style.ZIndex
	})

	// Draw each box
	for _, box := range boxes {
		drawBox(fb, box)
	}

	return fb
}

func calculateOffset(boxes []*layout.LayoutBox) (float64, float64) {
	minX, minY := 0.0, 0.0
	for _, box := range boxes {
		if box.Margin.X < minX {
			minX = box.Margin.X
		}
		if box.Margin.Y < minY {
			minY = box.Margin.Y
		}
	}
	return minX, minY
}

func calculateHeight(box *layout.LayoutBox) int {
	maxY := int(box.Margin.Y + box.Margin.Height)
	for _, child := range box.Children {
		ch := calculateHeight(child)
		if ch > maxY {
			maxY = ch
		}
	}
	return maxY
}

func shiftTree(box *layout.LayoutBox, offX, offY float64) {
	box.Content.X += offX
	box.Content.Y += offY
	box.Padding.X += offX
	box.Padding.Y += offY
	box.Border.X += offX
	box.Border.Y += offY
	box.Margin.X += offX
	box.Margin.Y += offY
	for i := range box.TextLines {
		box.TextLines[i].X += offX
		box.TextLines[i].Y += offY
	}
	for _, child := range box.Children {
		shiftTree(child, offX, offY)
	}
}

func collectBoxes(box *layout.LayoutBox, boxes *[]*layout.LayoutBox) {
	*boxes = append(*boxes, box)
	for _, child := range box.Children {
		collectBoxes(child, boxes)
	}
}

func drawBox(fb *framebuffer.Framebuffer, box *layout.LayoutBox) {
	style := box.Style

	// Background
	if style.BackgroundColor.IsSet {
		fb.DrawRoundedRect(
			box.Border.X, box.Border.Y,
			box.Border.Width, box.Border.Height,
			style.BorderRadiusTopLeft.Resolve(box.Border.Width),
			style.BorderRadiusTopRight.Resolve(box.Border.Width),
			style.BorderRadiusBottomRight.Resolve(box.Border.Width),
			style.BorderRadiusBottomLeft.Resolve(box.Border.Width),
			style.BackgroundColor,
		)
	} else if box.Node != nil && box.Node.TagName == "img" {
		if cells := quantizeImageBox(box); cells != nil {
			drawAsciiCells(fb, box, cells)
		} else {
			// Terminal fallback for raster images: reserve a visible neutral panel.
			placeholder := css.RGB(226, 232, 240)
			fb.DrawRoundedRect(
				box.Border.X, box.Border.Y,
				box.Border.Width, box.Border.Height,
				4, 4, 4, 4,
				placeholder,
			)
		}
	}

	// Border
	if style.BorderTop.Style != css.BorderNone || style.BorderRight.Style != css.BorderNone ||
		style.BorderBottom.Style != css.BorderNone || style.BorderLeft.Style != css.BorderNone {
		fb.DrawBorder(
			box.Border.X, box.Border.Y, box.Border.Width, box.Border.Height,
			style.BorderTop.Width, style.BorderRight.Width,
			style.BorderBottom.Width, style.BorderLeft.Width,
			style.BorderTop.Color, style.BorderRight.Color,
			style.BorderBottom.Color, style.BorderLeft.Color,
		)
	}
}

// quantizeImageBox converts the image referenced by an <img> layout box into
// an ASCII cell grid sized to the box content area. The result is cached on
// the box so ascii.Render can use the same characters.
func quantizeImageBox(box *layout.LayoutBox) [][]asciirender.Cell {
	if box.Node == nil {
		return nil
	}
	if box.AsciiCells != nil {
		return box.AsciiCells
	}
	if ImageProvider == nil {
		return nil
	}
	src := box.Node.GetAttr("src")
	if src == "" {
		return nil
	}
	img := ImageProvider(src)
	if img == nil {
		return nil
	}

	widthCells := int(box.Content.Width)
	if widthCells < 1 {
		widthCells = 1
	}
	rows := len(box.TextLines)
	heightCells := rows / 2
	if heightCells < 1 {
		heightCells = 1
	}

	bg := css.RGB(226, 232, 240)
	if box.Style.BackgroundColor.IsSet {
		bg = box.Style.BackgroundColor
	}
	grid := asciirender.Quantize(img, asciirender.Options{
		Columns: widthCells,
		Rows:    heightCells,
		BGColor: bg,
	})
	box.AsciiCells = grid
	// layoutImage seeds TextLines with the "[image] alt" placeholder. Once real
	// ASCII cells exist that placeholder must go, otherwise the text overlay is
	// painted on top of the picture and erases its left columns.
	box.TextLines = nil
	return grid
}

// drawAsciiCells paints the pre-quantized ASCII grid into the framebuffer.
// Each cell occupies one column and two pixel rows; both rows receive the
// cell's foreground color so the ANSI renderer can pick the right hue from
// either half when collapsing to terminal rows.
func drawAsciiCells(fb *framebuffer.Framebuffer, box *layout.LayoutBox, grid [][]asciirender.Cell) {
	if len(grid) == 0 {
		return
	}
	startX := int(box.Border.X)
	startY := int(box.Border.Y)
	cellW := box.Border.Width / float64(len(grid[0]))
	if cellW < 1 {
		cellW = 1
	}
	cellH := box.Border.Height / float64(len(grid))
	if cellH < 1 {
		cellH = 1
	}
	for r, row := range grid {
		for c, cell := range row {
			x0 := startX + int(float64(c)*cellW)
			y0 := startY + int(float64(r)*cellH)
			w := int(cellW)
			if x0+w > fb.Width {
				w = fb.Width - x0
			}
			if w < 1 {
				continue
			}
			topY := y0
			botY := y0 + int(cellH/2)
			for xx := x0; xx < x0+w && xx < fb.Width; xx++ {
				if cell.FG.IsSet {
					fb.Set(xx, topY, cell.FG)
				}
				if botY < fb.Height && cell.FG.IsSet {
					fb.Set(xx, botY, cell.FG)
				}
			}
		}
	}
}
