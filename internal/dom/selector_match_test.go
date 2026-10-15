//go:build legacy

package dom

import "testing"

func TestSelectorMatching(t *testing.T) {
	node := func(tag string, attrs map[string]string, children ...*Node) *Node {
		n := &Node{Type: ElementNode, TagName: tag, Attributes: attrs, Children: children}
		for _, c := range children {
			c.Parent = n
		}
		return n
	}
	text := func(s string) *Node { return &Node{Type: TextNode, Text: s} }
	root := node("section", map[string]string{"id": "root"})
	a := node("div", map[string]string{"class": "x y", "data": "a b", "lang": "en-US", "start": "prefix", "end": "suffix", "contains": "middle"})
	b := node("span", map[string]string{"disabled": "yes"})
	c := node("div", map[string]string{"class": "x"})
	root.Children = []*Node{a, text(" "), b, c}
	for _, child := range root.Children {
		child.Parent = root
	}
	inner := node("em", map[string]string{"checked": "yes"}, text("content"))
	a.Children = []*Node{inner}
	inner.Parent = a

	for _, tc := range []struct {
		name string
		sel  ComplexSelector
		want bool
	}{
		{"empty selector", ComplexSelector{}, false},
		{"case insensitive tag", ComplexSelector{Parts: []SelectorPart{{Tag: "DIV"}}}, false},
	} { // The selector applies to root; tag mismatch is expected.
		if got := matchesComplex(root, tc.sel); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
	if !matchesComplex(a, ComplexSelector{Parts: []SelectorPart{{Tag: "DIV", ID: "", Classes: []string{"x"}, Attributes: []AttrSelector{{Name: "data", Op: "~=", Value: "b"}}}}}) {
		t.Error("simple selector should match")
	}
	for _, comb := range []string{" ", "", ">"} {
		if !matchesComplex(inner, ComplexSelector{Parts: []SelectorPart{{Tag: "div"}, {Combinator: comb, Tag: "em"}}}) {
			t.Errorf("combinator %q", comb)
		}
	}
	for _, pair := range []struct {
		n       *Node
		comb    string
		prevTag string
	}{{b, "+", "div"}, {c, "~", "span"}} {
		if !matchesComplex(pair.n, ComplexSelector{Parts: []SelectorPart{{Tag: pair.prevTag}, {Combinator: pair.comb, Tag: pair.n.TagName}}}) {
			t.Errorf("combinator %s", pair.comb)
		}
	}
	if matchPart(root, []SelectorPart{{Tag: "section"}}, 0) != true {
		t.Error("matchPart")
	}
	if previousElementSibling(nil) != nil || previousElementSibling(root) != nil || previousElementSibling(b) != a || previousElementSibling(c) != b {
		t.Error("previousElementSibling")
	}
	if matchesSimple(nil, SelectorPart{}) || matchesSimple(text("x"), SelectorPart{}) || matchesSimple(root, SelectorPart{Tag: "aside"}) {
		t.Error("matchesSimple rejection")
	}
	if !matchesSimple(a, SelectorPart{Tag: "*", ID: "", Classes: []string{"y"}}) {
		t.Error("wildcard/class")
	}

	for _, tc := range []struct {
		op, value string
		want      bool
	}{
		{"", "", true}, {"=", "a b", true}, {"~=", "a", true}, {"|=", "en", true}, {"^=", "pre", true}, {"$=", "fix", true}, {"*=", "ddl", true}, {"?=", "", false},
	} {
		name := map[string]string{"": "data", "=": "data", "~=": "data", "|=": "lang", "^=": "start", "$=": "end", "*=": "contains", "?=": "data"}[tc.op]
		if got := matchesAttr(a, AttrSelector{Name: name, Op: tc.op, Value: tc.value}); got != tc.want {
			t.Errorf("attribute %s: got %v", tc.op, got)
		}
	}
	if matchesAttr(a, AttrSelector{Name: "missing", Op: "="}) {
		t.Error("missing attribute")
	}
	for _, tc := range []struct {
		name, arg string
		n         *Node
		want      bool
	}{
		{"first-child", "", a, true}, {"last-child", "", c, true}, {"only-child", "", inner, true},
		{"first-of-type", "", a, true}, {"last-of-type", "", c, true}, {"only-of-type", "", inner, true},
		{"nth-child", "2", b, true}, {"nth-of-type", "2", c, true}, {"nth-last-child", "1", c, true}, {"nth-last-of-type", "1", c, true},
		{"not", "span", a, true}, {"is", "div", a, true}, {"where", "div", a, true},
		{"empty", "", b, true}, {"checked", "", inner, true}, {"disabled", "", b, true}, {"enabled", "", a, true}, {"selected", "", a, false}, {"focus", "", a, false},
		{"unknown", "", a, false},
	} {
		if got := matchesPseudo(tc.n, PseudoSelector{Name: tc.name, Arg: tc.arg}); got != tc.want {
			t.Errorf("pseudo %s(%s): got %v want %v", tc.name, tc.arg, got, tc.want)
		}
	}
	doc := &Node{Type: DocumentNode}
	if !matchesPseudo(doc, PseudoSelector{Name: "root"}) {
		t.Error("root pseudo")
	}
	if len(buildPseudoMatchers()) == 0 {
		t.Error("pseudo matcher map")
	}
	for _, tc := range []struct {
		n    *Node
		want bool
	}{{node("p", nil), true}, {node("p", nil, text(" \n")), true}, {node("p", nil, text("x")), false}, {node("p", nil, node("i", nil)), false}} {
		if got := isEmptyNode(tc.n); got != tc.want {
			t.Errorf("isEmptyNode: got %v want %v", got, tc.want)
		}
	}
}
