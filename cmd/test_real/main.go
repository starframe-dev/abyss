package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/fetcher"
	"github.com/starframe-dev/abyss/internal/js"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

func main() {
	urls := []string{
		"https://en.wikipedia.org/wiki/ASCII_art",
		"https://news.ycombinator.com",
	}

	for _, url := range urls {
		start := time.Now()
		html, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Printf("[%s] FETCH ERROR: %v\n", url, err)
			continue
		}
		fetchTime := time.Since(start)

		doc := dom.Parse(html)
		parseTime := time.Since(start)

		jsc := js.NewContext()
		js.InitDOM(doc)
		jsc.RegisterBridgeFunctions()
		jsc.Evaluate(js.DOMAPIScript())
		js.ExecuteScripts(jsc, doc, url)
		jsc.Close()

		finalDOM := js.GetFinalDOM()
		styles := css.Compute(finalDOM)
		box := layout.Build(finalDOM, styles, 800)
		fb := render.Render(box, 800)
		lines := ascii.Render(fb, box, 800)
		totalTime := time.Since(start)

		nonEmpty := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmpty++
			}
		}

		fmt.Printf("=== %s ===\n", url)
		fmt.Printf("  HTML size: %d bytes\n", len(html))
		fmt.Printf("  Fetch: %v | Parse: %v | Total: %v\n", fetchTime, parseTime, totalTime)
		fmt.Printf("  Lines: %d total, %d non-empty\n", len(lines), nonEmpty)
		fmt.Printf("  First 20 non-empty lines:\n")
		count := 0
		for i := 0; i < len(lines) && count < 20; i++ {
			content := strings.TrimRight(lines[i], " ")
			if content != "" {
				fmt.Printf("    [%d] %s\n", i, content)
				count++
			}
		}
		fmt.Println()
	}
}
