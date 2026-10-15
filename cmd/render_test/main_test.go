package main

import (
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestStripANSI(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"hello", "hello"},
		{"\033[31mred\033[0m", "red"},
		{"a\033xb", "ab"},
	} {
		if got := stripANSI(tc.in); got != tc.want { t.Errorf("stripANSI(%q) = %q, want %q", tc.in, got, tc.want) }
	}
}

func TestResolveURL(t *testing.T) {
	for _, tc := range []struct{ base, rel, want string }{
		{"https://example.test/path", "https://other.test/a", "https://other.test/a"},
		{"https://example.test", "/a", "https://example.test/a"},
		{"https://example.test/path", "a", "https://example.test/path"},
	} {
		if got := resolveURL(tc.base, tc.rel); got != tc.want { t.Errorf("resolveURL(%q, %q) = %q, want %q", tc.base, tc.rel, got, tc.want) }
	}
}

func TestDecodeBase64PNG(t *testing.T) {
	var b strings.Builder
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf strings.Builder
	_ = b
	_ = buf
	data := make([]byte, 0)
	_ = data
	if _, err := decodeBase64PNG("not base64!"); err == nil { t.Fatal("expected invalid base64 error") }
	// Verify base64 decoding independently; DecodeBase64PNG also delegates image decoding.
	var out strings.Builder
	_ = out
	if _, err := decodeBase64("%%%="); err == nil { t.Fatal("expected decodeBase64 error") }
	_ = png.BestCompression
	_ = base64.StdEncoding
}

func TestFetchImagesNoImages(t *testing.T) {
	fetchImages(dom.Parse("<html><body>nothing</body></html>"), "https://example.test")
}

func TestFetchImagesDataPNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{G: 255, A: 255})
	var encoded strings.Builder
	if err := png.Encode(&stringWriter{b: &encoded}, img); err != nil { t.Fatal(err) }
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(encoded.String()))
	imageCacheMu.Lock(); imageCache = make(map[string]image.Image); imageCacheMu.Unlock()
	fetchImages(dom.Parse(`<img src="`+dataURL+`">`), "https://example.test")
	imageCacheMu.RLock(); got := imageCache[dataURL]; imageCacheMu.RUnlock()
	if got == nil { t.Fatal("expected decoded image in cache") }
}

type stringWriter struct{ b *strings.Builder }
func (w *stringWriter) Write(p []byte) (int, error) { return w.b.Write(p) }
