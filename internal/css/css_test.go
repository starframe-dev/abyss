package css

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/assert"
)

func TestCSSParseInline(t *testing.T) {
	decls := ParseInline("display: flex; color: red; margin: 10px")
	assert.Len(t, decls, 3)
	assert.Equal(t, "display", decls[0].Property)
	assert.Equal(t, "flex", decls[0].Value)
	assert.Equal(t, "color", decls[1].Property)
	assert.Equal(t, "red", decls[1].Value)
}

func TestCSSParseStyleSheet(t *testing.T) {
	css := `
        body { color: red; }
        .container { display: flex; margin: 10px; }
        #nav { width: 100%; }
    `
	rules := ParseStyleSheet(css)
	assert.Len(t, rules, 3)

	assert.Equal(t, []string{"body"}, rules[0].Selectors)
	assert.Len(t, rules[0].Declarations, 1)

	assert.Equal(t, []string{".container"}, rules[1].Selectors)
	assert.Len(t, rules[1].Declarations, 2)

	assert.Equal(t, []string{"#nav"}, rules[2].Selectors)
}

func TestParseStyleSheet_WithComments(t *testing.T) {
	css := `/* comment */ body { color: blue; }`
	rules := ParseStyleSheet(css)
	assert.Len(t, rules, 1)
	assert.Equal(t, "body", rules[0].Selectors[0])
}

func TestCSSSpecificity(t *testing.T) {
	assert.Equal(t, 1, Specificity("div"))
	assert.Equal(t, 10, Specificity(".class"))
	assert.Equal(t, 100, Specificity("#id"))
	assert.Equal(t, 11, Specificity("div.class"))
	assert.Equal(t, 101, Specificity("div#id"))
}

func TestCompute_InlineStyles(t *testing.T) {
	html := `<div style="display: flex; color: red; background-color: #ff0000;"></div>`
	doc := dom.Parse(html)
	styles := Compute(doc)

	div := doc.Children[0]
	style := styles[div]
	assert.Equal(t, DisplayFlex, style.Display)
	assert.Equal(t, RGB(255, 0, 0), style.Color)
	assert.Equal(t, RGB(255, 0, 0), style.BackgroundColor)
}

func TestCompute_StyleTag(t *testing.T) {
	html := `<style>.box { display: flex; width: 200px; }</style><div class="box"></div>`
	doc := dom.Parse(html)
	styles := Compute(doc)

	div := doc.Children[1]
	style := styles[div]
	assert.Equal(t, DisplayFlex, style.Display)
	assert.Equal(t, Px(200), style.Width)
}

func TestCompute_Specificity(t *testing.T) {
	html := `<style>div { color: blue; } .special { color: red; }</style><div class="special"></div>`
	doc := dom.Parse(html)
	styles := Compute(doc)

	div := doc.Children[1]
	style := styles[div]
	assert.Equal(t, RGB(255, 0, 0), style.Color)
}

func TestCSSParseLength(t *testing.T) {
	assert.Equal(t, Auto(), parseLength("auto"))
	assert.Equal(t, Px(10), parseLength("10px"))
	assert.Equal(t, Percent(50), parseLength("50%"))
	assert.Equal(t, Px(16), parseLength("1rem"))
	assert.Equal(t, Px(32), parseLength("2em"))
}

func TestCSSParseColor(t *testing.T) {
	assert.Equal(t, RGB(255, 0, 0), parseColor("red"))
	assert.Equal(t, RGB(0, 0, 0), parseColor("black"))
	assert.Equal(t, RGB(255, 255, 255), parseColor("white"))
	assert.Equal(t, RGB(255, 0, 0), parseColor("#ff0000"))
	assert.Equal(t, RGB(255, 0, 0), parseColor("#f00"))
	assert.Equal(t, RGB(255, 128, 0), parseColor("rgb(255, 128, 0)"))
	assert.Equal(t, Transparent(), parseColor("transparent"))
}

func TestCSSParseBoxModel(t *testing.T) {
	bm := parseBoxModel("10px")
	assert.Equal(t, Px(10), bm.Top)
	assert.Equal(t, Px(10), bm.Right)
	assert.Equal(t, Px(10), bm.Bottom)
	assert.Equal(t, Px(10), bm.Left)

	bm2 := parseBoxModel("10px 20px")
	assert.Equal(t, Px(10), bm2.Top)
	assert.Equal(t, Px(20), bm2.Right)
}

func TestCSSParseBorder(t *testing.T) {
	b := parseBorder("1px solid red")
	assert.Equal(t, 1.0, b.Width)
	assert.Equal(t, BorderSolid, b.Style)
	assert.Equal(t, RGB(255, 0, 0), b.Color)
}

func TestCSSParseLengthList(t *testing.T) {
	list := parseLengthList("100px 1fr 100px")
	assert.Len(t, list, 3)
	assert.Equal(t, Px(100), list[0])
	assert.Equal(t, Fr(1), list[1])
	assert.Equal(t, Px(100), list[2])

	list2 := parseLengthList("repeat(3, 1fr)")
	assert.Len(t, list2, 3)
}
