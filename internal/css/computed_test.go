package css

import (
	"fmt"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func BenchmarkComputeStyles(b *testing.B) {
	doc := dom.Parse(`<main><section class="card"><h2>Title</h2><p>Text</p></section></main>`)
	rules := []string{`.card { display: flex; color: red; gap: 8px; } .card p { color: blue; }`}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InvalidateCache()
		_ = ComputeWithExternal(doc, rules)
	}
}

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

func TestSpecificityCountsTypeSelectorsWithPseudoAndAttributeSelectors(t *testing.T) {
	for selector, want := range map[string]int{
		"div:hover":   11,
		"div[attr]":   11,
		"div::before": 2,
		".a.b":        20,
		"#id .a":      110,
	} {
		if got := Specificity(selector); got != want {
			t.Errorf("Specificity(%q) = %d, want %d", selector, got, want)
		}
	}
}

func TestParseStyleSheetBoundsNestedAtRuleDepth(t *testing.T) {
	input := strings.Repeat("@media screen {", maxNestedAtRuleDepth+8) + `.card { color: red; }` + strings.Repeat("}", maxNestedAtRuleDepth+8)
	if rules := ParseStyleSheet(input); len(rules) != 0 {
		t.Fatalf("over-depth nested rules = %d, want none", len(rules))
	}
}

func TestParseStyleSheetIgnoresEmptyRule(t *testing.T) {
	if rules := ParseStyleSheet(`{}`); len(rules) != 0 {
		t.Fatalf("empty rule produced %d parsed rules", len(rules))
	}
}

func TestParseStyleSheet_NestedAtRule(t *testing.T) {
	rules := ParseStyleSheet("@media (min-width: 1px) { .card { display: grid; } }")

	if len(rules) != 1 {
		t.Fatalf("ParseStyleSheet() returned %d rules, want 1", len(rules))
	}
	if rules[0].Selectors[0] != ".card" {
		t.Fatalf("selector = %q, want .card", rules[0].Selectors[0])
	}
	if rules[0].MediaQuery != "(min-width: 1px)" {
		t.Fatalf("media query = %q, want (min-width: 1px)", rules[0].MediaQuery)
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

func TestMediaQueriesMatchViewportWidthAndCacheByWidth(t *testing.T) {
	doc := dom.Parse(`<style>@media (min-width: 600px) { .card { color: red; } } @media screen and (max-width: 599px) { .card { color: blue; } } @supports (display: grid) { .card { display: grid; } } @supports (display: masonry) { .card { background-color: red; } }</style><style media="screen and (min-width: 600px)">.card { width: 123px; }</style><p class="card">text</p>`)
	card := doc.QuerySelector(".card")
	if got, want := ComputeAtWidth(doc, 400)[card].Color, RGB(0, 0, 255); got != want {
		t.Fatalf("narrow media color = %#v, want %#v", got, want)
	}
	wide := ComputeAtWidth(doc, 800)[card]
	if got, want := wide.Color, RGB(255, 0, 0); got != want {
		t.Fatalf("wide media color = %#v, want %#v", got, want)
	}
	if wide.Display != DisplayGrid {
		t.Fatalf("supported @supports rule was ignored: display=%v", wide.Display)
	}
	if wide.Width != Px(123) {
		t.Fatalf("style[media] width = %v, want 123px", wide.Width)
	}
	if wide.BackgroundColor.IsSet {
		t.Fatalf("unsupported @supports rule was applied: background=%v", wide.BackgroundColor)
	}
	InvalidateCache()
}

func TestCSSImportantOverridesSpecificityAndInlineNormal(t *testing.T) {
	doc := dom.Parse(`<p id="one" class="card" style="color:purple"></p><p id="two" class="card" style="color:#00ff00!important"></p>`)
	styles := ComputeWithExternal(doc, []string{
		`.card { color: red !important; } #one { color: blue; } #two { color: blue !important; }`,
	})
	if got, want := styles[doc.QuerySelector("#one")].Color, RGB(255, 0, 0); got != want {
		t.Fatalf("important stylesheet color = %#v, want %#v", got, want)
	}
	if got, want := styles[doc.QuerySelector("#two")].Color, RGB(0, 255, 0); got != want {
		t.Fatalf("important inline color = %#v, want %#v", got, want)
	}
	InvalidateCache()
}

func TestInvalidHexColorIsNotBlack(t *testing.T) {
	for _, value := range []string{"#12345", "#ggg", "123456"} {
		if color := Hex(value); color.IsSet {
			t.Errorf("Hex(%q) returned set color %#v", value, color)
		}
	}
}

func TestBoxModelIsZeroRejectsNonPixelZeroLengths(t *testing.T) {
	zero := BoxModel{Top: Px(0), Right: Px(0), Bottom: Px(0), Left: Px(0)}
	if !zero.IsZero() {
		t.Fatal("zero-pixel box model reported nonzero")
	}
	for _, length := range []Length{Auto(), Percent(0), Fr(0), Px(1)} {
		box := BoxModel{Top: length}
		if box.IsZero() {
			t.Errorf("box model with top=%v reported zero", length)
		}
	}
}

func TestCSSVariableCyclesAreBoundedAndUseFallback(t *testing.T) {
	variables := map[string]string{"--first": "var(--second)", "--second": "var(--first)"}
	if got, want := resolveVariables("VAR(--accent)", map[string]string{"--accent": "#abcdef"}), "#abcdef"; got != want {
		t.Fatalf("case-insensitive var() = %q, want %q", got, want)
	}
	if got, want := resolveVariables("var(--first, #123456)", variables), "#123456"; got != want {
		t.Fatalf("cycle fallback = %q, want %q", got, want)
	}
	if got := resolveVariables("var(--first)", variables); got != "" {
		t.Fatalf("unresolved cycle = %q, want empty", got)
	}

	chain := make(map[string]string, maxVariableResolutionDepth+1)
	for i := 0; i <= maxVariableResolutionDepth; i++ {
		name := fmt.Sprintf("--v%d", i)
		if i == maxVariableResolutionDepth {
			chain[name] = "done"
		} else {
			chain[name] = fmt.Sprintf("var(--v%d)", i+1)
		}
	}
	if got := resolveVariables("var(--v0)", chain); got != "" {
		t.Fatalf("over-depth variable chain = %q, want empty", got)
	}
}

func TestCSSVariableMayBeDeclaredAfterUse(t *testing.T) {
	doc := dom.Parse(`<p class="card">text</p>`)
	styles := ComputeWithExternal(doc, []string{`.card { color: var(--accent); --accent: #123456; }`})
	if got, want := styles[doc.QuerySelector(".card")].Color, Hex("#123456"); got != want {
		t.Fatalf("variable-resolved color = %#v, want %#v", got, want)
	}
	InvalidateCache()
}

func TestComputeClampsNegativeFlexFactors(t *testing.T) {
	doc := dom.Parse(`<div style="flex-grow:-1;flex-shrink:-2"></div>`)
	style := Compute(doc)[doc.QuerySelector("div")]
	if style.FlexGrow != 0 || style.FlexShrink != 0 {
		t.Fatalf("negative flex factors were retained: grow=%v shrink=%v", style.FlexGrow, style.FlexShrink)
	}
	InvalidateCache()
}

func TestComputeInheritsColorButAllowsChildOverride(t *testing.T) {
	doc := dom.Parse(`<div class="parent"><span class="child">text</span><span class="override">x</span></div>`)
	styles := ComputeWithExternal(doc, []string{`.parent { color: red; } .override { color: blue; }`})
	if got, want := styles[doc.QuerySelector(".child")].Color, RGB(255, 0, 0); got != want {
		t.Fatalf("inherited color = %#v, want %#v", got, want)
	}
	if got, want := styles[doc.QuerySelector(".override")].Color, RGB(0, 0, 255); got != want {
		t.Fatalf("child color = %#v, want %#v", got, want)
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
