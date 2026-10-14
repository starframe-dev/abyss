package layout

import (
	"fmt"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// TestVisualRegression_DeterministicArticleFixture checks a large but stable
// DOM + external CSS → layout pipeline without relying on files in /tmp.
func TestVisualRegression_DeterministicArticleFixture(t *testing.T) {
	var html strings.Builder
	html.WriteString(`<main class="feed">`)
	articleText := strings.Repeat("article body content for deterministic layout ", 30)
	for i := 0; i < 15; i++ {
		fmt.Fprintf(&html, `<article><img src="missing-%d.png" alt="article image"><h2>Article %d</h2><p>%s</p></article>`, i, i, articleText)
	}
	html.WriteString(`</main>`)

	doc := dom.Parse(html.String())
	styles := css.ComputeWithExternal(doc, []string{
		`.feed { width: 100%; } article { margin-bottom: 32px; } img { height: 48px; } p { margin: 4px; }`,
	})
	css.InvalidateCache()

	root := Build(doc, styles, 120)
	if root == nil {
		t.Fatal("root layout box is nil")
	}
	if root.Content.Height < 1000 {
		t.Fatalf("expected root height > 1000, got %v", root.Content.Height)
	}

	articleCount, imageCount := 0, 0
	var walk func(*LayoutBox)
	walk = func(box *LayoutBox) {
		if box.Node != nil && strings.EqualFold(box.Node.TagName, "img") {
			imageCount++
		}
		if box.Node != nil && strings.EqualFold(box.Node.TagName, "article") {
			articleCount++
			if box.Content.Width <= 0 || box.Content.Height <= 0 {
				t.Errorf("article at %v has zero dimensions", box.Content)
			}
			if len(box.Children) == 0 {
				t.Errorf("article at %v has no children", box.Content)
			}
		}
		for _, child := range box.Children {
			walk(child)
		}
	}
	walk(root)
	if articleCount != 15 {
		t.Fatalf("article count = %d, want 15", articleCount)
	}
	if imageCount != 15 {
		t.Fatalf("image count = %d, want 15", imageCount)
	}
	t.Logf("deterministic fixture: %d articles, %d images, root height %v", articleCount, imageCount, root.Content.Height)
}
