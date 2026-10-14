package browser

import (
    "fmt"
    "image"
    "os"
    "strings"

    "github.com/starframe/abyss/internal/ascii"
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/fetcher"
    "github.com/starframe/abyss/internal/layout"
    "github.com/starframe/abyss/internal/render"
)

// HeadlessDump loads pageURL, runs the same rendering pipeline as loadPage,
// and writes the resulting ANSI art to outputPath. It does not start the TUI;
// useful for deterministic E2E checks via PTY-less test runners.
func HeadlessDump(pageURL string, width int, outputPath string) {
    imageCacheMu.Lock()
    imageCache = make(map[string]image.Image)
    imageCacheMu.Unlock()
    render.ImageProvider = cachedImage

    html, err := fetcher.Fetch(pageURL)
    if err != nil {
        fmt.Fprintln(os.Stderr, "headless dump: fetch:", err)
        os.Exit(1)
    }
    doc := dom.Parse(html)
    fetchImages(doc, pageURL)

    styles := css.ComputeWithExternal(doc, fetchExternalCSS(doc, pageURL))
    box := layout.Build(doc, styles, float64(width))
    fb := render.Render(box, width)
    ansiLines := ascii.RenderANSI(fb, box, width)

    dump := strings.Join(ansiLines, "\n")
    if err := os.WriteFile(outputPath, []byte(dump), 0o644); err != nil {
        fmt.Fprintln(os.Stderr, "headless dump: write:", err)
        os.Exit(1)
    }
    fmt.Fprintf(os.Stderr, "headless dump: %d bytes -> %s\n", len(dump), outputPath)
}