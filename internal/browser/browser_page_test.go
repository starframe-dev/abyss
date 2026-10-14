package browser

import (
    "encoding/base64"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync"
    "sync/atomic"
    "testing"

    "github.com/starframe/abyss/internal/dom"
)

func TestFetchExternalCSS(t *testing.T) {
    doc := dom.Parse("<html><head></head><body></body></html>")
    got := fetchExternalCSS(doc, "http://example.com")
    if len(got) != 0 {
        t.Errorf("fetchExternalCSS() length mismatch: got %d, want 0", len(got))
    }
}

func TestFetchExternalCSS_LoadsAllUniqueStylesheetsInOrder(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = fmt.Fprintf(w, "/* %s */", r.URL.Path)
    }))
    defer server.Close()

    doc := dom.Parse(`<html><head>
        <link rel="stylesheet" href="/0.css">
        <link rel="stylesheet" href="/1.css">
        <link rel="stylesheet" href="/0.css">
        <link rel="stylesheet" href="/2.css">
        <link rel="stylesheet" href="/3.css">
        <link rel="stylesheet" href="/4.css">
    </head><body></body></html>`)
    got := fetchExternalCSS(doc, server.URL+"/page")

    want := []string{"/* /0.css */", "/* /1.css */", "/* /2.css */", "/* /3.css */", "/* /4.css */"}
    if len(got) != len(want) {
        t.Fatalf("fetchExternalCSS() returned %d stylesheets, want %d", len(got), len(want))
    }
    for i := range want {
        if got[i] != want[i] {
            t.Errorf("stylesheet[%d] = %q, want %q", i, got[i], want[i])
        }
    }
}

func TestResolveURL(t *testing.T) {
    tests := []struct {
        name    string
        base    string
        rel     string
        want    string
        wantErr bool
    }{
        {
            name:    "absolute URL",
            base:    "http://example.com/page.html",
            rel:     "http://example.com/styles.css",
            want:    "http://example.com/styles.css",
            wantErr: false,
        },
        {
            name:    "relative URL",
            base:    "http://example.com/path/page.html",
            rel:     "../styles.css",
            want:    "http://example.com/styles.css",
            wantErr: false,
        },
        {
            name:    "relative URL with same base path",
            base:    "http://example.com/path/page.html",
            rel:     "styles.css",
            want:    "http://example.com/path/styles.css",
            wantErr: false,
        },
        {
            name:    "query string preserved",
            base:    "http://example.com/page.html?foo=bar",
            rel:     "styles.css",
            want:    "http://example.com/styles.css",
            wantErr: false,
        },
        {
            name:    "fragment preserved",
            base:    "http://example.com/page.html#section",
            rel:     "styles.css",
            want:    "http://example.com/styles.css",
            wantErr: false,
        },
        {
            name:    "invalid base URL",
            base:    "not-a-valid-url",
            rel:     "styles.css",
            want:    "/styles.css",
            wantErr: true,
        },
        {
            name:    "invalid relative URL",
            base:    "http://example.com/page.html",
            rel:     "not-a-valid-url",
            want:    "http://example.com/not-a-valid-url",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ResolveURL(tt.base, tt.rel)
            if got != tt.want {
                t.Errorf("ResolveURL() = %q, want %q", got, tt.want)
            }
        })
    }
}

func TestResolveURL_Parallel(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            got := ResolveURL("http://example.com/page.html", "styles.css")
            if got != "http://example.com/styles.css" {
                t.Errorf("resolveURL() case %d: got %q, want %q", i, got, "http://example.com/styles.css")
            }
        }(i)
    }
    wg.Wait()
}

func TestFetchImages_PopulatesCacheAndDedupes(t *testing.T) {
    var hits int32
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.URL.Path, ".png") {
            http.NotFound(w, r)
            return
        }
        // Encode a real 2x1 RGBA PNG so the decoder accepts it.
        img := image.NewRGBA(image.Rect(0, 0, 2, 1))
        img.Set(0, 0, color.RGBA{200, 30, 30, 255})
        img.Set(1, 0, color.RGBA{30, 200, 30, 255})
        w.Header().Set("Content-Type", "image/png")
        _ = png.Encode(w, img)
        atomic.AddInt32(&hits, 1)
    }))
    defer server.Close()

    imageCacheMu.Lock()
    imageCache = make(map[string]image.Image)
    imageCacheMu.Unlock()

    doc := dom.Parse(`<html><body>
        <img src="` + server.URL + `/a.png"/>
        <img src="` + server.URL + `/a.png"/>
        <img src="` + server.URL + `/b.png"/>
    </body></html>`)
    fetchImages(doc, server.URL)

    if got := atomic.LoadInt32(&hits); got != 2 {
        t.Errorf("hits = %d, want 2 (deduplicated)", got)
    }
    imageCacheMu.RLock()
    defer imageCacheMu.RUnlock()
    if len(imageCache) != 2 {
        t.Errorf("cache size = %d, want 2", len(imageCache))
    }
}

func TestFetchImages_IgnoresDataURLs(t *testing.T) {
    imageCacheMu.Lock()
    imageCache = make(map[string]image.Image)
    imageCacheMu.Unlock()

    doc := dom.Parse(`<html><body><img src="data:image/png;base64,iVBORw0KGgo="/></body></html>`)
    fetchImages(doc, "http://example.test/")

    imageCacheMu.RLock()
    defer imageCacheMu.RUnlock()
    if len(imageCache) != 0 {
        t.Errorf("cache should remain empty for data URLs, got %d", len(imageCache))
    }
}

// TestCachedImage_DataURLReturnsDecodedImage verifies that the render-time
// hook (which calls ImageProvider = cachedImage) decodes inline data: URLs
// on demand instead of falling back to the [image] placeholder.
func TestCachedImage_DataURLReturnsDecodedImage(t *testing.T) {
    imageCacheMu.Lock()
    imageCache = make(map[string]image.Image)
    imageCacheMu.Unlock()

    img := image.NewRGBA(image.Rect(0, 0, 4, 4))
    for x := 0; x < 4; x++ {
        for y := 0; y < 4; y++ {
            img.Set(x, y, color.RGBA{R: 200, G: 30, B: 30, A: 255})
        }
    }
    var buf strings.Builder
    if err := (&png.Encoder{}).Encode(stringWriter{&buf}, img); err != nil {
        t.Fatalf("encode png: %v", err)
    }
    b64 := base64.StdEncoding.EncodeToString([]byte(buf.String()))
    dataURL := "data:image/png;base64," + b64

    got := cachedImage(dataURL)
    if got == nil {
        t.Fatalf("cachedImage(data:URL) returned nil")
    }
    bounds := got.Bounds()
    if bounds.Dx() != 4 || bounds.Dy() != 4 {
        t.Errorf("decoded image dimensions: got %dx%d, want 4x4", bounds.Dx(), bounds.Dy())
    }

    imageCacheMu.RLock()
    _, cached := imageCache[dataURL]
    imageCacheMu.RUnlock()
    if !cached {
        t.Errorf("data:URL image should be cached after first lookup")
    }
}

// stringWriter adapts strings.Builder to io.Writer for the png encoder.
type stringWriter struct {
    b *strings.Builder
}

func (s stringWriter) Write(p []byte) (int, error) {
    return s.b.Write(p)
}

func TestCachedImage_InvalidDataURLReturnsNil(t *testing.T) {
    if cachedImage("data:image/png;base64,!!!not-base64!!!") != nil {
        t.Errorf("invalid base64 should yield nil image")
    }
    if cachedImage("data:image/png,no-base64-marker") != nil {
        t.Errorf("non-base64 data URL should yield nil image")
    }
    if cachedImage("https://example.test/img.png") != nil {
        t.Errorf("http URL not in cache should yield nil")
    }
}
