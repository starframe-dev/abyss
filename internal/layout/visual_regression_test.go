package layout

import (
    "fmt"
    "os"
    "strings"
    "testing"

    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
)

// TestVisualRegression_HabrFixture_PipelineProducesArticles verifies that
// the pipeline (DOM + external CSS → layout) for Habr fixture produces
// well-formed article boxes with non-zero dimensions and image placeholders.
func TestVisualRegression_HabrFixture_PipelineProducesArticles(t *testing.T) {
    htmlBytes, err := os.ReadFile("/tmp/abyss-habr-fixture/index.html")
    if err != nil {
        t.Skipf("habr fixture not found: %v", err)
    }
    doc := dom.Parse(string(htmlBytes))

    var external []string
    for i := 0; i < 53; i++ {
        b, err := os.ReadFile(fmt.Sprintf("/tmp/abyss-habr-fixture/css-%d.css", i))
        if err == nil {
            external = append(external, string(b))
        }
    }

    styles := css.ComputeWithExternal(doc, external)
    css.InvalidateCache()

    root := Build(doc, styles, 120)
    if root == nil {
        t.Fatal("root layout box is nil")
    }

    if root.Content.Height < 1000 {
        t.Fatalf("expected root height > 1000, got %v", root.Content.Height)
    }

    articleCount := 0
    var imageCount int
    var walk func(*LayoutBox, *dom.Node)
    walk = func(b *LayoutBox, n *dom.Node) {
        if b.Node != nil && strings.EqualFold(b.Node.TagName, "img") {
            imageCount++
        }
        if b.Node != nil && strings.EqualFold(b.Node.TagName, "article") {
            articleCount++
            if b.Content.Width <= 0 || b.Content.Height <= 0 {
                t.Errorf("article at %v has zero dimensions", b.Content)
            }
            if len(b.Children) == 0 {
                t.Errorf("article at %v has no children", b.Content)
            }
        }
        for _, c := range b.Children {
            walk(c, nil)
        }
    }
    walk(root, nil)

    if articleCount < 10 {
        t.Fatalf("expected >=10 articles, got %d", articleCount)
    }
    if imageCount < 10 {
        t.Errorf("expected >=10 image placeholders, got %d", imageCount)
    }
    t.Logf("habr fixture: %d articles, %d images, root height %v",
        articleCount, imageCount, root.Content.Height)
}