package dom

import "testing"

// Whitespace between two elements separates their words, so the parser must
// keep it as a text node instead of dropping it.
func TestWhitespaceBetweenElementsIsKept(t *testing.T) {
	doc := Parse(`<body><div><a>Для</a> <a>Конф</a></div></body>`)
	div := findElement(doc, "div")
	if div == nil {
		t.Fatal("div not found")
	}
	if len(div.Children) != 3 {
		t.Fatalf("children = %d, want 3 (a, space, a)", len(div.Children))
	}
	if div.Children[1].Type != TextNode || div.Children[1].Text != " " {
		t.Fatalf("middle child = %+v, want text node with a single space", div.Children[1])
	}
}

func findElement(n *Node, tag string) *Node {
	if n.Type == ElementNode && n.TagName == tag {
		return n
	}
	for _, c := range n.Children {
		if f := findElement(c, tag); f != nil {
			return f
		}
	}
	return nil
}
