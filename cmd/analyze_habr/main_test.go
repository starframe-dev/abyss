package main

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestAnalyzeHelpers_DOMQuery(t *testing.T) {
	doc := dom.Parse(`<html><head><style>body{color:red}</style></head><body><a href="/x">link</a><link rel="stylesheet" href="/style.css"></body></html>`)

	stylesheets := doc.QueryAll("link[rel=stylesheet]")
	if len(stylesheets) != 1 {
		t.Errorf("expected 1 stylesheet, got %d", len(stylesheets))
	}

	styles := doc.QueryAll("style")
	if len(styles) != 1 {
		t.Errorf("expected 1 inline style, got %d", len(styles))
	}
}

func TestTagCountingLogic(t *testing.T) {
	doc := dom.Parse(`<html><body><p>a</p><p>b</p><div><span>x</span></div></body></html>`)

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

	if tags["p"] != 2 {
		t.Errorf("expected 2 <p>, got %d", tags["p"])
	}
	if tags["div"] != 1 {
		t.Errorf("expected 1 <div>, got %d", tags["div"])
	}
	if tags["span"] != 1 {
		t.Errorf("expected 1 <span>, got %d", tags["span"])
	}
}

func TestBodyTextExtraction(t *testing.T) {
	doc := dom.Parse(`<html><body>Hello World</body></html>`)
	body := doc.QueryAll("body")
	if len(body) == 0 {
		t.Fatal("body not found")
	}
	text := strings.TrimSpace(body[0].TextContent())
	if text != "Hello World" {
		t.Errorf("body text = %q, want 'Hello World'", text)
	}
}
