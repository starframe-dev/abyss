package dom

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestParse_Basic(t *testing.T) {
    html := `<html><body><div class="container" id="main"><p>Hello</p></div></body></html>`
    doc := Parse(html)
    assert.NotNil(t, doc)
    assert.Equal(t, DocumentNode, doc.Type)

    body := doc.Children[0].Children[0]
    assert.Equal(t, "body", body.TagName)

    div := body.Children[0]
    assert.Equal(t, "div", div.TagName)
    assert.Equal(t, "container", div.GetAttr("class"))
    assert.Equal(t, "main", div.GetAttr("id"))

    p := div.Children[0]
    assert.Equal(t, "p", p.TagName)
    assert.Equal(t, "Hello", p.Children[0].Text)
}

// TestParse_VoidElements ensures HTML5 void elements (<meta>, <link>, <br>, etc.)
// are not pushed onto the stack even when written without a trailing slash.
// Regression test: previously the x/net/html tokenizer's StartTagToken for these
// would swallow all subsequent siblings as nested children.
func TestParse_VoidElements(t *testing.T) {
    html := `<html><head><meta charset="utf-8"><link rel="icon" href="x.ico"><title>T</title></head><body>B</body></html>`
    doc := Parse(html)

    htmlNode := doc.Children[0]
    assert.Equal(t, "html", htmlNode.TagName)

    // head must contain title, but meta/link must NOT contain body
    head := htmlNode.Children[0]
    assert.Equal(t, "head", head.TagName)

    var headChildren []string
    for _, c := range head.Children {
        headChildren = append(headChildren, c.TagName)
    }
    assert.Equal(t, []string{"meta", "link", "title"}, headChildren)

    // meta and link must have no children
    for _, c := range head.Children {
        if c.TagName == "meta" || c.TagName == "link" {
            assert.Empty(t, c.Children, "%s must be a void element with no children", c.TagName)
        }
    }

    // body must be a sibling of head, not nested in meta/link
    body := htmlNode.Children[1]
    assert.Equal(t, "body", body.TagName)
    assert.Equal(t, "B", body.Children[0].Text)
}

func TestParse_AllVoidElements(t *testing.T) {
    // Note: <area> is excluded because HTML5 parser requires it inside <map>;
    // x/net/html will rewrite the tree when <area> appears standalone.
    voids := []string{"base", "br", "col", "embed", "hr", "img", "input", "keygen", "link", "meta", "param", "source", "track", "wbr"}
    for _, v := range voids {
        t.Run(v, func(t *testing.T) {
            html := `<body><` + v + ` id="x"><p>after</p></body>`
            doc := Parse(html)
            var body *Node
            for _, c := range doc.Children {
                if c.Type == ElementNode && c.TagName == "body" {
                    body = c
                    break
                }
            }
            if body == nil {
                t.Fatalf("no <body> in output for void=%s", v)
            }
            // Find the void element by id within body subtree
            var voidNode *Node
            var walk func(n *Node)
            walk = func(n *Node) {
                if voidNode != nil {
                    return
                }
                if n.Type == ElementNode && n.TagName == v {
                    voidNode = n
                    return
                }
                for _, c := range n.Children {
                    walk(c)
                }
            }
            walk(body)
            if voidNode == nil {
                t.Fatalf("void <%s> not found in body for input with <%s>", v, v)
            }
            assert.Empty(t, voidNode.Children, "<%s> must be void (no children)", v)
        })
    }
}

func TestQueryAll_ByClass(t *testing.T) {
    html := `<div class="a"></div><div class="b"></div><div class="a"></div>`
    doc := Parse(html)
    results := doc.QueryAll(".a")
    assert.Len(t, results, 2)
}

func TestQueryAll_ByID(t *testing.T) {
    html := `<div id="unique"></div><div id="other"></div>`
    doc := Parse(html)
    results := doc.QueryAll("#unique")
    assert.Len(t, results, 1)
    assert.Equal(t, "unique", results[0].GetAttr("id"))
}

func TestQueryAll_ByTag(t *testing.T) {
    html := `<p>One</p><div>Two</div><p>Three</p>`
    doc := Parse(html)
    results := doc.QueryAll("p")
    assert.Len(t, results, 2)
}

func TestTextContent(t *testing.T) {
    html := `<div><p>Hello <span>world</span></p></div>`
    doc := Parse(html)
    p := doc.QueryAll("p")[0]
    assert.Equal(t, "Hello world", p.TextContent())
}

func TestSelector_ChildCombinator(t *testing.T) {
    html := `<div><p><span>inner</span></p><span>outer</span></div>`
    doc := Parse(html)
    results := doc.QueryAll("div > span")
    assert.Len(t, results, 1)
    assert.Equal(t, "outer", results[0].TextContent())
}

func TestSelector_DescendantCombinator(t *testing.T) {
    html := `<div><p><span>inner</span></p><span>outer</span></div>`
    doc := Parse(html)
    results := doc.QueryAll("div span")
    assert.Len(t, results, 2)
}

func TestSelector_Attribute(t *testing.T) {
    html := `<div data-x="1"></div><div data-x="2"></div><div data-y></div>`
    doc := Parse(html)
    assert.Len(t, doc.QueryAll("[data-x]"), 2)
    assert.Len(t, doc.QueryAll("[data-x=1]"), 1)
    assert.Len(t, doc.QueryAll("[data-y]"), 1)
}

func TestSelector_PseudoNthChild(t *testing.T) {
    html := `<ul><li>a</li><li>b</li><li>c</li></ul>`
    doc := Parse(html)
    assert.Len(t, doc.QueryAll("li:nth-child(2)"), 1)
    assert.Equal(t, "b", doc.QueryAll("li:nth-child(2)")[0].TextContent())
    assert.Len(t, doc.QueryAll("li:first-child"), 1)
    assert.Len(t, doc.QueryAll("li:last-child"), 1)
}

func TestSelector_MultipleClasses(t *testing.T) {
    html := `<div class="a b"></div><div class="a"></div><div class="b"></div>`
    doc := Parse(html)
    assert.Len(t, doc.QueryAll(".a.b"), 1)
}

func TestSelector_Matches(t *testing.T) {
    html := `<div id="x" class="a b"></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    assert.True(t, div.Matches("#x"))
    assert.True(t, div.Matches(".a.b"))
    assert.True(t, div.Matches("div#x.a.b"))
    assert.False(t, div.Matches(".c"))
}

func TestSelector_Closest(t *testing.T) {
    html := `<section><div><span>x</span></div></section>`
    doc := Parse(html)
    span := doc.QueryAll("span")[0]
    closest := span.Closest("section")
    assert.NotNil(t, closest)
    assert.Equal(t, "section", closest.TagName)
}

func TestQuerySelector(t *testing.T) {
    html := `<div><span class="target">found</span></div>`
    doc := Parse(html)
    target := doc.QuerySelector("span.target")
    assert.NotNil(t, target)
    assert.Equal(t, "span", target.TagName)
    assert.Equal(t, "found", target.TextContent())

    // Test with no match
    nilResult := doc.QuerySelector("nonexistent")
    assert.Nil(t, nilResult)
}

func TestGetAttr(t *testing.T) {
    html := `<div class="test" id="main" data-custom="value">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    assert.Equal(t, "test", div.GetAttr("class"))
    assert.Equal(t, "main", div.GetAttr("id"))
    assert.Equal(t, "value", div.GetAttr("data-custom"))
    assert.Equal(t, "", div.GetAttr("nonexistent"))
    assert.Equal(t, "", div.GetAttr("CASE-SENSITIVE"))
}

func TestRemoveChild(t *testing.T) {
    html := `<div><span class="to-remove">removed</span><span class="keep">kept</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    // Find the child to remove
    toRemove := div.QueryAll(".to-remove")[0]
    kept := div.QueryAll(".keep")[0]

    // Remove it
    removed := div.RemoveChild(toRemove)
    assert.Equal(t, removed, toRemove)
    assert.Nil(t, toRemove.Parent)
    assert.Len(t, div.Children, 1)

    // Verify the other child is still there
    assert.Equal(t, "span", kept.TagName)
    assert.NotNil(t, kept.Parent)
}

func TestRemoveChild_NotFound(t *testing.T) {
    html := `<div><span class="only">only</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    // Try to remove a node that's not a child
    other := Parse("<span>other</span>").QueryAll("span")[0]
    removed := div.RemoveChild(other)
    assert.Nil(t, removed)
    assert.Len(t, div.Children, 1)
}

func TestAppendChild(t *testing.T) {
    html := `<div><span>existing</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    // Create a proper new node
    newNode := &Node{
        Type:       ElementNode,
        TagName:    "span",
        Text:       "new",
        Attributes: map[string]string{"class": "new"},
        Children:   []*Node{},
        Parent:     nil,
    }

    result := div.AppendChild(newNode)
    assert.Equal(t, result, newNode)
    assert.NotNil(t, newNode.Parent)
    assert.NotNil(t, newNode.Parent.Parent) // div's parent is doc
    assert.Len(t, div.Children, 2)
}

func TestClone_Shallow(t *testing.T) {
    html := `<div id="x" class="a b"><span>nested</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    clone := div.Clone(false)

    assert.Equal(t, "div", clone.TagName)
    assert.Equal(t, "x", clone.GetAttr("id"))
    assert.Equal(t, "a b", clone.GetAttr("class"))

    // Children should be empty for shallow clone
    assert.Empty(t, clone.Children)

    // Original should be unchanged
    assert.Len(t, div.Children, 1)
    assert.Equal(t, "nested", div.Children[0].Children[0].Text)
}

func TestClone_Deep(t *testing.T) {
    html := `<div id="x" class="a b"><span>nested</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    clone := div.Clone(true)

    assert.Equal(t, "div", clone.TagName)
    assert.Equal(t, "x", clone.GetAttr("id"))
    assert.Equal(t, "a b", clone.GetAttr("class"))
    assert.Len(t, clone.Children, 1)
    assert.Equal(t, "nested", clone.Children[0].Children[0].Text)
    assert.Equal(t, "nested", div.Children[0].Children[0].Text)
}

func TestClone_AttributeCopy(t *testing.T) {
    html := `<div data-custom="value">content</div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    clone := div.Clone(true)

    assert.Equal(t, "value", clone.GetAttr("data-custom"))
    assert.Equal(t, "value", div.GetAttr("data-custom"))
}

func TestOuterHTML_Document(t *testing.T) {
    html := `<html><head><title>Test</title></head><body><div>content</div></body></html>`
    doc := Parse(html)

    result := doc.OuterHTML()
    assert.Contains(t, result, "<html>")
    assert.Contains(t, result, "<title>")
    assert.Contains(t, result, "</body>")
}

func TestOuterHTML_Element(t *testing.T) {
    html := `<div class="test"><span>nested</span></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]

    result := div.OuterHTML()
    assert.Contains(t, result, `<div class="test">`)
    assert.Contains(t, result, `</div>`)
    assert.Contains(t, result, `<span`)
    assert.Contains(t, result, `</span>`)
}

func TestOuterHTML_SelfClosing(t *testing.T) {
    html := `<html><body><br><img src="x.png"><br><img src="y.png"></body></html>`
    doc := Parse(html)
    body := doc.QueryAll("body")[0]

    // Get the br and img nodes
    br := body.QueryAll("br")[0]
    img1 := body.QueryAll("img")[0]

    assert.Equal(t, `<br />`, br.OuterHTML())
    assert.Contains(t, img1.OuterHTML(), `<img`)
}

func TestOuterHTML_TextNode(t *testing.T) {
    html := `<div><p>Hello World</p></div>`
    doc := Parse(html)
    p := doc.QueryAll("p")[0]

    result := p.OuterHTML()
    assert.Contains(t, result, `<p>`)
    assert.Contains(t, result, "Hello World")
    assert.Contains(t, result, `</p>`)
}

func TestOuterHTML_Nil(t *testing.T) {
    result := (*Node)(nil).OuterHTML()
    assert.Equal(t, "", result)
}

func TestFindStyleTags(t *testing.T) {
    html := `<html><head><style>.a{}</style><style>.b{}</style><link rel="stylesheet" href="x.css"></head></html>`
    doc := Parse(html)
    styleTags := doc.FindStyleTags()

    assert.Len(t, styleTags, 2)
    assert.Equal(t, "style", styleTags[0].TagName)
    assert.Equal(t, "style", styleTags[1].TagName)
}

func TestFindLinkStylesheets(t *testing.T) {
    html := `<html><head><link rel="stylesheet" href="a.css"><link rel="stylesheet" href="b.css"><style>.x{}</style></head></html>`
    doc := Parse(html)
    linkStylesheets := doc.FindLinkStylesheets()

    assert.Len(t, linkStylesheets, 2)
    assert.Equal(t, "link", linkStylesheets[0].TagName)
    assert.Equal(t, "link", linkStylesheets[1].TagName)
}

func TestParse_EmptyString(t *testing.T) {
    html := ``
    doc := Parse(html)
    assert.NotNil(t, doc)
    assert.Equal(t, DocumentNode, doc.Type)
    assert.Empty(t, doc.Children)
}

func TestParse_WhitespaceOnly(t *testing.T) {
    html := `   `
    doc := Parse(html)
    assert.NotNil(t, doc)
    // Whitespace-only text nodes should be filtered
    assert.Empty(t, doc.Children)
}

func TestParse_SingleTag(t *testing.T) {
    html := `<div>content</div>`
    doc := Parse(html)
    assert.NotNil(t, doc)
    div := doc.QueryAll("div")[0]
    assert.Equal(t, "content", div.TextContent())
}

func TestParse_WithComments(t *testing.T) {
    html := `<div><!-- comment --><p>text</p></div>`
    doc := Parse(html)
    div := doc.QueryAll("div")[0]
    assert.NotNil(t, div)
    // Comments are typically stripped by the parser, so this mainly tests robustness
}

func TestClosest_NilNode(t *testing.T) {
    result := (*Node)(nil).Closest("div")
    assert.Nil(t, result)
}
