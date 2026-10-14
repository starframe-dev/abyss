package dom

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestParseSelectorGroup_Empty(t *testing.T) {
    got, err := ParseSelectorGroup("")
    assert.NoError(t, err)
    assert.NotNil(t, got)
    assert.Empty(t, got.Selectors)
}

func TestParseSelectorGroup_SimpleTag(t *testing.T) {
    got, err := ParseSelectorGroup("div")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
    assert.Empty(t, got.Selectors[0].Parts[0].Classes)
    assert.Empty(t, got.Selectors[0].Parts[0].ID)
}

func TestParseSelectorGroup_Class(t *testing.T) {
    got, err := ParseSelectorGroup(".class")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Empty(t, got.Selectors[0].Parts[0].Tag)
    assert.Len(t, got.Selectors[0].Parts[0].Classes, 1)
    assert.Equal(t, "class", got.Selectors[0].Parts[0].Classes[0])
}

func TestParseSelectorGroup_ID(t *testing.T) {
    got, err := ParseSelectorGroup("#id")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Empty(t, got.Selectors[0].Parts[0].Tag)
    assert.Equal(t, "id", got.Selectors[0].Parts[0].ID)
}

func TestParseSelectorGroup_Attribute(t *testing.T) {
    got, err := ParseSelectorGroup("[attr]")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Len(t, got.Selectors[0].Parts[0].Attributes, 1)
    assert.Equal(t, "attr", got.Selectors[0].Parts[0].Attributes[0].Name)
    assert.Equal(t, "", got.Selectors[0].Parts[0].Attributes[0].Op)
    assert.Empty(t, got.Selectors[0].Parts[0].Attributes[0].Value)
}

func TestParseSelectorGroup_AttributeEq(t *testing.T) {
    got, err := ParseSelectorGroup("[attr=value]")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Len(t, got.Selectors[0].Parts[0].Attributes, 1)
    assert.Equal(t, "=", got.Selectors[0].Parts[0].Attributes[0].Op)
    assert.Equal(t, "value", got.Selectors[0].Parts[0].Attributes[0].Value)
}

func TestParseSelectorGroup_AttributeStartsWith(t *testing.T) {
    got, err := ParseSelectorGroup("[attr^=val]")
    assert.NoError(t, err)
    assert.Equal(t, "^=", got.Selectors[0].Parts[0].Attributes[0].Op)
}

func TestParseSelectorGroup_AttributeEndsWith(t *testing.T) {
    got, err := ParseSelectorGroup("[attr$=val]")
    assert.NoError(t, err)
    assert.Equal(t, "$=", got.Selectors[0].Parts[0].Attributes[0].Op)
}

func TestParseSelectorGroup_AttributeContains(t *testing.T) {
    got, err := ParseSelectorGroup("[attr*=val]")
    assert.NoError(t, err)
    assert.Equal(t, "*=", got.Selectors[0].Parts[0].Attributes[0].Op)
}

func TestParseSelectorGroup_AttributeIncludes(t *testing.T) {
    got, err := ParseSelectorGroup("[attr~=val]")
    assert.NoError(t, err)
    assert.Equal(t, "~=", got.Selectors[0].Parts[0].Attributes[0].Op)
}

func TestParseSelectorGroup_AttributePartial(t *testing.T) {
    got, err := ParseSelectorGroup("[attr|=val]")
    assert.NoError(t, err)
    assert.Equal(t, "|=", got.Selectors[0].Parts[0].Attributes[0].Op)
}

func TestParseSelectorGroup_Pseudo(t *testing.T) {
    got, err := ParseSelectorGroup(":pseudo")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors[0].Parts[0].Pseudos, 1)
    assert.Equal(t, "pseudo", got.Selectors[0].Parts[0].Pseudos[0].Name)
}

func TestParseSelectorGroup_PseudoWithArgs(t *testing.T) {
    got, err := ParseSelectorGroup(":pseudo(arg)")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors[0].Parts[0].Pseudos, 1)
    assert.Equal(t, "pseudo", got.Selectors[0].Parts[0].Pseudos[0].Name)
}

func TestParseSelectorGroup_TagClassId(t *testing.T) {
    got, err := ParseSelectorGroup("div.class#id")
    assert.NoError(t, err)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
    assert.Len(t, got.Selectors[0].Parts[0].Classes, 1)
    assert.Equal(t, "class", got.Selectors[0].Parts[0].Classes[0])
    assert.Equal(t, "id", got.Selectors[0].Parts[0].ID)
}

func TestParseSelectorGroup_TagAttr(t *testing.T) {
    got, err := ParseSelectorGroup("div[attr]")
    assert.NoError(t, err)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
    assert.Len(t, got.Selectors[0].Parts[0].Attributes, 1)
}

func TestParseSelectorGroup_MultipleClasses(t *testing.T) {
    got, err := ParseSelectorGroup(".class1.class2")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors[0].Parts[0].Classes, 2)
    assert.Equal(t, "class1", got.Selectors[0].Parts[0].Classes[0])
    assert.Equal(t, "class2", got.Selectors[0].Parts[0].Classes[1])
}

func TestParseSelectorGroup_WhitespaceTrimming(t *testing.T) {
    got, err := ParseSelectorGroup("  div  ")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
}

func TestParseSelectorGroup_WithComma(t *testing.T) {
    got, err := ParseSelectorGroup("div,span")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 2)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
    assert.Equal(t, "span", got.Selectors[1].Parts[0].Tag)
}

func TestParseSelectorGroup_WithCombinators(t *testing.T) {
    got, err := ParseSelectorGroup("div>span")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
    assert.Equal(t, "span", got.Selectors[0].Parts[1].Tag)
    assert.Equal(t, ">", got.Selectors[0].Parts[1].Combinator)
}

func TestParseSelectorGroup_WithAdjacentSibling(t *testing.T) {
    got, err := ParseSelectorGroup("div + span")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
    assert.Equal(t, "span", got.Selectors[0].Parts[1].Tag)
    assert.Equal(t, "+", got.Selectors[0].Parts[1].Combinator)
}

func TestParseSelectorGroup_UniversalSelector(t *testing.T) {
    got, err := ParseSelectorGroup("*")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Equal(t, "*", got.Selectors[0].Parts[0].Tag)
}

func TestParseSelectorGroup_Complex(t *testing.T) {
    got, err := ParseSelectorGroup("div.class#id[attr]:pseudo")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors, 1)
    assert.Equal(t, "div", got.Selectors[0].Parts[0].Tag)
}

func TestMatchesSelector_TagMatch(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    assert.True(t, MatchesSelector(div, "div"))
}

func TestMatchesSelector_TagMismatch(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    assert.False(t, MatchesSelector(div, "span"))
}

func TestMatchesSelector_ClassMatch(t *testing.T) {
    html := `<div class="test">content</div>`
    doc := Parse(html)
    div := doc.QueryAll(".test")[0]
    assert.True(t, MatchesSelector(div, ".test"))
}

func TestMatchesSelector_ClassPartialMatch(t *testing.T) {
    html := `<div class="a b">content</div>`
    doc := Parse(html)
    div := doc.QueryAll(".a")[0]
    assert.True(t, MatchesSelector(div, ".a"))
    assert.True(t, MatchesSelector(div, ".b"))
    assert.False(t, MatchesSelector(div, ".c"))
}

func TestMatchesSelector_IDMatch(t *testing.T) {
    html := `<div id="main">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("#main")[0]
    assert.True(t, MatchesSelector(div, "#main"))
}

func TestMatchesSelector_IDMismatch(t *testing.T) {
    html := `<div id="main">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("#main")[0]
    assert.False(t, MatchesSelector(div, "#other"))
}

func TestMatchesSelector_AttributeExist(t *testing.T) {
    html := `<div data-x="1">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x]")[0]
    assert.True(t, MatchesSelector(div, "[data-x]"))
}

func TestMatchesSelector_AttributeMissing(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    assert.False(t, MatchesSelector(div, "[data-x]"))
}

func TestMatchesSelector_AttributeEq(t *testing.T) {
    html := `<div data-x="1">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x=1]")[0]
    assert.True(t, MatchesSelector(div, "[data-x=1]"))
}

func TestMatchesSelector_AttributeEqMismatch(t *testing.T) {
    html := `<div data-x="1">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x=1]")[0]
    assert.False(t, MatchesSelector(div, "[data-x=2]"))
}

func TestMatchesSelector_AttributeStartsWith(t *testing.T) {
    html := `<div data-x="x1">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x^=x]")[0]
    assert.True(t, MatchesSelector(div, "[data-x^=x]"))
}

func TestMatchesSelector_AttributeEndsWith(t *testing.T) {
    html := `<div data-x="1x">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x$=x]")[0]
    assert.True(t, MatchesSelector(div, "[data-x$=x]"))
}

func TestMatchesSelector_AttributeContains(t *testing.T) {
    html := `<div data-x="xyz">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x*=x]")[0]
    assert.True(t, MatchesSelector(div, "[data-x*=x]"))
}

func TestMatchesSelector_AttributeIncludes(t *testing.T) {
    html := `<div class="x y">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[class~=y]")[0]
    assert.True(t, MatchesSelector(div, "[class~=y]"))
}

func TestMatchesSelector_AttributePartial(t *testing.T) {
    html := `<div data-x="x-icon">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("[data-x|=x]")[0]
    assert.True(t, MatchesSelector(div, "[data-x|=x]"))
}

func TestMatchesSelector_PseudoFirstChild(t *testing.T) {
    html := `<div><span>first</span><span>second</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    span := div.QueryAll("span")[0]
    assert.True(t, MatchesSelector(span, ":first-child"))
}

func TestMatchesSelector_PseudoLastChild(t *testing.T) {
    html := `<div><span>first</span><span>second</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    spans := div.QueryAll("span")
    span := spans[1]
    assert.True(t, MatchesSelector(span, ":last-child"))
}

func TestMatchesSelector_PseudoNthChild(t *testing.T) {
    html := `<ul><li>a</li><li>b</li><li>c</li></ul>`
    doc := Parse(html)
    assert.Len(t, doc.QueryAll("li:nth-child(2)"), 1)
    assert.Equal(t, "b", doc.QueryAll("li:nth-child(2)")[0].TextContent())
}

func TestMatchesSelector_PseudoNot(t *testing.T) {
    html := `<div><span class="x"></span><span class="y"></span></div>`
    doc := Parse(html)
    span := doc.QueryAll("span")[1]
    assert.True(t, MatchesSelector(span, ":not(.x)"))
}

func TestMatchesSelector_NilElement(t *testing.T) {
    assert.False(t, MatchesSelector(nil, "div"))
}

func TestMatchesSelector_TextNode(t *testing.T) {
    html := `<div>text</div>`
    doc := Parse(html)
    textNode := doc.QueryAll("div")[0].Children[0]
    assert.False(t, MatchesSelector(textNode, "div"))
}

func TestQueryAllSelector_All(t *testing.T) {
    html := `<div><span>a</span><span>b</span><span>c</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span")
    assert.Len(t, results, 3)
}

func TestQueryAllSelector_SpecificClass(t *testing.T) {
    html := `<div><span class="x">a</span><span class="y">b</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span.x")
    assert.Len(t, results, 1)
    assert.Equal(t, "a", results[0].TextContent())
}

func TestQueryAllSelector_SpecificID(t *testing.T) {
    html := `<div><span id="first">a</span><span id="second">b</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span#first")
    assert.Len(t, results, 1)
    assert.Equal(t, "a", results[0].TextContent())
}

func TestQueryAllSelector_Attribute(t *testing.T) {
    html := `<div><span data-x="1">a</span><span data-x="2">b</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span[data-x]")
    assert.Len(t, results, 2)
}

func TestQueryAllSelector_AttributeEq(t *testing.T) {
    html := `<div><span data-x="1">a</span><span data-x="2">b</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span[data-x=1]")
    assert.Len(t, results, 1)
    assert.Equal(t, "a", results[0].TextContent())
}

func TestQueryAllSelector_Tag(t *testing.T) {
    html := `<div><p>one</p><div>two</div><p>three</p></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "p")
    assert.Len(t, results, 2)
}

func TestQueryAllSelector_Universal(t *testing.T) {
    html := `<div><p>one</p><span>two</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "*")
    assert.Len(t, results, 3) // div, p, span
}

func TestQueryAllSelector_ComplexCombinator(t *testing.T) {
    html := `<div><p><span>nested</span></p><span>sibling</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "div > span")
    assert.Len(t, results, 1)
    assert.Equal(t, "sibling", results[0].TextContent())
}

func TestQueryAllSelector_ComplexDescendant(t *testing.T) {
    html := `<div><p><span>nested</span></p><span>sibling</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "div span")
    assert.Len(t, results, 2)
}

func TestQueryAllSelector_ComplexClass(t *testing.T) {
    html := `<div><span class="a b">multi</span><span class="a">single</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span.a.b")
    assert.Len(t, results, 1)
    assert.Equal(t, "multi", results[0].TextContent())
}

func TestQueryAllSelector_ComplexID(t *testing.T) {
    html := `<div><span id="x" class="a">x</span><span id="y" class="a">y</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span#x.a")
    assert.Len(t, results, 1)
}

func TestQueryAllSelector_ComplexAttr(t *testing.T) {
    html := `<div><span data-x="1" class="a">x</span><span data-x="2" class="a">y</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span[data-x=1].a")
    assert.Len(t, results, 1)
}

func TestQueryAllSelector_ComplexPseudo(t *testing.T) {
    html := `<div><span class="a">1</span><span class="a">2</span><span class="a">3</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "span.a:first-child")
    assert.Len(t, results, 1)
}

func TestQueryAllSelector_EmptySelector(t *testing.T) {
    html := `<div><span>content</span></div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "")
    assert.Empty(t, results)
}

func TestQueryAllSelector_InvalidSelector(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)
    results := QueryAllSelector(doc, "[")
    assert.Nil(t, results)
}

func TestQuerySelectorSelector_First(t *testing.T) {
    html := `<div><span>a</span><span>b</span></div>`
    doc := Parse(html)
    result := QuerySelectorSelector(doc, "span")
    assert.NotNil(t, result)
    assert.Equal(t, "a", result.TextContent())
}

func TestQuerySelectorSelector_None(t *testing.T) {
    html := `<div><span>a</span></div>`
    doc := Parse(html)
    result := QuerySelectorSelector(doc, "nonexistent")
    assert.Nil(t, result)
}

func TestQuerySelectorSelector_Universal(t *testing.T) {
    html := `<div><p>one</p><span>two</span></div>`
    doc := Parse(html)
    result := QuerySelectorSelector(doc, "*")
    assert.NotNil(t, result)
    assert.Equal(t, "div", result.TagName)
}

func TestQuerySelectorSelector_Complex(t *testing.T) {
    html := `<div><span class="target">found</span></div>`
    doc := Parse(html)
    result := QuerySelectorSelector(doc, "span.target")
    assert.NotNil(t, result)
    assert.Equal(t, "span", result.TagName)
}

func TestQuerySelectorSelector_Attribute(t *testing.T) {
    html := `<div><span data-x="1">a</span><span data-x="2">b</span></div>`
    doc := Parse(html)
    result := QuerySelectorSelector(doc, "span[data-x=1]")
    assert.NotNil(t, result)
    assert.Equal(t, "a", result.TextContent())
}

func TestQuerySelectorSelector_Pseudo(t *testing.T) {
    html := `<ul><li>a</li><li>b</li><li>c</li></ul>`
    doc := Parse(html)
    result := QuerySelectorSelector(doc, "li:first-child")
    assert.NotNil(t, result)
    assert.Equal(t, "a", result.TextContent())
}

func TestParseSelectorGroup_ErrorCases(t *testing.T) {
    _, err := ParseSelectorGroup("[attr")
    assert.Error(t, err)
}

func TestMatchesSelector_ErrorCases(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    // Invalid selector with element
    result := MatchesSelector(div, "[attr")
    assert.False(t, result)
}

func TestQueryAllSelector_ErrorCases(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)

    // Invalid selector
    results := QueryAllSelector(doc, "[attr")
    assert.Nil(t, results)
}

func TestQuerySelectorSelector_ErrorCases(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)

    // Invalid selector
    result := QuerySelectorSelector(doc, "[attr")
    assert.Nil(t, result)
}

func TestParseSelectorGroup_WhitespaceHandling(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantTags []string
    }{
        {"leading whitespace", "  div  ", []string{"div"}},
        {"trailing whitespace", "div  ", []string{"div"}},
        {"surrounding whitespace", "  div  ", []string{"div"}},
        {"multiple selectors", " div , span ", []string{"div", "span"}},
        {"newline in selector", "div\nspan", []string{"div"}},
        {"tab in selector", "div\tspan", []string{"div"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseSelectorGroup(tt.input)
            assert.NoError(t, err)
            if len(tt.wantTags) != len(got.Selectors) {
                t.Errorf("got %d selectors, want %d", len(got.Selectors), len(tt.wantTags))
            }
        })
    }
}

func TestParseSelectorGroup_AttributeValueQuotes(t *testing.T) {
    got, err := ParseSelectorGroup("[attr='value']")
    assert.NoError(t, err)
    assert.Equal(t, "value", got.Selectors[0].Parts[0].Attributes[0].Value)

    got2, err := ParseSelectorGroup(`[attr="value"]`)
    assert.NoError(t, err)
    assert.Equal(t, "value", got2.Selectors[0].Parts[0].Attributes[0].Value)
}

func TestParseSelectorGroup_PseudoElement(t *testing.T) {
    got, err := ParseSelectorGroup("div::before")
    assert.NoError(t, err)
    assert.Len(t, got.Selectors[0].Parts[0].Pseudos, 1)
    assert.Equal(t, "before", got.Selectors[0].Parts[0].Pseudos[0].Name)
}

func TestMatchesSelector_TagNameCaseInsensitive(t *testing.T) {
    html := `<DIV>content</DIV>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    assert.True(t, MatchesSelector(div, "div"))
}

func TestMatchesSelector_ClassNameCaseSensitive(t *testing.T) {
    html := `<div class="MyClass">content</div>`
    doc := Parse(html)
    div := doc.QueryAll(".MyClass")[0]
    assert.True(t, MatchesSelector(div, ".MyClass"))
    assert.False(t, MatchesSelector(div, ".myclass"))
}

func TestQueryAllSelector_ByClass(t *testing.T) {
    html := `<div class="a"></div><div class="b"></div><div class="a"></div>`
    doc := Parse(html)
    results := doc.QueryAll(".a")
    assert.Len(t, results, 2)
}

func TestQueryAllSelector_ByID(t *testing.T) {
    html := `<div id="unique"></div><div id="other"></div>`
    doc := Parse(html)
    results := doc.QueryAll("#unique")
    assert.Len(t, results, 1)
    assert.Equal(t, "unique", results[0].GetAttr("id"))
}

func TestQueryAllSelector_ByTag(t *testing.T) {
    html := `<p>One</p><div>Two</div><p>Three</p>`
    doc := Parse(html)
    results := doc.QueryAll("p")
    assert.Len(t, results, 2)
}

func TestQueryAllSelector_ComplexSelector(t *testing.T) {
    html := `<div class="container"><span class="item">one</span><span class="item">two</span></div>`
    doc := Parse(html)
    results := doc.QueryAll(".container .item")
    assert.Len(t, results, 2)
}

func TestQueryAllSelector_ComplexWithAttr(t *testing.T) {
    html := `<div data-x="1"><span class="target">found</span></div>`
    doc := Parse(html)
    results := doc.QueryAll("[data-x=1] .target")
    assert.Len(t, results, 1)
}

func TestQueryAllSelector_MultipleCombinators(t *testing.T) {
    html := `<div><p><span>nested</span></p></div>`
    doc := Parse(html)
    results := doc.QueryAll("div > p > span")
    assert.Len(t, results, 1)
    assert.Equal(t, "nested", results[0].TextContent())
}

func TestMatchesSelector_PseudoOnlyChild(t *testing.T) {
    html := `<div><span>only</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    span := div.QueryAll("span")[0]
    assert.True(t, MatchesSelector(span, ":only-child"))
}

func TestMatchesSelector_PseudoOnlyOfType(t *testing.T) {
    html := `<div><span>one</span><span>two</span><span>three</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    span := div.QueryAll("span")[2]
    assert.False(t, MatchesSelector(span, ":only-of-type"))
}

func TestMatchesSelector_PseudoNthOfChild(t *testing.T) {
    html := `<ul><li>a</li><li>b</li><li>c</li></ul>`
    doc := Parse(html)
    assert.Len(t, doc.QueryAll("li:nth-of-type(2)"), 1)
}

func TestMatchesSelector_PseudoNthLastChild(t *testing.T) {
    html := `<div><span>one</span><span>two</span><span>three</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    spans := div.QueryAll("span")
    span := spans[2]
    assert.True(t, MatchesSelector(span, ":nth-last-child(1)"))
}

func TestMatchesSelector_PseudoNthLastOfType(t *testing.T) {
    html := `<div><span>one</span><span>two</span><span>three</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    spans := div.QueryAll("span")
    span := spans[2]
    assert.True(t, MatchesSelector(span, ":nth-last-of-type(1)"))
}

func TestQueryAllSelector_PseudoNthLastChild(t *testing.T) {
    html := `<ul><li>a</li><li>b</li><li>c</li></ul>`
    doc := Parse(html)
    results := doc.QueryAll("li:nth-last-child(3)")
    assert.Len(t, results, 1)
    assert.Equal(t, "a", results[0].TextContent())
}

func TestQueryAllSelector_PseudoNthLastOfType(t *testing.T) {
    html := `<ul><li>a</li><li>b</li><li>c</li></ul>`
    doc := Parse(html)
    results := doc.QueryAll("li:nth-last-of-type(3)")
    assert.Len(t, results, 1)
    assert.Equal(t, "a", results[0].TextContent())
}
