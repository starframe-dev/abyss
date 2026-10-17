package js

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func insertTestSetup() {
	nodeMu.Lock()
	defer nodeMu.Unlock()
	nodes = map[NodeID]*dom.Node{}
	nodeIDs = map[*dom.Node]NodeID{}
	nodeNextID = 1
}

// insertTestNode registers a node the way production does, so later inserts
// allocate IDs that do not collide with it.
func insertTestNode(id NodeID, typ dom.NodeType) *dom.Node {
	n := &dom.Node{Type: typ}
	nodes[id] = n
	nodeIDs[n] = id
	if id >= nodeNextID {
		nodeNextID = id + 1
	}
	return n
}

func TestInsertBefore(t *testing.T) {
	insertTestSetup()
	p := insertTestNode(1, dom.ElementNode)
	a := insertTestNode(2, dom.ElementNode)
	b := insertTestNode(3, dom.ElementNode)
	r := insertTestNode(4, dom.ElementNode)
	p.Children = []*dom.Node{a, r}
	a.Parent, r.Parent = p, p
	insertBefore(1, 3, 4)
	if len(p.Children) != 3 || p.Children[1] != b || b.Parent != p {
		t.Fatal("insert before failed")
	}
	f := insertTestNode(5, dom.DocumentFragmentNode)
	c := &dom.Node{Type: dom.TextNode}
	f.Children = []*dom.Node{c}
	insertBefore(1, 5, 4)
	if len(f.Children) != 0 || p.Children[2] != c || c.Parent != p {
		t.Fatal("fragment insertion failed")
	}
}

func TestReplaceChild(t *testing.T) {
	insertTestSetup()
	p := insertTestNode(1, dom.ElementNode)
	a := insertTestNode(2, dom.ElementNode)
	b := insertTestNode(3, dom.ElementNode)
	p.Children = []*dom.Node{a}
	a.Parent = p
	replaceChild(1, 3, 2)
	if p.Children[0] != b || b.Parent != p || a.Parent != nil {
		t.Fatal("replacement failed")
	}
	f := insertTestNode(4, dom.DocumentFragmentNode)
	c := &dom.Node{Type: dom.TextNode}
	f.Children = []*dom.Node{c}
	replaceChild(1, 4, 3)
	if p.Children[0] != c || len(f.Children) != 0 {
		t.Fatal("fragment replacement failed")
	}
}

func TestInsertAdjacentHTMLAndText(t *testing.T) {
	insertTestSetup()
	p := insertTestNode(1, dom.ElementNode)
	n := insertTestNode(2, dom.ElementNode)
	p.Children = []*dom.Node{n}
	n.Parent = p
	InsertAdjacentHTML(2, "beforeend", "<b>x</b>")
	if len(n.Children) != 1 || n.Children[0].Parent != n {
		t.Fatal("HTML insertion failed")
	}
	insertAdjacentText(2, "afterbegin", "text")
	if len(n.Children) != 2 || n.Children[0].Text != "text" {
		t.Fatal("text insertion failed")
	}
	insertAdjacentText(2, "afterend", "tail")
	if len(p.Children) != 2 || p.Children[1].Text != "tail" {
		t.Fatal("sibling text insertion failed")
	}
}

func TestInsertAdjacentElement(t *testing.T) {
	insertTestSetup()
	p := insertTestNode(1, dom.ElementNode)
	n := insertTestNode(2, dom.ElementNode)
	e := insertTestNode(3, dom.ElementNode)
	p.Children = []*dom.Node{n}
	n.Parent = p
	insertAdjacentElement(2, 3, "beforebegin")
	if len(p.Children) != 2 || p.Children[0] != e || e.Parent != p {
		t.Fatal("element insertion failed")
	}
}
