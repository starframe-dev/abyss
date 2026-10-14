// Package asciirender converts raster images into ASCII art suitable for
// terminal display. It uses a density-based character ramp and emits 24-bit
// ANSI color codes per cell so the result preserves the original colors.
package asciirender

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/starframe-dev/abyss/internal/css"
)

// Cell describes one rendered terminal cell.
type Cell struct {
	Char      rune
	FG        css.Color
	BG        css.Color
	IsHalfRow bool // true if this is a "top half" pixel paired with the next row's "bottom half"
}

// Decode parses the given bytes as PNG, JPEG, GIF or any other format
// registered with the standard image package.
const (
	MaxImageWidth    = 8192
	MaxImageHeight   = 8192
	MaxImagePixels   = 8 * 1024 * 1024
	MaxOutputColumns = 4096
	MaxOutputCells   = 1 << 18
)

// Decode parses image bytes after validating dimensions before pixel allocation.
func Decode(data []byte) (image.Image, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if !imageDimensionsAllowed(config.Width, config.Height) {
		return nil, fmt.Errorf("image dimensions %dx%d exceed supported limits", config.Width, config.Height)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// Palette returns the density ramp used to convert luminance into runes.
// Lower density → light character, higher density → solid block.
var Palette = []rune(" .,:;i1tfLCG08@#")

// Options controls how an image is quantized into ASCII cells.
type Options struct {
	Columns int       // number of terminal columns (each cell = 2 px height)
	Rows    int       // number of terminal rows; 0 = derive from aspect ratio
	BGColor css.Color // background underlay when cells use a foreground-only ramp
}

func imageDimensionsAllowed(width, height int) bool {
	return width > 0 && height > 0 && width <= MaxImageWidth && height <= MaxImageHeight && width <= MaxImagePixels/height
}

// Quantize converts the image into a 2D grid of ASCII cells using the
// configured options. Each terminal cell corresponds to one column and
// roughly two rows of source pixels (for 1:2 aspect ratio of terminal cells).
func Quantize(img image.Image, opts Options) [][]Cell {
	if img == nil {
		return nil
	}
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return nil
	}

	cols := opts.Columns
	if cols <= 0 {
		cols = 60
	}
	cols = min(cols, MaxOutputColumns)
	maxRows := max(1, MaxOutputCells/cols)
	// Terminal cell aspect ratio: ~2:1 (height:width), without overflowing int.
	rows := opts.Rows
	if rows <= 0 {
		derived := float64(cols) * float64(srcH) / float64(srcW) / 2
		if derived >= float64(maxRows) {
			rows = maxRows
		} else {
			rows = int(derived)
		}
		if rows < 1 {
			rows = 1
		}
	} else if rows > maxRows {
		rows = maxRows
	}

	bg := opts.BGColor
	if !bg.IsSet {
		bg = css.RGB(226, 232, 240)
	}

	grid := make([][]Cell, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]Cell, cols)
		for c := 0; c < cols; c++ {
			r16, g16, b16, lum := sampleCell(img, srcBounds, c, r, cols, rows)
			ch := pickRune(lum)
			grid[r][c] = Cell{
				Char: ch,
				FG:   css.RGB(uint8(r16), uint8(g16), uint8(b16)),
				BG:   bg,
			}
		}
	}
	return grid
}

// sample aggregates the source pixels for a single terminal cell and returns
// the average color along with luminance.
func sampleCell(img image.Image, bounds image.Rectangle, col, row, cols, rows int) (r, g, b uint16, lum uint16) {
	x0 := bounds.Min.X + col*bounds.Dx()/cols
	y0 := bounds.Min.Y + row*bounds.Dy()/rows
	x1 := bounds.Min.X + (col+1)*bounds.Dx()/cols
	y1 := bounds.Min.Y + (row+1)*bounds.Dy()/rows

	var sumR, sumG, sumB uint64
	var count uint64
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			rr, gg, bb, aa := rgba(img.At(x, y))
			// Simple alpha compositing over white.
			alpha := uint32(aa)
			inv := uint32(255 - aa)
			rr = uint16((uint32(rr)*alpha + 255*inv) / 255)
			gg = uint16((uint32(gg)*alpha + 255*inv) / 255)
			bb = uint16((uint32(bb)*alpha + 255*inv) / 255)

			sumR += uint64(rr)
			sumG += uint64(gg)
			sumB += uint64(bb)
			count++
		}
	}
	if count == 0 {
		return 0, 0, 0, 0
	}
	r = uint16(sumR / count)
	g = uint16(sumG / count)
	b = uint16(sumB / count)
	lum = luminance(r, g, b)
	return
}

// pickRune chooses a character based on the perceived luminance.
// Dark pixels (low luminance) map to dense glyphs, light pixels to thin glyphs.
func pickRune(lum uint16) rune {
	if lum > 255 {
		lum = 255
	}
	inv := 255 - int(lum)
	idx := inv * (len(Palette) - 1) / 255
	if idx < 0 {
		idx = 0
	}
	if idx >= len(Palette) {
		idx = len(Palette) - 1
	}
	return Palette[idx]
}

// luminance computes ITU-R BT.601 luma from RGB channels.
func luminance(r, g, b uint16) uint16 {
	// Y = 0.299R + 0.587G + 0.114B
	y := 299*uint32(r) + 587*uint32(g) + 114*uint32(b)
	return uint16(y / 1000)
}

func rgba(c color.Color) (r, g, b, a uint16) {
	rr, gg, bb, aa := c.RGBA()
	return uint16(rr >> 8), uint16(gg >> 8), uint16(bb >> 8), uint16(aa >> 8)
}

// Flatten converts a 2D grid into a single string with a trailing newline
// per row, suitable for terminal preview without ANSI codes.
func Flatten(grid [][]Cell) string {
	var b bytes.Buffer
	for _, row := range grid {
		for _, cell := range row {
			if cell.Char == 0 {
				b.WriteRune(' ')
			} else {
				b.WriteRune(cell.Char)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
