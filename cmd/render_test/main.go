// render_test renders a URL to ANSI art, suitable for inspection of the
// ASCII image renderer. Used only for development.
package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"os"
	"strings"
	"sync"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/fetcher"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: render_test <url>")
		os.Exit(2)
	}
	url := os.Args[1]
	width := 120
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &width)
	}

	imageCacheMu.Lock()
	imageCache = make(map[string]image.Image)
	imageCacheMu.Unlock()
	render.ImageProvider = func(src string) image.Image {
		imageCacheMu.RLock()
		defer imageCacheMu.RUnlock()
		if img, ok := imageCache[src]; ok {
			return img
		}
		return nil
	}

	html, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch error:", err)
		os.Exit(1)
	}
	doc := dom.Parse(html)
	fetchImages(doc, url)

	styles := css.Compute(doc)
	box := layout.Build(doc, styles, float64(width))
	fb := render.Render(box, width)
	ansi := ascii.RenderANSI(fb, box, width)

	var b strings.Builder
	for _, line := range ansi {
		b.WriteString(line)
		b.WriteString("\n")
	}
	out := b.String()
	_ = os.WriteFile("/tmp/img_fixture_ansi.txt", []byte(out), 0o644)

	plain := stripANSI(out)
	var plainB strings.Builder
	for _, line := range strings.Split(plain, "\n") {
		plainB.WriteString(strings.TrimRight(line, " "))
		plainB.WriteString("\n")
	}
	_ = os.WriteFile("/tmp/img_fixture_plain.txt", []byte(plainB.String()), 0o644)

	fmt.Println("wrote /tmp/img_fixture_ansi.txt and /tmp/img_fixture_plain.txt")
	fmt.Println("non-empty plain lines:")
	for i, line := range strings.Split(strings.TrimRight(plainB.String(), "\n"), "\n") {
		if strings.TrimSpace(line) != "" {
			fmt.Printf("  [%2d] %s\n", i, line)
		}
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\033' {
			b.WriteByte(s[i])
			continue
		}
		i++
		if i >= len(s) || s[i] != '[' {
			continue
		}
		i++
		for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == ';') {
			i++
		}
		if i < len(s) && s[i] == 'm' {
			continue
		}
		i--
	}
	return b.String()
}

var (
	imageCacheMu sync.RWMutex
	imageCache   = make(map[string]image.Image)
)

func fetchImages(doc *dom.Node, baseURL string) {
	imgs := doc.QueryAll("img")
	if len(imgs) == 0 {
		return
	}
	var sources []string
	seen := make(map[string]struct{})
	for _, node := range imgs {
		src := strings.TrimSpace(node.GetAttr("src"))
		if src == "" {
			continue
		}
		var abs string
		if strings.HasPrefix(src, "data:") {
			abs = src
		} else {
			abs = resolveURL(baseURL, src)
		}
		if abs == "" {
			continue
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		sources = append(sources, abs)
	}
	for _, src := range sources {
		if strings.HasPrefix(src, "data:image/png;base64,") {
			body := strings.TrimPrefix(src, "data:image/png;base64,")
			img, err := decodeBase64PNG(body)
			if err == nil {
				imageCacheMu.Lock()
				imageCache[src] = img
				imageCacheMu.Unlock()
			}
			continue
		}
		body, _, err := fetcher.FetchBytes(src)
		if err != nil || len(body) == 0 {
			continue
		}
		img, err := asciirender.Decode(body)
		if err == nil {
			imageCacheMu.Lock()
			imageCache[src] = img
			imageCacheMu.Unlock()
		}
	}
}

func resolveURL(base, rel string) string {
	if strings.HasPrefix(rel, "http://") || strings.HasPrefix(rel, "https://") {
		return rel
	}
	if strings.HasPrefix(rel, "/") {
		return base + rel
	}
	return base
}

func decodeBase64PNG(b64 string) (image.Image, error) {
	data, err := decodeBase64(b64)
	if err != nil {
		return nil, err
	}
	return asciirender.Decode(data)
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
