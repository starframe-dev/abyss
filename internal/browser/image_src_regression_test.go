package browser

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

// A relative <img src> must resolve to the same key the renderer looks up,
// otherwise the decoded image is cached but never drawn.
func TestFetchImagesMakesRelativeSrcFindCachedImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 4, 4))
		img.Set(0, 0, color.RGBA{10, 10, 10, 255})
		w.Header().Set("Content-Type", "image/png")
		_ = png.Encode(w, img)
	}))
	defer server.Close()

	resetImageCache()
	doc := dom.Parse(`<html><body><img alt="Google" src="/logo.png"></body></html>`)
	fetchImages(doc, server.URL)

	imgNode := doc.QueryAll("img")[0]
	src := imgNode.GetAttr("src")
	if src != server.URL+"/logo.png" {
		t.Fatalf("src = %q, want absolute URL", src)
	}
	if cachedImage(src) == nil {
		t.Fatal("renderer lookup by rewritten src returned no image")
	}
}
