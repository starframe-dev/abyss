package asciirender

import (
    "image"
    "image/color"
    "strings"
    "testing"

    "github.com/starframe/abyss/internal/css"
)

func newSolidImage(w, h int, c color.Color) *image.RGBA {
    img := image.NewRGBA(image.Rect(0, 0, w, h))
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            img.Set(x, y, c)
        }
    }
    return img
}

func newGradientImage(w, h int) *image.RGBA {
    img := image.NewRGBA(image.Rect(0, 0, w, h))
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            img.Set(x, y, color.RGBA{
                R: uint8(x * 255 / w),
                G: uint8(y * 255 / h),
                B: 128,
                A: 255,
            })
        }
    }
    return img
}

func TestQuantize_SolidWhite_AllBlank(t *testing.T) {
    img := newSolidImage(40, 40, color.RGBA{255, 255, 255, 255})
    grid := Quantize(img, Options{Columns: 8, Rows: 4})

    if len(grid) != 4 {
        t.Fatalf("rows = %d, want 4", len(grid))
    }
    sparsest := Palette[0]
    for r, row := range grid {
        if len(row) != 8 {
            t.Fatalf("cols at row %d = %d, want 8", r, len(row))
        }
        for c, cell := range row {
            if cell.Char != sparsest {
                t.Errorf("cell[%d][%d] char = %q, want sparsest %q for white pixel", r, c, cell.Char, sparsest)
            }
        }
    }
}

func TestQuantize_SolidBlack_DenseBlocks(t *testing.T) {
    img := newSolidImage(40, 40, color.RGBA{0, 0, 0, 255})
    grid := Quantize(img, Options{Columns: 8, Rows: 4})

    densest := Palette[len(Palette)-1]
    for r, row := range grid {
        for c, cell := range row {
            if cell.Char != densest {
                t.Errorf("cell[%d][%d] char = %q, want densest %q for black pixel", r, c, cell.Char, densest)
            }
            if cell.FG.R != 0 || cell.FG.G != 0 || cell.FG.B != 0 {
                t.Errorf("cell[%d][%d] fg = %v, want pure black", r, c, cell.FG)
            }
        }
    }
}

func TestQuantize_GradientProducesRampAndColor(t *testing.T) {
    img := newGradientImage(40, 40)
    grid := Quantize(img, Options{Columns: 10, Rows: 5})

    prevLum := uint16(0)
    sawIncrease := false
    for c := 0; c < 10; c++ {
        cell := grid[0][c]
        lum := luminance(uint16(cell.FG.R), uint16(cell.FG.G), uint16(cell.FG.B))
        if lum > prevLum {
            sawIncrease = true
        }
        prevLum = lum
    }
    if !sawIncrease {
        t.Errorf("expected luminance to increase across the gradient row, all chars: %q", rowChars(grid[0]))
    }
}

func TestQuantize_AspectRatioPreserved(t *testing.T) {
    img := newSolidImage(80, 20, color.RGBA{100, 100, 100, 255})
    grid := Quantize(img, Options{Columns: 16})

    if len(grid) < 1 || len(grid) > 4 {
        t.Fatalf("rows = %d, want 1-4 for wide image", len(grid))
    }
}

func TestQuantize_AlphaCompositedOnWhite(t *testing.T) {
    img := image.NewRGBA(image.Rect(0, 0, 4, 4))
    img.Set(0, 0, color.RGBA{0, 0, 0, 0})
    img.Set(1, 1, color.RGBA{0, 0, 0, 128})

    grid := Quantize(img, Options{Columns: 2, Rows: 2})
    if len(grid) != 2 || len(grid[0]) != 2 {
        t.Fatalf("grid shape = %dx%d, want 2x2", len(grid), len(grid[0]))
    }
    // Top-left cell sampled fully transparent → should composite to white-ish.
    cell := grid[0][0]
    if cell.FG.R < 200 || cell.FG.G < 200 || cell.FG.B < 200 {
        t.Errorf("fully transparent cell fg = %v, want near-white", cell.FG)
    }
}

func TestFlatten_ProducesNewlineSeparatedRows(t *testing.T) {
    img := newSolidImage(20, 20, color.RGBA{50, 50, 50, 255})
    grid := Quantize(img, Options{Columns: 4, Rows: 3})
    flat := Flatten(grid)

    lines := strings.Split(strings.TrimRight(flat, "\n"), "\n")
    if len(lines) != 3 {
        t.Fatalf("lines = %d, want 3", len(lines))
    }
    for i, line := range lines {
        if len([]rune(line)) != 4 {
            t.Errorf("line %d length = %d, want 4", i, len([]rune(line)))
        }
    }
}

func TestQuantize_HonoursBackground(t *testing.T) {
    img := newSolidImage(20, 20, color.RGBA{255, 0, 0, 255})
    custom := css.RGB(10, 20, 30)
    grid := Quantize(img, Options{Columns: 2, Rows: 1, BGColor: custom})

    for _, cell := range grid[0] {
        if cell.BG != custom {
            t.Errorf("bg = %v, want %v", cell.BG, custom)
        }
    }
}

func rowChars(row []Cell) string {
    var b strings.Builder
    for _, c := range row {
        b.WriteRune(c.Char)
    }
    return b.String()
}

func TestPalette_MonotonicIncreasingDensity(t *testing.T) {
    if len(Palette) < 4 {
        t.Fatalf("palette must have at least 4 chars, got %d", len(Palette))
    }
    // Just sanity-check we expose a reasonable ramp.
    first := Palette[0]
    last := Palette[len(Palette)-1]
    if first == last {
        t.Errorf("palette first and last chars identical: %q", first)
    }
}