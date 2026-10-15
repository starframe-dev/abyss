package js

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestDOMQueryHelpers(t *testing.T) {
	text := &dom.Node{Type: dom.TextNode, Text: "text"}
	a := &dom.Node{Type: dom.ElementNode, TagName: "A", Attributes: map[string]string{"class": "one two"}}
	b := &dom.Node{Type: dom.ElementNode, TagName: "b", Attributes: map[string]string{"class": "two"}}
	root := &dom.Node{Type: dom.DocumentNode, Children: []*dom.Node{text, a, b}}
	text.Parent, a.Parent, b.Parent = root, root, root
	InitDOM(root)
	defer resetDOMState()
	id := func(n *dom.Node) NodeID { return getNodeID(n) }
	if getParentNode(id(a)) != root || getParentNode(-1) != nil { t.Fatal("getParentNode") }
	if len(getChildren(id(root))) != 3 || getChildren(-1) != nil { t.Fatal("getChildren") }
	if got := getElementsByTagName("A"); len(got) != 1 || got[0] != a { t.Fatal("getElementsByTagName") }
	if got := getElementsByClassName("two"); len(got) != 2 { t.Fatal("getElementsByClassName") }
	if querySelector(id(root), "b") != b || querySelector(-1, "b") != nil { t.Fatal("querySelector") }
	if len(querySelectorAll(id(root), "*") ) != 2 || querySelectorAll(-1, "*") != nil { t.Fatal("querySelectorAll") }
	if !matches(id(a), ".one") || matches(-1, "*") { t.Fatal("matches") }
	if closest(id(a), "A") != a || closest(-1, "*") != nil { t.Fatal("closest") }
	if contains(id(root), id(a)) == false || !contains(id(a), id(a)) || contains(id(a), id(b)) { t.Fatal("contains") }
	if previousElementSibling(id(b)) != a || nextElementSibling(id(a)) != b { t.Fatal("sibling query") }
	if firstElementChild(id(root)) != a || lastElementChild(id(root)) != b { t.Fatal("element child query") }
	if previousElementSibling(id(a)) != nil || nextElementSibling(id(b)) != nil || firstElementChild(id(a)) != nil || lastElementChild(id(a)) != nil { t.Fatal("empty boundaries") }
	shallow := cloneNode(id(root), false)
	if shallow == nil || len(shallow.Children) != 0 { t.Fatal("shallow clone") }
	deep := cloneNode(id(root), true)
	if deep == nil || len(deep.Children) != 3 { t.Fatal("deep clone") }
	if cloneNode(-1, true) != nil { t.Fatal("clone invalid node") }
}
