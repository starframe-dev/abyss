package browser

import (
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestPublicBrowserPageHelpers(t *testing.T) {
	if got := ResolveURLForTest("https://example.test/a/page", "../b"); got != "https://example.test/b" {
		t.Fatalf("ResolveURLForTest() = %q", got)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(".x { color: red }"))
	}))
	defer server.Close()
	doc := dom.Parse(`<link rel="stylesheet" href="/site.css">`)
	got := FetchExternalCSSForTest(doc, server.URL)
	if len(got) != 1 || got[0] != ".x { color: red }" {
		t.Fatalf("FetchExternalCSSForTest() = %#v", got)
	}
}

func TestImageCacheHelpers(t *testing.T) {
	resetImageCache()
	t.Cleanup(resetImageCache)
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	storeImage("first", img)
	if got := cachedImage("first"); got != img {
		t.Fatal("cached image lookup did not return stored image")
	}
	if got := imgOrNil(img, true); got != img || imgOrNil(img, false) != nil {
		t.Fatal("imgOrNil returned an unexpected value")
	}
	if got := imageMemory(img); got != 32 {
		t.Fatalf("imageMemory() = %d, want 32", got)
	}
	if got := removeImageCacheKey([]string{"a", "b", "c"}, "b"); strings.Join(got, ",") != "a,c" {
		t.Fatalf("removeImageCacheKey() = %v", got)
	}
	if got := removeImageCacheKey([]string{"a"}, "missing"); len(got) != 1 {
		t.Fatalf("remove missing key = %v", got)
	}
}

func TestDecodeDataImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	var buf strings.Builder
	if err := (&png.Encoder{}).Encode(stringWriter{&buf}, img); err != nil {
		t.Fatal(err)
	}
	valid := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(buf.String()))
	if got := decodeDataImage(valid); got == nil || got.Bounds().Dx() != 2 || got.Bounds().Dy() != 3 {
		t.Fatalf("decodeDataImage() = %v", got)
	}
	for _, src := range []string{"https://example.test/x", "data:,x", "data:image/png;base64,!!!", "data:image/png;base64,"} {
		if decodeDataImage(src) != nil {
			t.Errorf("decodeDataImage(%q) should fail", src)
		}
	}
}

func TestFetchImagesContextHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	doc := dom.Parse(`<img src="http://127.0.0.1:1/image.png">`)
	fetchImagesContext(ctx, doc, "http://example.test/")
}
