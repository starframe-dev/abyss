package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/fetcher"
	"github.com/starframe-dev/abyss/internal/js"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

func main() {
	url := "https://example.com"
	html, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fetch error: %v\n", err)
		os.Exit(1)
	}

	doc := dom.Parse(html)

	jsc := js.NewContext()
	defer jsc.Close()
	js.InitDOM(doc)
	if err := jsc.RegisterBridgeFunctions(); err != nil {
		fmt.Fprintf(os.Stderr, "Bridge error: %v\n", err)
		os.Exit(1)
	}
	if _, err := jsc.Evaluate(js.DOMAPIScript()); err != nil {
		fmt.Fprintf(os.Stderr, "DOM API error: %v\n", err)
		os.Exit(1)
	}
	if err := js.ExecuteScripts(jsc, doc, url); err != nil {
		fmt.Printf("Script exec error: %v\n", err)
	}

	finalDOM := js.GetFinalDOM()
	styles := css.Compute(finalDOM)
	box := layout.Build(finalDOM, styles, 800)
	fb := render.Render(box, 800)
	lines := ascii.Render(fb, box, 800)

	// Save full file
	os.WriteFile("/tmp/abyss_full_ansi.txt", []byte(strings.Join(lines, "\n")), 0644)

	// Save SHORT version (120 chars per line, terminal-friendly)
	var shortLines []string
	for _, line := range lines {
		clean := stripANSI(line)
		// Truncate text content to 120 chars max
		if len(clean) > 120 {
			clean = clean[:120]
		}
		shortLines = append(shortLines, strings.TrimRight(clean, " "))
	}
	os.WriteFile("/tmp/abyss_short_ansi.txt", []byte(strings.Join(shortLines, "\n")), 0644)

	fmt.Println("Done. Files saved:")
	fmt.Println("  /tmp/abyss_full_ansi.txt (full ANSI, 800 chars per line)")
	fmt.Println("  /tmp/abyss_short_ansi.txt (trimmed to 120 chars)")

	// Show content count
	nonEmpty := 0
	for _, line := range shortLines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}
	fmt.Printf("Non-empty lines: %d out of %d\n", nonEmpty, len(shortLines))

	// Show which lines have content
	for i, line := range shortLines {
		if strings.TrimSpace(line) != "" {
			fmt.Printf("  [%2d] %s\n", i, line)
		}
	}
}

func stripANSI(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\033' {
			result.WriteByte(s[i])
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
	return result.String()
}
