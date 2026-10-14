package dom

import (
	"strings"
	"sync"
)

// matchesComplex проверяет, соответствует ли узел n сложному селектору.
func matchesComplex(n *Node, sel ComplexSelector) bool {
	if len(sel.Parts) == 0 {
		return false
	}
	return matchPart(n, sel.Parts, len(sel.Parts)-1)
}

// matchPart проверяет соответствие узла части селектора и предшествующей ему цепочки комбинаторов.
func matchPart(n *Node, parts []SelectorPart, idx int) bool {
	if !matchesSimple(n, parts[idx]) {
		return false
	}
	if idx == 0 {
		return true
	}

	comb := parts[idx].Combinator
	switch comb {
	case " ", "":
		// descendant
		for anc := n.Parent; anc != nil; anc = anc.Parent {
			if matchPart(anc, parts, idx-1) {
				return true
			}
		}
	case ">":
		if n.Parent != nil && matchPart(n.Parent, parts, idx-1) {
			return true
		}
	case "+":
		prev := previousElementSibling(n)
		if prev != nil && matchPart(prev, parts, idx-1) {
			return true
		}
	case "~":
		for prev := previousElementSibling(n); prev != nil; prev = previousElementSibling(prev) {
			if matchPart(prev, parts, idx-1) {
				return true
			}
		}
	}
	return false
}

// previousElementSibling возвращает предыдущего элемента-брата узла n.
func previousElementSibling(n *Node) *Node {
	if n == nil || n.Parent == nil {
		return nil
	}
	found := false
	for i := len(n.Parent.Children) - 1; i >= 0; i-- {
		child := n.Parent.Children[i]
		if child == n {
			found = true
			continue
		}
		if found && child.Type == ElementNode {
			return child
		}
	}
	return nil
}

// matchesSimple проверяет, соответствует ли узел простой части селектора.
func matchesSimple(n *Node, part SelectorPart) bool {
	if n == nil || n.Type != ElementNode {
		return false
	}

	// Tag check
	if part.Tag != "" && part.Tag != "*" {
		if !strings.EqualFold(n.TagName, part.Tag) {
			return false
		}
	}

	// ID check
	if part.ID != "" && n.Attributes["id"] != part.ID {
		return false
	}

	// Class check
	classAttr := n.Attributes["class"]
	for _, c := range part.Classes {
		if !hasClass(classAttr, c) {
			return false
		}
	}

	// Attribute check
	for _, attr := range part.Attributes {
		if !matchesAttr(n, attr) {
			return false
		}
	}

	// Pseudo-class check
	for _, pseudo := range part.Pseudos {
		if !matchesPseudo(n, pseudo) {
			return false
		}
	}

	return true
}

// matchesAttr проверяет соответствие узла атрибутному селектору.
func matchesAttr(n *Node, attr AttrSelector) bool {
	val, exists := n.Attributes[attr.Name]
	switch attr.Op {
	case "":
		return exists
	case "=":
		return val == attr.Value
	case "~=":
		for _, f := range strings.Fields(val) {
			if f == attr.Value {
				return true
			}
		}
		return false
	case "|=":
		return val == attr.Value || strings.HasPrefix(val, attr.Value+"-")
	case "^=":
		return strings.HasPrefix(val, attr.Value)
	case "$=":
		return strings.HasSuffix(val, attr.Value)
	case "*=":
		return strings.Contains(val, attr.Value)
	}
	return false
}

// matchesPseudo проверяет псевдокласс узла.
func matchesPseudo(n *Node, pseudo PseudoSelector) bool {
	pseudoMatchersOnce.Do(func() {
		pseudoMatchers = buildPseudoMatchers()
	})
	if fn, ok := pseudoMatchers[strings.ToLower(pseudo.Name)]; ok {
		return fn(n, pseudo)
	}
	return false
}

var (
	pseudoMatchersOnce sync.Once
	pseudoMatchers     map[string]func(*Node, PseudoSelector) bool
)

func buildPseudoMatchers() map[string]func(*Node, PseudoSelector) bool {
	return map[string]func(*Node, PseudoSelector) bool{
		"first-child": func(n *Node, _ PseudoSelector) bool { return childIndex(n, false) == 1 },
		"last-child":  func(n *Node, _ PseudoSelector) bool { return lastChildIndex(n, false) == 1 },
		"only-child": func(n *Node, _ PseudoSelector) bool {
			return childIndex(n, false) == 1 && lastChildIndex(n, false) == 1
		},
		"first-of-type":    func(n *Node, _ PseudoSelector) bool { return childIndex(n, true) == 1 },
		"last-of-type":     func(n *Node, _ PseudoSelector) bool { return lastChildIndex(n, true) == 1 },
		"only-of-type":     func(n *Node, _ PseudoSelector) bool { return childIndex(n, true) == 1 && lastChildIndex(n, true) == 1 },
		"nth-child":        func(n *Node, p PseudoSelector) bool { return matchNth(p.Arg, childIndex(n, false)) },
		"nth-of-type":      func(n *Node, p PseudoSelector) bool { return matchNth(p.Arg, childIndex(n, true)) },
		"nth-last-child":   func(n *Node, p PseudoSelector) bool { return matchNth(p.Arg, lastChildIndex(n, false)) },
		"nth-last-of-type": func(n *Node, p PseudoSelector) bool { return matchNth(p.Arg, lastChildIndex(n, true)) },
		"not":              func(n *Node, p PseudoSelector) bool { return !MatchesSelector(n, p.Arg) },
		"is":               func(n *Node, p PseudoSelector) bool { return MatchesSelector(n, p.Arg) },
		"where":            func(n *Node, p PseudoSelector) bool { return MatchesSelector(n, p.Arg) },
		"empty":            func(n *Node, _ PseudoSelector) bool { return isEmptyNode(n) },
		"checked":          func(n *Node, _ PseudoSelector) bool { return n.GetAttr("checked") != "" },
		"disabled":         func(n *Node, _ PseudoSelector) bool { return n.GetAttr("disabled") != "" },
		"enabled":          func(n *Node, _ PseudoSelector) bool { return n.GetAttr("disabled") == "" },
		"selected":         func(n *Node, _ PseudoSelector) bool { return n.GetAttr("selected") != "" },
		"focus":            func(_ *Node, _ PseudoSelector) bool { return false },
		"root":             func(n *Node, _ PseudoSelector) bool { return n.Type == DocumentNode },
	}
}

// isEmptyNode проверяет, является ли узел пустым.
func isEmptyNode(n *Node) bool {
	for _, c := range n.Children {
		if c.Type == ElementNode || (c.Type == TextNode && strings.TrimSpace(c.Text) != "") {
			return false
		}
	}
	return true
}
