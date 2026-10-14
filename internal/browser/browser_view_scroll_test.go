package browser

import (
    "image"
    "image/color"
    "strings"
    "testing"

    "github.com/starframe/abyss/internal/ascii"
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/layout"
    "github.com/starframe/abyss/internal/render"
)

// imageModel builds a Model in view mode whose ANSI lines look exactly like
// real truecolor image output (many runes per visible column).
func imageModel(t *testing.T, width int) Model {
    t.Helper()
    const src = "https://example.test/pic.png"
    prev := render.ImageProvider
    render.ImageProvider = func(s string) image.Image {
        if s != src {
            return nil
        }
        img := image.NewRGBA(image.Rect(0, 0, 32, 8))
        for y := 0; y < 8; y++ {
            for x := 0; x < 32; x++ {
                img.Set(x, y, color.RGBA{R: 220, G: 30, B: 30, A: 255})
            }
        }
        return img
    }
    defer func() { render.ImageProvider = prev }()

    doc := dom.Parse(`<html><body><h1>TITLE</h1><img src="` + src + `" alt="pic"/></body></html>`)
    styles := css.Compute(doc)
    box := layout.Build(doc, styles, float64(width))
    fb := render.Render(box, width)
    lines := ascii.RenderANSI(fb, box, width)

    m := New()
    m.Width = width
    m.Height = 40
    m.Mode = ModeView
    m.Tabs = []Tab{{Mode: ModeView, ANSILines: lines, ZoomFactor: 1.0}}
    m.ActiveTab = 0
    m.Viewport.Width = width
    m.Viewport.Height = 30
    return m
}

func TestView_HorizontalScrollKeepsImageColumnsRoundTrip(t *testing.T) {
    const width = 60
    m := imageModel(t, width)

    base := m.View()
    if !strings.Contains(base, "TITLE") {
        t.Fatalf("baseline view lost the heading:\n%s", base)
    }
    baseCols := imageColumns(t, base)
    if baseCols < 10 {
        t.Fatalf("baseline view has no image columns (%d); fixture broken", baseCols)
    }

    // Scroll right once: the heading sits in columns 0-5 and must leave the
    // window, while the picture is clipped but still present.
    m.active().OffsetX += 10
    scrolled := m.View()
    if strings.Contains(stripSGRForTest(scrolled), "TITLE") {
        t.Errorf("heading should scroll out of view at offset 10")
    }
    midCols := imageColumns(t, scrolled)
    if midCols == 0 {
        t.Errorf("scrolling right erased the picture entirely")
    }
    // Scroll back: the stored lines must not have been destroyed in place.
    m.active().OffsetX = 0
    back := m.View()
    if got := imageColumns(t, back); got != baseCols {
        t.Errorf("round trip lost image columns: %d -> %d -> %d", baseCols, midCols, got)
    }
    if back != base {
        t.Errorf("view at offset 0 differs after a scroll round trip")
    }
    if strings.Contains(back, "[K") {
        t.Errorf("erase-to-EOL artifact leaked into view after round trip")
    }
}

func stripSGRForTest(view string) string {
    var b strings.Builder
    for _, line := range strings.Split(view, "\n") {
        for _, seg := range splitANSI(line) {
            b.WriteString(seg.text)
        }
    }
    return b.String()
}

// imageColumns counts printable columns that carry a truecolor foreground,
// i.e. the pixels of a rendered picture.
func imageColumns(t *testing.T, view string) int {
    t.Helper()
    n := 0
    for _, line := range strings.Split(view, "\n") {
        for _, seg := range splitANSI(line) {
            if len(seg.style) == 0 {
                continue
            }
            if strings.Contains(strings.Join(seg.style, ""), "220;30;30") {
                n += len([]rune(seg.text))
            }
        }
    }
    return n
}
