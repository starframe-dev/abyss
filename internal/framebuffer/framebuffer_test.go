package framebuffer

import (
	"math"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
)

func color(r, g, b uint8) css.Color { return css.Color{R: r, G: g, B: b, IsSet: true} }

func TestNew(t *testing.T) {
	for _, d := range [][2]int{{1, 1}, {MaxWidth + 1, 1}, {1, MaxHeight + 1}, {MaxPixels, MaxPixels}, {0, 1}, {1, 0}, {-1, 2}} {
		fb := New(d[0], d[1])
		if d[0] == 1 && d[1] == 1 {
			if fb.Width != 1 || fb.Height != 1 || len(fb.Pixels) != 1 {
				t.Fatalf("valid dimensions: %+v", fb)
			}
		} else if fb.Width != 0 || fb.Height != 0 || len(fb.Pixels) != 0 {
			t.Errorf("invalid dimensions %v accepted", d)
		}
	}
	if fb := New(MaxWidth, MaxHeight); fb.Width != 0 {
		t.Fatal("pixel limit should reject dimensions")
	}
}

func TestSetGetAndClear(t *testing.T) {
	c := color(1, 2, 3)
	var nilFB *Framebuffer
	nilFB.Set(0, 0, c)
	if nilFB.Get(0, 0) != css.Transparent() {
		t.Fatal("nil Get should be transparent")
	}
	fb := New(2, 2)
	fb.Set(1, 1, c)
	if fb.Get(1, 1) != c {
		t.Fatal("pixel not set")
	}
	for _, p := range [][2]int{{-1, 0}, {2, 0}, {0, -1}, {0, 2}} {
		fb.Set(p[0], p[1], c)
		if fb.Get(p[0], p[1]) != css.Transparent() {
			t.Errorf("invalid coordinate %v", p)
		}
	}
	fb.Clear(color(4, 5, 6))
	for _, p := range fb.Pixels {
		if p != color(4, 5, 6) {
			t.Fatal("clear failed")
		}
	}
	nilFB.Clear(c)
}

func TestFillRectClippingAndInvalid(t *testing.T) {
	fb := New(4, 4)
	c := color(9, 8, 7)
	fb.FillRect(-1.2, -1.2, 3, 3, c)
	if fb.Get(0, 0) != c || fb.Get(1, 1) != c || fb.Get(2, 2) != css.Transparent() {
		t.Fatal("clipping/fractional fill incorrect")
	}
	for _, v := range [][4]float64{{0, 0, 0, 1}, {0, 0, math.NaN(), 1}, {math.Inf(1), 0, 1, 1}, {math.MaxFloat64, 0, math.MaxFloat64, 1}} {
		fb.FillRect(v[0], v[1], v[2], v[3], c)
	}
	var nilFB *Framebuffer
	nilFB.FillRect(0, 0, 1, 1, c)
}

func TestDrawBorder(t *testing.T) {
	fb := New(5, 5)
	top, right, bottom, left := color(1, 0, 0), color(0, 1, 0), color(0, 0, 1), color(1, 1, 1)
	fb.DrawBorder(1, 1, 3, 3, 1, 1, 1, 1, top, right, bottom, left)
	if fb.Get(2, 1) != top || fb.Get(3, 2) != right || fb.Get(2, 3) != bottom || fb.Get(1, 2) != left {
		t.Fatal("border side missing")
	}
	fb.DrawBorder(0, 0, 1, 1, 0, 0, 0, 0, top, right, bottom, left)
	var nilFB *Framebuffer
	nilFB.DrawBorder(0, 0, 1, 1, 1, 1, 1, 1, top, right, bottom, left)
}

func TestDrawRoundedRect(t *testing.T) {
	fb := New(8, 8)
	c := color(255, 255, 255)
	fb.DrawRoundedRect(1, 1, 6, 6, 3, 3, 3, 3, c)
	if fb.Get(3, 3) != c || fb.Get(1, 1) != css.Transparent() {
		t.Fatal("rounded rectangle fill or clipping incorrect")
	}
	fb.DrawRoundedRect(0, 0, 2, 2, 0, 0, 0, 0, c)
	var nilFB *Framebuffer
	nilFB.DrawRoundedRect(0, 0, 1, 1, 1, 1, 1, 1, c)
}
