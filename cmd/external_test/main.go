package main

import (
    "fmt"
    "os"
    "time"

    "github.com/starframe/abyss/internal/ascii"
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/fetcher"
    "github.com/starframe/abyss/internal/js"
    "github.com/starframe/abyss/internal/layout"
    "github.com/starframe/abyss/internal/render"
)

const (
    viewportWidth        = 800
    defaultClientTimeout = 10 * time.Second
)

var defaultClient = fetcher.DefaultClient

func main() {
    url := "http://localhost:9999/test_external_js.html"

    html, err := fetcher.Fetch(url)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fetch error: %v\n", err)
        os.Exit(1)
    }

    doc := dom.Parse(string(html))

    jsc := js.NewContext()
    defer jsc.Close()

    if err := jsc.RegisterBridgeFunctions(); err != nil {
        fmt.Fprintf(os.Stderr, "Bridge init error: %v\n", err)
        os.Exit(1)
    }

    js.InitDOM(doc)

    if _, err := jsc.Evaluate(js.DOMAPIScript()); err != nil {
        fmt.Fprintf(os.Stderr, "DOM API error: %v\n", err)
        os.Exit(1)
    }

    if err := js.ExecuteScripts(jsc, doc, url); err != nil {
        fmt.Fprintf(os.Stderr, "Script error: %v\n", err)
        os.Exit(1)
    }

    finalDOM := js.GetFinalDOM()

    styles := css.Compute(finalDOM)
    box := layout.Build(finalDOM, styles, viewportWidth)
    fb := render.Render(box, viewportWidth)
    lines := ascii.Render(fb, box, viewportWidth)

    for _, l := range lines {
        fmt.Println(l)
    }
    fmt.Println("=== END ===")
}
