package dom

import "testing"

func TestChildIndex(t *testing.T) {
	parent := &Node{}
	text := &Node{Type: TextNode}
	a := &Node{Type: ElementNode, TagName: "div"}
	b := &Node{Type: ElementNode, TagName: "span"}
	c := &Node{Type: ElementNode, TagName: "DIV"}
	parent.Children = []*Node{text, a, b, c}
	for _, n := range parent.Children {
		n.Parent = parent
	}
	cases := []struct {
		n      *Node
		ofType bool
		want   int
	}{
		{nil, false, 0}, {a, false, 1}, {b, false, 2}, {c, false, 3},
		{c, true, 2}, {b, true, 1}, {&Node{Parent: parent}, false, 0},
		{&Node{Type: ElementNode}, false, 0},
	}
	for _, tc := range cases {
		if got := childIndex(tc.n, tc.ofType); got != tc.want {
			t.Errorf("childIndex(%v, %v)=%d, want %d", tc.n, tc.ofType, got, tc.want)
		}
	}
}

func TestLastChildIndex(t *testing.T) {
	parent := &Node{}
	a := &Node{Type: ElementNode, TagName: "div"}
	b := &Node{Type: ElementNode, TagName: "span"}
	c := &Node{Type: ElementNode, TagName: "DIV"}
	parent.Children = []*Node{a, {Type: TextNode}, b, c}
	for _, n := range parent.Children {
		n.Parent = parent
	}
	cases := []struct {
		n      *Node
		ofType bool
		want   int
	}{{nil, false, 0}, {a, false, 3}, {b, false, 2}, {c, false, 1}, {a, true, 2}, {c, true, 1}, {&Node{Parent: parent}, false, 0}, {&Node{Type: ElementNode}, false, 0}}
	for _, tc := range cases {
		if got := lastChildIndex(tc.n, tc.ofType); got != tc.want {
			t.Errorf("lastChildIndex=%d, want %d", got, tc.want)
		}
	}
}

func TestMatchNth(t *testing.T) {
	cases := []struct {
		arg  string
		idx  int
		want bool
	}{
		{"", 1, false}, {"  ", 1, false}, {"0", 0, true}, {"0", 1, false}, {"EVEN", 4, true}, {"odd", 3, true},
		{"2", 2, true}, {"bad", 2, false}, {"2n+1", 5, true}, {"2n+1", 4, false}, {"3n", 6, true},
		{"-n+3", 4, false}, {"-n+3", 2, true}, {"0n+2", 2, true}, {"0n+2", 3, false},
		{"n", 1, true}, {"2n+1", 0, false},
	}
	for _, tc := range cases {
		if got := matchNth(tc.arg, tc.idx); got != tc.want {
			t.Errorf("matchNth(%q,%d)=%v, want %v", tc.arg, tc.idx, got, tc.want)
		}
	}
}

func TestParseNth(t *testing.T) {
	cases := []struct {
		s    string
		a, b int
		ok   bool
	}{
		{"", 0, 0, false}, {"0", 0, 0, true}, {"n", 1, 0, true}, {"-n+3", -1, 3, true},
		{"2n-1", 2, -1, true}, {" 3 n + 2 ", 3, 2, true}, {"2x+1", 0, 0, false},
		{"n+bad", 0, 0, false}, {"999999999999999999999n", 0, 0, false}, {"+n", 0, 0, false},
	}
	for _, tc := range cases {
		a, b, ok := parseNth(tc.s)
		if a != tc.a || b != tc.b || ok != tc.ok {
			t.Errorf("parseNth(%q)=(%d,%d,%v), want (%d,%d,%v)", tc.s, a, b, ok, tc.a, tc.b, tc.ok)
		}
	}
}
