package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/fetcher"
)

func main() {
    html, err := fetcher.Fetch("https://habr.com")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fetch error: %v\n", err)
        os.Exit(1)
    }

    doc := dom.Parse(html)

    // Find all <link rel="stylesheet"> and <style> tags
    stylesheets := doc.QueryAll("link[rel=stylesheet]")
    fmt.Printf("External stylesheets: %d\n", len(stylesheets))
    for _, s := range stylesheets {
        href := s.GetAttr("href")
        if len(href) > 100 {
            href = href[:100] + "..."
        }
        fmt.Printf("  %s\n", href)
    }

    styles := doc.QueryAll("style")
    fmt.Printf("\nInline <style> tags: %d\n", len(styles))
    for i, s := range styles {
        text := s.TextContent()
        if len(text) > 500 {
            text = text[:500] + "..."
        }
        fmt.Printf("  [%d] %s\n\n", i, text)
    }

    // Show all tags in the page
    tags := make(map[string]int)
    var allTags []string
    var walk func(n *dom.Node)
    walk = func(n *dom.Node) {
        if n.Type == dom.ElementNode {
            if _, ok := tags[n.TagName]; !ok {
                allTags = append(allTags, n.TagName)
            }
            tags[n.TagName]++
        }
        for _, c := range n.Children {
            walk(c)
        }
    }
    walk(doc)

    fmt.Printf("\nTags (%d unique):\n", len(allTags))
    for _, tag := range allTags {
        fmt.Printf("  %s: %d\n", tag, tags[tag])
    }

    // Show body text content (first 2000 chars)
    body := doc.QueryAll("body")
    if len(body) > 0 {
        text := strings.TrimSpace(body[0].TextContent())
        if len(text) > 2000 {
            text = text[:2000] + "..."
        }
        fmt.Printf("\nBody text (%d chars):\n%s\n", len(body[0].TextContent()), text)
    }
}
