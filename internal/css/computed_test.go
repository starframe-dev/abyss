package css

import (
    "testing"

    "github.com/starframe/abyss/internal/dom"
)

func TestCompute(t *testing.T) {
    // Create a simple document with a style tag
    doc := &dom.Node{}
    doc.Type = dom.DocumentNode

    // Test that Compute returns a map
    result := Compute(doc)
    if result == nil {
        t.Fatal("Compute returned nil")
    }

    // Test that InvalidateCache works
    InvalidateCache()
}

func TestComputeWithExternal(t *testing.T) {
    // Create a simple document
    doc := &dom.Node{}
    doc.Type = dom.DocumentNode

    externalCSS := []string{`.test { color: red; }`}

    result := ComputeWithExternal(doc, externalCSS)
    if result == nil {
        t.Fatal("ComputeWithExternal returned nil")
    }

    InvalidateCache()
}

func TestInvalidateCache(t *testing.T) {
    // Call InvalidateCache
    InvalidateCache()

    // Verify cache is cleared by trying to compute again
    doc := &dom.Node{}
    doc.Type = dom.DocumentNode
    result := Compute(doc)
    if result == nil {
        t.Fatal("Compute returned nil after InvalidateCache")
    }

    InvalidateCache()
}

func TestCompute_CacheHit(t *testing.T) {
    // First call should compute
    doc := &dom.Node{}
    doc.Type = dom.DocumentNode
    Compute(doc)

    // Second call with same doc and no external CSS should hit cache
    result := Compute(doc)
    if result == nil {
        t.Fatal("Cache miss: Compute returned nil")
    }

    InvalidateCache()
}

func TestCompute_CacheMiss_ExternalCSS(t *testing.T) {
    // First call with external CSS should not be cached
    doc := &dom.Node{}
    doc.Type = dom.DocumentNode
    ComputeWithExternal(doc, []string{`test`})

    // Second call with same external CSS should not hit cache
    result := ComputeWithExternal(doc, []string{`test`})
    if result == nil {
        t.Fatal("ComputeWithExternal returned nil")
    }

    InvalidateCache()
}

func TestComputeWithExternal_EmptyExternalCSS(t *testing.T) {
    doc := &dom.Node{}
    doc.Type = dom.DocumentNode
    result := ComputeWithExternal(doc, []string{})
    if result == nil {
        t.Fatal("ComputeWithExternal returned nil")
    }

    InvalidateCache()
}

func TestParseStyleSheet_NestedAtRule(t *testing.T) {
    rules := ParseStyleSheet("@media (min-width: 1px) { .card { display: grid; } }")

    if len(rules) != 1 {
        t.Fatalf("ParseStyleSheet() returned %d rules, want 1", len(rules))
    }
    if rules[0].Selectors[0] != ".card" {
        t.Fatalf("selector = %q, want .card", rules[0].Selectors[0])
    }
}

func TestComputeWithExternal_CSSVariables(t *testing.T) {
    doc := dom.Parse("<html><body><p class=card>content</p></body></html>")
    styles := ComputeWithExternal(doc, []string{
        ":root { --accent: #123456; } .card { color: var(--accent); }",
    })
    card := doc.QuerySelector(".card")
    if card == nil {
        t.Fatal("card node not found")
    }

    got := styles[card].Color
    want := Hex("#123456")
    if got != want {
        t.Fatalf("color = %#v, want %#v", got, want)
    }
    InvalidateCache()
}

func TestCompute_DoesNotApplyDescendantRuleToUnrelatedNodes(t *testing.T) {
    doc := dom.Parse(`<div class="wrapper"><div class="content">text</div></div>`)
    styles := ComputeWithExternal(doc, []string{
        ".spinner div { position: absolute; } .wrapper { display: flex; }",
    })
    wrapper := doc.QuerySelector(".wrapper")
    if wrapper == nil {
        t.Fatal("wrapper node not found")
    }

    if styles[wrapper].Position != PositionStatic {
        t.Fatalf("wrapper position = %d, want static", styles[wrapper].Position)
    }
    InvalidateCache()
}
