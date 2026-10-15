// Package framebuffer provides a pixel buffer for rendering layout boxes.
package framebuffer

import (
	"math"

	"github.com/starframe-dev/abyss/internal/css"
)

// Limits bound the memory requested by framebuffer dimensions.
const (
	MaxWidth  = 4096
	MaxHeight = 4096
	MaxPixels = 16 * 1024 * 1024
)

// Framebuffer is an RGB pixel buffer.
type Framebuffer struct {
	Width  int
	Height int
	Pixels []css.Color
}

// New creates a framebuffer. Invalid or excessive dimensions return an empty buffer.
func New(width, height int) *Framebuffer {
	fb := &Framebuffer{}
	if width <= 0 || height <= 0 || width > MaxWidth || height > MaxHeight || width > MaxPixels/height {
		return fb
	}
	fb.Width, fb.Height = width, height
	fb.Pixels = make([]css.Color, width*height)
	return fb
}

// Set sets a pixel at (x, y).
func (fb *Framebuffer) Set(x, y int, c css.Color) {
	if fb == nil || x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
		return
	}
	fb.Pixels[y*fb.Width+x] = c
}

// Get gets a pixel at (x, y).
func (fb *Framebuffer) Get(x, y int) css.Color {
	if fb == nil || x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
		return css.Transparent()
	}
	return fb.Pixels[y*fb.Width+x]
}

// Clear fills the entire framebuffer with a color.
func (fb *Framebuffer) Clear(c css.Color) {
	if fb == nil {
		return
	}
	for i := range fb.Pixels {
		fb.Pixels[i] = c
	}
}

func clippedRange(start, extent float64, maxSize int) (int, int, bool) {
	if math.IsNaN(start) || math.IsNaN(extent) || math.IsInf(start, 0) || math.IsInf(extent, 0) || extent <= 0 {
		return 0, 0, false
	}
	end := start + extent
	if math.IsInf(end, 0) {
		return 0, 0, false
	}
	lo := max(0, int(math.Floor(max(-float64(maxSize), start))))
	hi := min(maxSize, int(math.Ceil(min(float64(maxSize), end))))
	return lo, hi, hi > lo
}

// FillRect fills a rectangle with a color.
func (fb *Framebuffer) FillRect(x, y, w, h float64, c css.Color) {
	if fb == nil {
		return
	}
	x0, x1, okX := clippedRange(x, w, fb.Width)
	y0, y1, okY := clippedRange(y, h, fb.Height)
	if !okX || !okY {
		return
	}
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			fb.Set(px, py, c)
		}
	}
}

// DrawBorder draws borders for a layout box.
func (fb *Framebuffer) DrawBorder(x, y, w, h, topW, rightW, bottomW, leftW float64, topC, rightC, bottomC, leftC css.Color) {
	if topW > 0 {
		fb.FillRect(x, y, w, topW, topC)
	}
	if rightW > 0 {
		fb.FillRect(x+w-rightW, y, rightW, h, rightC)
	}
	if bottomW > 0 {
		fb.FillRect(x, y+h-bottomW, w, bottomW, bottomC)
	}
	if leftW > 0 {
		fb.FillRect(x, y, leftW, h, leftC)
	}
}

// DrawRoundedRect draws a rectangle with rounded corners.
func (fb *Framebuffer) DrawRoundedRect(x, y, w, h, rtl, rtr, rbr, rbl float64, c css.Color) {
	fb.FillRect(x, y, w, h, c)
	if rtl > 0 {
		fb.clipCorner(x, y, rtl, 0)
	}
	if rtr > 0 {
		fb.clipCorner(x+w-rtr, y, rtr, 1)
	}
	if rbr > 0 {
		fb.clipCorner(x+w-rbr, y+h-rbr, rbr, 2)
	}
	if rbl > 0 {
		fb.clipCorner(x, y+h-rbl, rbl, 3)
	}
}

// clipCorner clears the pixels outside a rounded corner. (cx, cy) is the top-left
// of the r-sized square that holds the corner; quadrant 0..3 is top-left,
// top-right, bottom-right, bottom-left. The arc centre lies inside the box.
func (fb *Framebuffer) clipCorner(cx, cy, r float64, quadrant int) {
	if fb == nil || math.IsNaN(cx) || math.IsNaN(cy) || math.IsNaN(r) || math.IsInf(cx, 0) || math.IsInf(cy, 0) || math.IsInf(r, 0) || r <= 0 {
		return
	}
	centerX, centerY := cx+r, cy+r
	switch quadrant {
	case 1:
		centerX, centerY = cx, cy+r
	case 2:
		centerX, centerY = cx, cy
	case 3:
		centerX, centerY = cx+r, cy
	}
	x0, x1, okX := clippedRange(cx, r, fb.Width)
	y0, y1, okY := clippedRange(cy, r, fb.Height)
	if !okX || !okY {
		return
	}
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			dx := float64(px) + 0.5 - centerX
			dy := float64(py) + 0.5 - centerY
			if dx*dx+dy*dy > r*r {
				fb.Set(px, py, css.Transparent())
			}
		}
	}
}
