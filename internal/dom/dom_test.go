package dom

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseClosingTagMatchesOpenAncestor(t *testing.T) {
	doc := Parse(`<main><section><p>text</main><footer>safe</footer>`)
	main := doc.QuerySelector("main")
	if main == nil || main.QuerySelector("section") == nil {
		t.Fatal("expected malformed descendants to remain under main")
	}
	if doc.QuerySelector("footer") == nil {
		t.Fatal("closing main should pop through its descendants")
	}
}

func TestParseFragmentReturnsDirectFragmentChildren(t *testing.T) {
	fragment := ParseFragment(`before <em>inside</em> after`)
	if fragment.Type != DocumentFragmentNode || len(fragment.Children) != 3 {
		t.Fatalf("fragment structure unexpected: %#v", fragment)
	}
	if fragment.Children[1].TagName != "em" || fragment.Children[1].Parent != fragment {
		t.Fatal("expected direct em child")
	}
	assert.Equal(t, "before inside after", fragment.TextContent())
}

func TestParseEnforcesNodeAndDepthLimits(t *testing.T) {
	if got := countTreeNodes(Parse(strings.Repeat("<x>", MaxDOMDepth+20) + "tail")); got > MaxDOMDepth+1 {
		t.Fatalf("depth limit exceeded: %d", got)
	}
	if got := countTreeNodes(Parse(strings.Repeat("<x></x>", MaxDOMNodes+20))); got > MaxDOMNodes {
		t.Fatalf("node limit exceeded: %d", got)
	}
}

func countTreeNodes(n *Node) int {
	if n == nil {
		return 0
	}
	count := 1
	for _, child := range n.Children {
		count += countTreeNodes(child)
	}
	return count
}

func TestParseRecoversImplicitClosuresAndAttributes(t *testing.T) {
	doc := Parse(`<div><p>one<p>two</div><ul><li>a<li>b</ul><input disabled>`)
	ps, lis := doc.QueryAll("p"), doc.QueryAll("li")
	if len(ps) != 2 || ps[0].Parent != ps[1].Parent || len(lis) != 2 || lis[0].Parent != lis[1].Parent {
		t.Fatal("implicit closure failed")
	}
	assert.Equal(t, "", doc.QuerySelector("input").GetAttr("missing"))
}

func TestDOMPublicOperations(t *testing.T) {
	doc := Parse(`<div id="x" class="a b"><span>hi</span><br></div>`)
	div := doc.QuerySelector("div")
	assert.True(t, div.Matches("#x.a.b"))
	assert.False(t, div.Matches(".missing"))
	assert.Equal(t, "div", div.QuerySelector("span").Closest("div").TagName)
	assert.Len(t, div.QueryAll("span"), 1)
	assert.Equal(t, "x", div.GetAttr("ID"))
	assert.Equal(t, "", (&Node{}).GetAttr("x"))
	assert.Equal(t, "hi", div.TextContent())
	assert.Contains(t, div.OuterHTML(), "<br />")

	shallow := div.Clone(false)
	assert.Empty(t, shallow.Children)
	deep := div.Clone(true)
	assert.Equal(t, "hi", deep.TextContent())
	assert.Same(t, deep, deep.Children[0].Parent)
	deep.Children[0].Attributes["new"] = "yes"
	assert.Empty(t, div.Children[0].GetAttr("new"))

	added := &Node{Type: ElementNode, TagName: "i"}
	assert.Same(t, added, div.AppendChild(added))
	assert.Same(t, div, added.Parent)
	assert.Same(t, added, div.RemoveChild(added))
	assert.Nil(t, added.Parent)
	assert.Nil(t, div.RemoveChild(added))
}

func TestSelectorAndSerializationHelpers(t *testing.T) {
	doc := Parse(`<style>x</style><link rel="stylesheet"><link rel="icon"><p>hello</p>`)
	assert.Len(t, doc.FindStyleTags(), 1)
	assert.Len(t, doc.FindLinkStylesheets(), 1)
	assert.Equal(t, "p", doc.QuerySelector("p").TagName)
	assert.Nil(t, doc.QuerySelector("absent"))
	assert.Nil(t, (*Node)(nil).Closest("div"))
	assert.Equal(t, "", (*Node)(nil).OuterHTML())
	assert.Equal(t, `&lt;&amp;&gt;&quot;`, (&Node{Type: TextNode, Text: `<&>"`}).OuterHTML())
	assert.Equal(t, `<div title="a&amp;b">x</div>`, (&Node{Type: ElementNode, TagName: "div", Attributes: map[string]string{"title": "a&b"}, Text: "x"}).OuterHTML())
	assert.Equal(t, "", (&Node{Type: CommentNode, Text: ""}).TextContent())
}

func TestParseVoidAndWhitespace(t *testing.T) {
	doc := Parse(`<div>   <br><span>x</span></div>`)
	div := doc.QuerySelector("div")
	assert.Empty(t, div.QuerySelector("br").Children)
	assert.Equal(t, "x", div.TextContent())
	assert.Empty(t, Parse("   ").Children)
	assert.Equal(t, DocumentNode, Parse("").Type)
}
