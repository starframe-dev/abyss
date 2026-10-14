package main

import (
    "strings"
    "testing"
    "time"

    "github.com/starframe/abyss/internal/ascii"
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/fetcher"
    "github.com/starframe/abyss/internal/js"
    "github.com/starframe/abyss/internal/layout"
    "github.com/starframe/abyss/internal/render"
)

func TestMain_SuccessfulFetchAndProcessing(t *testing.T) {
    // Render includes the layout tree for text overlay.
    // Test that main processes URLs correctly when fetch succeeds
    urls := []string{
        "https://en.wikipedia.org/wiki/ASCII_art",
    }

    for _, url := range urls {
        start := time.Now()
        html, err := fetcher.Fetch(url)
        if err != nil {
            t.Logf("SKIP: [%s] FETCH ERROR: %v", url, err)
            continue
        }
        fetchTime := time.Since(start)

        if len(html) == 0 {
            t.Errorf("Expected non-empty HTML for URL: %s", url)
        }

        doc := dom.Parse(html)
        if doc == nil {
            t.Errorf("Expected non-nil DOM for URL: %s", url)
        }

        jsc := js.NewContext()
        if jsc == nil {
            t.Errorf("Expected non-nil JS context")
        }
        js.InitDOM(doc)
        jsc.RegisterBridgeFunctions()
        jsc.Evaluate(js.DOMAPIScript())
        js.ExecuteScripts(jsc, doc, url)
        jsc.Close()

        finalDOM := js.GetFinalDOM()
        if finalDOM == nil {
            t.Errorf("Expected non-nil final DOM")
        }

        styles := css.Compute(finalDOM)
        if styles == nil {
            t.Errorf("Expected non-nil styles")
        }

        box := layout.Build(finalDOM, styles, 800)
        if box == nil {
            t.Errorf("Expected non-nil box")
        }

        fb := render.Render(box, 800)
        if fb == nil {
            t.Errorf("Expected non-nil framebuffer")
        }

        lines := ascii.Render(fb, box, 800)
        if lines == nil {
            t.Errorf("Expected non-nil lines")
        }

        nonEmpty := 0
        for _, line := range lines {
            if strings.TrimSpace(line) != "" {
                nonEmpty++
            }
        }

        if len(lines) == 0 {
            t.Errorf("Expected non-empty lines slice")
        }

        t.Logf("Processed URL: %s, Fetch: %v, Lines: %d/%d", url, fetchTime, nonEmpty, len(lines))
    }
}

func TestMain_NonExistentURL(t *testing.T) {
    // Render includes the layout tree for text overlay.
    // Test that main handles non-existent URLs gracefully
    urls := []string{
        "https://httpbin.org/status/404",
    }

    for _, url := range urls {
        start := time.Now()
        html, err := fetcher.Fetch(url)
        if err != nil {
            t.Logf("SKIP: [%s] FETCH ERROR (expected): %v", url, err)
            continue
        }
        fetchTime := time.Since(start)

        if len(html) == 0 {
            t.Errorf("Expected non-empty HTML for URL: %s", url)
        }

        doc := dom.Parse(html)
        if doc == nil {
            t.Errorf("Expected non-nil DOM for URL: %s", url)
        }

        jsc := js.NewContext()
        if jsc == nil {
            t.Errorf("Expected non-nil JS context")
        }
        js.InitDOM(doc)
        jsc.RegisterBridgeFunctions()
        jsc.Evaluate(js.DOMAPIScript())
        js.ExecuteScripts(jsc, doc, url)
        jsc.Close()

        finalDOM := js.GetFinalDOM()
        if finalDOM == nil {
            t.Errorf("Expected non-nil final DOM")
        }

        styles := css.Compute(finalDOM)
        if styles == nil {
            t.Errorf("Expected non-nil styles")
        }

        box := layout.Build(finalDOM, styles, 800)
        if box == nil {
            t.Errorf("Expected non-nil box")
        }

        fb := render.Render(box, 800)
        if fb == nil {
            t.Errorf("Expected non-nil framebuffer")
        }

        lines := ascii.Render(fb, box, 800)
        if lines == nil {
            t.Errorf("Expected non-nil lines")
        }

        nonEmpty := 0
        for _, line := range lines {
            if strings.TrimSpace(line) != "" {
                nonEmpty++
            }
        }

        if len(lines) == 0 {
            t.Errorf("Expected non-empty lines slice")
        }

        t.Logf("Processed URL: %s, Fetch: %v, Lines: %d/%d", url, fetchTime, nonEmpty, len(lines))
    }
}

func TestMain_MultipleURLs(t *testing.T) {
    // Render includes the layout tree for text overlay.
    // Test that main processes multiple URLs
    urls := []string{
        "https://en.wikipedia.org/wiki/HTML",
        "https://example.org",
    }

    for _, url := range urls {
        start := time.Now()
        html, err := fetcher.Fetch(url)
        if err != nil {
            t.Logf("SKIP: [%s] FETCH ERROR: %v", url, err)
            continue
        }
        fetchTime := time.Since(start)

        if len(html) == 0 {
            t.Errorf("Expected non-empty HTML for URL: %s", url)
        }

        doc := dom.Parse(html)
        if doc == nil {
            t.Errorf("Expected non-nil DOM for URL: %s", url)
        }

        jsc := js.NewContext()
        if jsc == nil {
            t.Errorf("Expected non-nil JS context")
        }
        js.InitDOM(doc)
        jsc.RegisterBridgeFunctions()
        jsc.Evaluate(js.DOMAPIScript())
        js.ExecuteScripts(jsc, doc, url)
        jsc.Close()

        finalDOM := js.GetFinalDOM()
        if finalDOM == nil {
            t.Errorf("Expected non-nil final DOM")
        }

        styles := css.Compute(finalDOM)
        if styles == nil {
            t.Errorf("Expected non-nil styles")
        }

        box := layout.Build(finalDOM, styles, 800)
        if box == nil {
            t.Errorf("Expected non-nil box")
        }

        fb := render.Render(box, 800)
        if fb == nil {
            t.Errorf("Expected non-nil framebuffer")
        }

        lines := ascii.Render(fb, box, 800)
        if lines == nil {
            t.Errorf("Expected non-nil lines")
        }

        nonEmpty := 0
        for _, line := range lines {
            if strings.TrimSpace(line) != "" {
                nonEmpty++
            }
        }

        if len(lines) == 0 {
            t.Errorf("Expected non-empty lines slice")
        }

        t.Logf("Processed URL: %s, Fetch: %v, Lines: %d/%d", url, fetchTime, nonEmpty, len(lines))
    }
}

func TestMain_EmptyURLSlice(t *testing.T) {
    // Render includes the layout tree for text overlay.
    // Test that main handles empty URL slice gracefully
    urls := []string{}

    for _, url := range urls {
        start := time.Now()
        html, err := fetcher.Fetch(url)
        if err != nil {
            t.Errorf("Unexpected error for empty URL: %v", err)
        }
        fetchTime := time.Since(start)

        if len(html) == 0 {
            t.Errorf("Expected non-empty HTML for empty URL")
        }

        doc := dom.Parse(html)
        if doc == nil {
            t.Errorf("Expected non-nil DOM for empty URL")
        }

        jsc := js.NewContext()
        if jsc == nil {
            t.Errorf("Expected non-nil JS context")
        }
        js.InitDOM(doc)
        jsc.RegisterBridgeFunctions()
        jsc.Evaluate(js.DOMAPIScript())
        js.ExecuteScripts(jsc, doc, url)
        jsc.Close()

        finalDOM := js.GetFinalDOM()
        if finalDOM == nil {
            t.Errorf("Expected non-nil final DOM")
        }

        styles := css.Compute(finalDOM)
        if styles == nil {
            t.Errorf("Expected non-nil styles")
        }

        box := layout.Build(finalDOM, styles, 800)
        if box == nil {
            t.Errorf("Expected non-nil box")
        }

        fb := render.Render(box, 800)
        if fb == nil {
            t.Errorf("Expected non-nil framebuffer")
        }

        lines := ascii.Render(fb, box, 800)
        if lines == nil {
            t.Errorf("Expected non-nil lines")
        }

        nonEmpty := 0
        for _, line := range lines {
            if strings.TrimSpace(line) != "" {
                nonEmpty++
            }
        }

        if len(lines) == 0 {
            t.Errorf("Expected non-empty lines slice")
        }

        t.Logf("Processed empty URL, Fetch: %v, Lines: %d/%d", fetchTime, nonEmpty, len(lines))
    }
}
