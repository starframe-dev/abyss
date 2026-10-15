package js

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestDOMBridgeBasics(t *testing.T) {
	InitDOM(dom.Parse("<html><body><div id=\"target\">old</div></body></html>"))
	defer resetDOMState()

	body := getBody()
	if body == nil || body.TagName != "body" { t.Fatal("getBody did not find body") }
	el := getElementByID("target")
	if el == nil { t.Fatal("getElementByID did not find element") }
	id := getNodeID(el)
	if id < 0 { t.Fatal("registered node has no ID") }
	setAttribute(id, "DATA-X", "yes")
	if got := getAttribute(id, "data-x"); got != "yes" { t.Fatalf("attribute = %q", got) }
	setTextContent(id, "new")
	if len(el.Children) != 1 || el.Children[0].Text != "new" { t.Fatal("setTextContent failed") }
	setInnerHTML(id, "<span>html</span>")
	if len(el.Children) != 1 || el.Children[0].TagName != "span" { t.Fatal("setInnerHTML failed") }

	child := createElement("SECTION")
	if child.TagName != "section" { t.Fatal("createElement did not normalize tag") }
	text := createTextNode("text")
	frag := createDocumentFragment()
	frag.Children = append(frag.Children, text)
	text.Parent = frag
	appendChild(id, getNodeID(frag))
	if len(frag.Children) != 0 || text.Parent != el { t.Fatal("fragment append failed") }
	removeChild(id, getNodeID(text))
	if text.Parent != nil { t.Fatal("removeChild failed") }
	if getAttribute(999999, "x") != "" || getNodeID(nil) != -1 { t.Fatal("invalid node lookup") }
	setAttribute(999999, "x", "y")
	setTextContent(999999, "x")
	setInnerHTML(999999, "x")
	appendChild(999999, getNodeID(child))
	removeChild(id, 999999)
	if getElementByID("missing") != nil { t.Fatal("unexpected element") }
}

func TestDOMNodeOperationsAndDirtyState(t *testing.T) {
	InitDOM(&dom.Node{Type: dom.DocumentNode, Children: []*dom.Node{}})
	defer resetDOMState()
	if getBody() != nil { t.Fatal("body should not exist") }
	SetDirty()
	if !IsDirty() { t.Fatal("SetDirty failed") }
	MarkClean()
	if IsDirty() { t.Fatal("MarkClean failed") }
	SetDirty()
	resetDOMState()
	if IsDirty() { t.Fatal("reset did not clear dirty state") }

	InitDOM(&dom.Node{Type: dom.DocumentNode, Children: []*dom.Node{}})
	parent := createElement("div")
	child := createElement("p")
	pid, cid := getNodeID(parent), getNodeID(child)
	appendChild(pid, cid)
	appendChild(pid, cid)
	if len(parent.Children) != 1 { t.Fatal("append duplicated child") }
	appendChild(cid, pid)
	if child.Parent != parent { t.Fatal("cycle was not rejected") }
	removeChild(pid, cid)
	if len(parent.Children) != 0 || child.Parent != nil { t.Fatal("child removal failed") }
	removeNodeFromParentLocked(child)
	moveAllChildren(parent, createDocumentFragment())
	if wouldCreateCycle(parent, child) { t.Fatal("unexpected cycle") }
}
