package integration

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/fetcher"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
	"github.com/stretchr/testify/require"
)

func TestPageLoadFetchParseLayoutAndRender(t *testing.T) {
	var mu sync.Mutex
	requests := map[string]int{}
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()
		w.Header().Set("Content-Type", "image/png")
		require.NoError(t, png.Encode(w, image.NewUniform(color.RGBA{R: 220, G: 40, B: 30, A: 255})))
	}))
	defer imageServer.Close()

	var pageServer *httptest.Server
	pageServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html><html><head><link rel="stylesheet" href="/site.css"><script src="/app.js"></script></head><body><main><h1>Integration headline</h1><p id="status">Original content</p><img src="` + imageServer.URL + `/pixel.png"></main></body></html>`))
		case "/site.css":
			w.Header().Set("Content-Type", "text/css")
			_, _ = w.Write([]byte("body { color: #123456; background-color: #ffffff; } h1 { color: #ff0000; } main { margin: 1px; }"))
		case "/app.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte("document.getElementById('status').textContent = 'Script executed';"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer pageServer.Close()

	ctx := context.Background()
	html, err := fetcher.FetchContext(ctx, pageServer.URL+"/")
	require.NoError(t, err)
	require.NotEmpty(t, html)
	doc := dom.Parse(html)
	require.NotNil(t, doc)
	require.NotNil(t, doc.QuerySelector("h1"))

	cssText, err := fetcher.FetchCSS(ctx, pageServer.URL+"/site.css")
	require.NoError(t, err)
	require.Contains(t, cssText, "#ff0000")
	jsText, err := fetcher.FetchJS(ctx, pageServer.URL+"/app.js")
	require.NoError(t, err)
	require.Contains(t, jsText, "Script executed")

	// Fetch the linked image as the browser pipeline does before rendering.
	imgBytes, _, err := fetcher.FetchBytesContext(ctx, imageServer.URL+"/pixel.png")
	require.NoError(t, err)
	require.NotEmpty(t, imgBytes)

	styles := css.ComputeWithExternalAtWidth(doc, []string{cssText}, float64(80*cellmetrics.WidthPx))
	box := layout.Build(doc, styles, float64(80*cellmetrics.WidthPx))
	require.NotNil(t, box)
	fb := render.Render(box, 80*cellmetrics.WidthPx)
	ansiLines := ascii.RenderANSI(fb, box, 80)
	require.NotEmpty(t, ansiLines)
	plainLines := make([]string, len(ansiLines))
	for i, line := range ansiLines {
		plainLines[i] = strings.TrimSpace(line)
	}
	plain := strings.Join(plainLines, "\n")
	require.Contains(t, plain, "Integration headline")
	require.Contains(t, plain, "Original content")

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, requests["/"])
	require.Equal(t, 1, requests["/site.css"])
	require.Equal(t, 1, requests["/app.js"])
	require.Equal(t, 1, requests["/pixel.png"])
}
