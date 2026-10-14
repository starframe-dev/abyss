package browser

import (
	"context"
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
	"time"

	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/js"
)

func TestCloseJSCContextForTabOnlyClosesOwner(t *testing.T) {
	closeJSCContext()
	ctx := js.NewContext()
	js.SetCurrentContext(ctx, "https://owner.test")
	jsLifecycleMu.Lock()
	jsRuntimeTabID = 17
	jsRuntimeActive = true
	jsLifecycleMu.Unlock()
	t.Cleanup(closeJSCContext)

	closeJSCContextForTab(18)
	if _, ok := jsRuntimes[17]; !ok {
		t.Fatal("closing another tab discarded the suspended runtime")
	}
	jsLifecycleMu.Lock()
	if !activateJSCContextForTabLocked(17) {
		jsLifecycleMu.Unlock()
		t.Fatal("could not restore the other tab's runtime")
	}
	if got, _ := js.GetCurrentContext(); got != ctx {
		jsLifecycleMu.Unlock()
		t.Fatal("restored runtime belongs to a different context")
	}
	jsLifecycleMu.Unlock()

	closeJSCContextForTab(17)
	if got, _ := js.GetCurrentContext(); got != nil {
		t.Fatal("closing the owning tab retained its JSC context")
	}
	if jsRuntimeTabID != 0 {
		t.Fatalf("runtime owner = %d, want 0", jsRuntimeTabID)
	}
}

func TestLoadPageKeepsJSCContextsPerTab(t *testing.T) {
	closeJSCContext()
	t.Cleanup(closeJSCContext)
	newFixture := func(text string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprintf(w, `<html><body><p>%s</p><script>setTimeout(function(){document.body.setAttribute('data-fired','%s')},1)</script></body></html>`, text, text)
		}))
	}
	serverOne := newFixture("one")
	defer serverOne.Close()
	serverTwo := newFixture("two")
	defer serverTwo.Close()

	for i, fixture := range []struct {
		url string
		id  uint64
	}{{serverOne.URL, 101}, {serverTwo.URL, 202}} {
		msg := loadPageForTab(context.Background(), fixture.url, 80, fixture.id)()
		if _, ok := msg.(PageLoadedMsg); !ok {
			t.Fatalf("tab %d load returned %T, want PageLoadedMsg", i, msg)
		}
	}
	time.Sleep(10 * time.Millisecond)
	if !processJSTimers(101, serverOne.URL) || !processJSTimers(202, serverTwo.URL) {
		t.Fatal("per-tab timers did not execute and mark their DOMs dirty")
	}

	jsLifecycleMu.Lock()
	defer jsLifecycleMu.Unlock()
	if !activateJSCContextForTabLocked(101) {
		t.Fatal("first tab JSC runtime missing")
	}
	contextOne, urlOne := js.GetCurrentContext()
	docOne := js.GetFinalDOM()
	textOne := docOne.QuerySelector("p").TextContent()
	captureActiveJSCContextForTabLocked(101)

	if !activateJSCContextForTabLocked(202) {
		t.Fatal("second tab JSC runtime missing")
	}
	contextTwo, urlTwo := js.GetCurrentContext()
	docTwo := js.GetFinalDOM()
	textTwo := docTwo.QuerySelector("p").TextContent()
	captureActiveJSCContextForTabLocked(202)

	if contextOne == contextTwo || urlOne != serverOne.URL || urlTwo != serverTwo.URL {
		t.Fatalf("tab runtimes not isolated: ctx same=%v urls=(%q,%q)", contextOne == contextTwo, urlOne, urlTwo)
	}
	if docOne == docTwo || textOne != "one" || textTwo != "two" {
		t.Fatalf("tab DOMs crossed: first=(%p,%q) second=(%p,%q)", docOne, textOne, docTwo, textTwo)
	}
	if got := docOne.QuerySelector("body").GetAttr("data-fired"); got != "one" {
		t.Fatalf("tab one timer state = %q, want one", got)
	}
	if got := docTwo.QuerySelector("body").GetAttr("data-fired"); got != "two" {
		t.Fatalf("tab two timer state = %q, want two", got)
	}
}

func TestFetchExternalCSS(t *testing.T) {
	doc := dom.Parse("<html><head></head><body></body></html>")
	got := fetchExternalCSS(doc, "http://example.com")
	if len(got) != 0 {
		t.Errorf("fetchExternalCSS() length mismatch: got %d, want 0", len(got))
	}
}

func TestFetchExternalCSSRespectsAggregateByteBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := strings.Repeat(string(r.URL.Path[1]), 1<<20)
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()
	var html strings.Builder
	html.WriteString("<html><head>")
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&html, `<link rel="stylesheet" href="/%d.css">`, i)
	}
	html.WriteString("</head></html>")
	styles := fetchExternalCSS(dom.Parse(html.String()), server.URL)
	var total int
	for _, sheet := range styles {
		total += len(sheet)
	}
	if total > maxExternalCSSBytes {
		t.Fatalf("CSS bytes = %d, limit %d", total, maxExternalCSSBytes)
	}
	if len(styles) != 4 {
		t.Fatalf("accepted stylesheets = %d, want 4", len(styles))
	}
	for i, style := range styles {
		if style[0] != byte('0'+i) {
			t.Fatalf("stylesheet order at %d starts with %q", i, style[0])
		}
	}
}

func TestFetchExternalCSSAppliesLinkMediaAndKeepsDistinctConditions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`.card { color: red; }`))
	}))
	defer server.Close()
	doc := dom.Parse(`<link rel="stylesheet" href="/shared.css" media="screen and (min-width: 600px)"><link rel="stylesheet" href="/shared.css" media="screen and (max-width: 599px)">`)

	got := fetchExternalCSS(doc, server.URL)
	if len(got) != 2 {
		t.Fatalf("stylesheets = %d, want 2 media variants", len(got))
	}
	if !strings.HasPrefix(got[0], "@media screen and (min-width: 600px) {") {
		t.Fatalf("first stylesheet lost media query: %q", got[0])
	}
	if !strings.HasPrefix(got[1], "@media screen and (max-width: 599px) {") {
		t.Fatalf("second stylesheet lost media query: %q", got[1])
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

func TestImageMemoryUsesConservativeEightBytePixelEstimate(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 3, 2))
	if got, want := imageMemory(img), int64(3*2*8); got != want {
		t.Fatalf("estimated bytes = %d, want %d", got, want)
	}
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

	resetImageCache()

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
		t.Fatalf("cache size = %d, want 2", len(imageCache))
	}
	if len(imageCacheOrder) != len(imageCache) {
		t.Fatalf("FIFO index size = %d, cache size = %d", len(imageCacheOrder), len(imageCache))
	}
	if imageCacheOrder[0] != server.URL+"/a.png" || imageCacheOrder[1] != server.URL+"/b.png" {
		t.Fatalf("FIFO order = %v", imageCacheOrder)
	}
	var accounted int64
	for _, src := range imageCacheOrder {
		img, ok := imageCache[src]
		if !ok {
			t.Fatalf("FIFO index contains missing cache key %q", src)
		}
		accounted += imageMemory(img)
	}
	if accounted != imageCacheBytes {
		t.Fatalf("cache byte accounting = %d, actual = %d", imageCacheBytes, accounted)
	}
}

func TestFetchImagesCapsExternalResources(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("not-an-image"))
	}))
	defer server.Close()
	resetImageCache()
	var html strings.Builder
	html.WriteString("<html><body>")
	for i := 0; i < maxExternalImages+10; i++ {
		fmt.Fprintf(&html, `<img src="%s/%d.png">`, server.URL, i)
	}
	html.WriteString("</body></html>")
	fetchImages(dom.Parse(html.String()), server.URL)
	if got := atomic.LoadInt32(&hits); got != maxExternalImages {
		t.Fatalf("external image requests = %d, want %d", got, maxExternalImages)
	}
}

func TestFetchImages_IgnoresDataURLs(t *testing.T) {
	resetImageCache()

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
	resetImageCache()

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
