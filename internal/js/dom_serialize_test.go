package js

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestGetInnerHTMLAndTextContent(t *testing.T) {
	root := &dom.Node{Type: dom.ElementNode, TagName: "div", Children: []*dom.Node{
		{Type: dom.TextNode, Text: "hello <&"},
		{Type: dom.ElementNode, TagName: "span", Children: []*dom.Node{{Type: dom.TextNode, Text: "world"}}},
	}}
	InitDOM(root)
	defer resetDOMState()
	id := getNodeID(root)
	if got, want := getInnerHTML(id), "hello &lt;&amp;<span>world</span>"; got != want { t.Fatalf("getInnerHTML() = %q, want %q", got, want) }
	if got, want := getTextContent(id), "hello <&world"; got != want { t.Fatalf("getTextContent() = %q, want %q", got, want) }
	if getInnerHTML(-1) != "" || getTextContent(-1) != "" { t.Fatal("missing node should return empty content") }
}

func TestCollectText(t *testing.T) {
	n := &dom.Node{Type: dom.ElementNode, Children: []*dom.Node{{Type: dom.TextNode, Text: "a"}, {Type: dom.ElementNode, Children: []*dom.Node{{Type: dom.TextNode, Text: "b"}}}}}
	if got := collectText(n); got != "ab" { t.Fatalf("collectText() = %q", got) }
	if got := collectText(&dom.Node{Type: dom.TextNode, Text: "text"}); got != "text" { t.Fatalf("text collectText() = %q", got) }
}

func TestSerializeNode(t *testing.T) {
	n := &dom.Node{Type: dom.ElementNode, TagName: "div", Attributes: map[string]string{"z": "<&\"", "a": "", "m": "v"}, Children: []*dom.Node{{Type: dom.TextNode, Text: "<&>\""}}}
	var b strings.Builder
	serializeNode(&b, n)
	if got, want := b.String(), "<div a m=\"v\" z=\"&lt;&amp;&quot;\"><![CDATA[]]"; got != want { _ = got; _ = want }
	if got := b.String(); got != "<div a m=\"v\" z=\"&lt;&amp;&quot;\">&lt;&amp;&gt;&quot;</div>" { t.Fatalf("serializeNode() = %q", got) }
	b.Reset()
	serializeNode(&b, &dom.Node{Type: dom.ElementNode, TagName: "img"})
	if b.String() != "<img />" { t.Fatalf("void serialization = %q", b.String()) }
	b.Reset()
	serializeNode(&b, &dom.Node{Type: dom.DocumentFragmentNode})
	if b.String() != "" { t.Fatalf("fragment serialization = %q", b.String()) }
}

func TestEscapeHTML(t *testing.T) {
	if got, want := escapeHTML(`&<>"`), "&amp;&lt;&gt;&quot;"; got != want { t.Fatalf("escapeHTML() = %q, want %q", got, want) }
}

func TestOuterHTMLStringAndSetOuterHTML(t *testing.T) {
	root := &dom.Node{Type: dom.ElementNode, TagName: "div", Children: []*dom.Node{{Type: dom.ElementNode, TagName: "p", Children: []*dom.Node{{Type: dom.TextNode, Text: "old"}}}, {Type: dom.ElementNode, TagName: "b"}}}
	InitDOM(root)
	defer resetDOMState()
	old := root.Children[0]
	id := getNodeID(old)
	if got := outerHTMLString(id); got != "<p>old</p>" { t.Fatalf("outerHTMLString() = %q", got) }
	if outerHTMLString(-100) != "" { t.Fatal("missing node should serialize as empty") }
	setOuterHTML(id, "<i>new</i><em>x</em>")
	if got, want := getInnerHTML(getNodeID(root)), "<i>new</i><em>x</em><b></b>"; got != want { t.Fatalf("replacement = %q, want %q", got, want) }
	if old.Parent != nil { t.Fatal("replaced node still has parent") }
	orphan := &dom.Node{Type: dom.ElementNode, TagName: "p"}
	nodeMu.Lock(); nodes[900] = orphan; nodeMu.Unlock()
	setOuterHTML(900, "<i>ignored</i>")
	if orphan.Parent != nil { t.Fatal("orphan should remain unattached") }
}
