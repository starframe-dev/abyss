// Package framebuffer provides a pixel buffer for rendering layout boxes.
package framebuffer

import (
    "math"

    "github.com/starframe/abyss/internal/css"
)

// Framebuffer is an RGB pixel buffer.
type Framebuffer struct {
    Width  int
    Height int
    Pixels []css.Color
}

// New creates a new framebuffer with the given dimensions.
func New(width, height int) *Framebuffer {
    return &Framebuffer{
        Width:  width,
        Height: height,
        Pixels: make([]css.Color, width*height),
    }
}

// Set sets a pixel at (x, y).
func (fb *Framebuffer) Set(x, y int, c css.Color) {
    if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
        return
    }
    fb.Pixels[y*fb.Width+x] = c
}

// Get gets a pixel at (x, y).
func (fb *Framebuffer) Get(x, y int) css.Color {
    if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
        return css.Transparent()
    }
    return fb.Pixels[y*fb.Width+x]
}

// Clear fills the entire framebuffer with a color.
func (fb *Framebuffer) Clear(c css.Color) {
    for i := range fb.Pixels {
        fb.Pixels[i] = c
    }
}

// FillRect fills a rectangle with a color.
func (fb *Framebuffer) FillRect(x, y, w, h float64, c css.Color) {
    x0 := int(math.Floor(x))
    y0 := int(math.Floor(y))
    x1 := int(math.Ceil(x + w))
    y1 := int(math.Ceil(y + h))
    for py := y0; py < y1; py++ {
        for px := x0; px < x1; px++ {
            fb.Set(px, py, c)
        }
    }
}

// DrawBorder draws borders for a layout box.
func (fb *Framebuffer) DrawBorder(x, y, w, h, topW, rightW, bottomW, leftW float64, topC, rightC, bottomC, leftC css.Color) {
    // Top border
    if topW > 0 {
        fb.FillRect(x, y, w, topW, topC)
    }
    // Right border
    if rightW > 0 {
        fb.FillRect(x+w-rightW, y, rightW, h, rightC)
    }
    // Bottom border
    if bottomW > 0 {
        fb.FillRect(x, y+h-bottomW, w, bottomW, bottomC)
    }
    // Left border
    if leftW > 0 {
        fb.FillRect(x, y, leftW, h, leftC)
    }
}

// DrawRoundedRect draws a rectangle with rounded corners.
func (fb *Framebuffer) DrawRoundedRect(x, y, w, h float64, rtl, rtr, rbr, rbl float64, c css.Color) {
    fb.FillRect(x, y, w, h, c)

    // Clip corners
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

// clipCorner clips a corner of the framebuffer.
// quadrant: 0=top-left, 1=top-right, 2=bottom-right, 3=bottom-left
func (fb *Framebuffer) clipCorner(cx, cy, r float64, quadrant int) {
    x0 := int(math.Floor(cx))
    y0 := int(math.Floor(cy))
    x1 := int(math.Ceil(cx + r))
    y1 := int(math.Ceil(cy + r))

    for py := y0; py < y1; py++ {
        for px := x0; px < x1; px++ {
            dx := float64(px) - cx
            dy := float64(py) - cy
            if dx*dx+dy*dy > r*r {
                fb.Set(px, py, css.Transparent())
            }
        }
    }
}
