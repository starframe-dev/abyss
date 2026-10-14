package dom

import (
	"fmt"
	"strings"
	"unicode"
)

// SelectorGroup represents a comma-separated list of selectors.
type SelectorGroup struct {
	Selectors []ComplexSelector
}

// ComplexSelector represents a sequence of simple selectors connected by combinators.
type ComplexSelector struct {
	Parts []SelectorPart
}

// SelectorPart is a simple selector plus the combinator that precedes it (for parts[1:]).
type SelectorPart struct {
	Tag        string // empty means any tag; "*" means universal
	ID         string
	Classes    []string
	Attributes []AttrSelector
	Pseudos    []PseudoSelector
	Combinator string // "", " ", ">", "+", "~"
}

// AttrSelector represents an attribute selector.
type AttrSelector struct {
	Name  string
	Op    string // "", "=", "~=", "|=", "^=", "$=", "*="
	Value string
}

// PseudoSelector represents a pseudo-class.
type PseudoSelector struct {
	Name string
	Arg  string
}

// ParseSelectorGroup parses a selector string into a selector group.
func ParseSelectorGroup(s string) (*SelectorGroup, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return &SelectorGroup{}, nil
	}

	var selectors []ComplexSelector
	for _, part := range splitByComma(s) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		sel, err := parseComplexSelector(part)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, sel)
	}
	return &SelectorGroup{Selectors: selectors}, nil
}

// MatchesSelector checks if a node matches the given selector string.
func MatchesSelector(n *Node, selector string) bool {
	if n == nil || n.Type != ElementNode {
		return false
	}
	group, err := ParseSelectorGroup(selector)
	if err != nil {
		return false
	}
	for _, sel := range group.Selectors {
		if matchesComplex(n, sel) {
			return true
		}
	}
	return false
}

// QueryAllSelector returns all descendant elements matching the selector.
func QueryAllSelector(root *Node, selector string) []*Node {
	group, err := ParseSelectorGroup(selector)
	if err != nil {
		return nil
	}
	var results []*Node
	queryAllSelectorRecursive(root, group, &results)
	return results
}

// QuerySelectorSelector returns the first descendant element matching the selector.
func QuerySelectorSelector(root *Node, selector string) *Node {
	all := QueryAllSelector(root, selector)
	if len(all) == 0 {
		return nil
	}
	return all[0]
}

func queryAllSelectorRecursive(n *Node, group *SelectorGroup, results *[]*Node) {
	for _, child := range n.Children {
		if child.Type != ElementNode {
			continue
		}
		for _, sel := range group.Selectors {
			if matchesComplex(child, sel) {
				*results = append(*results, child)
				break
			}
		}
		queryAllSelectorRecursive(child, group, results)
	}
}

func splitByComma(s string) []string {
	var parts []string
	var depth int
	var inQuote rune
	start := 0
	for i, r := range s {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		if r == '"' || r == '\'' {
			inQuote = r
			continue
		}
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func parseComplexSelector(s string) (ComplexSelector, error) {
	var parts []SelectorPart
	i := 0

	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		i++
	}
	if i >= len(s) {
		return ComplexSelector{}, nil
	}

	part, consumed, err := parseSimpleSelectorSequence(s[i:])
	if err != nil {
		return ComplexSelector{}, err
	}
	if consumed == 0 {
		return ComplexSelector{}, fmt.Errorf("invalid selector near %q", s[i:])
	}
	parts = append(parts, part)
	i += consumed

	for i < len(s) {
		hasSpace := false
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			hasSpace = true
			i++
		}
		if i >= len(s) {
			break
		}

		comb := ""
		if s[i] == '>' || s[i] == '+' || s[i] == '~' {
			comb = string(s[i])
			i++
			for i < len(s) && unicode.IsSpace(rune(s[i])) {
				i++
			}
		} else if hasSpace {
			comb = " "
		} else {
			return ComplexSelector{}, fmt.Errorf("invalid selector near %q", s[i:])
		}

		if i >= len(s) {
			return ComplexSelector{}, fmt.Errorf("missing selector after %q", comb)
		}
		nextPart, nextConsumed, err := parseSimpleSelectorSequence(s[i:])
		if err != nil {
			return ComplexSelector{}, err
		}
		if nextConsumed == 0 {
			return ComplexSelector{}, fmt.Errorf("invalid selector near %q", s[i:])
		}
		nextPart.Combinator = comb
		parts = append(parts, nextPart)
		i += nextConsumed
	}

	return ComplexSelector{Parts: parts}, nil
}

func parseSimpleSelectorSequence(s string) (SelectorPart, int, error) {
	var part SelectorPart
	i := 0
	first := true
	for i < len(s) {
		if unicode.IsSpace(rune(s[i])) {
			break
		}
		next, done, stillFirst, err := parseSimpleSelectorAtom(s, i, first, &part)
		if err != nil {
			return part, i, err
		}
		if done {
			return part, i, nil
		}
		i = next
		first = stillFirst
	}
	return part, i, nil
}

func parseSimpleSelectorAtom(s string, i int, first bool, part *SelectorPart) (int, bool, bool, error) {
	switch s[i] {
	case '*':
		if !first {
			return i, true, first, nil
		}
		part.Tag = "*"
		return i + 1, false, false, nil
	case '#':
		id, n := parseHash(s[i:])
		part.ID = id
		return i + n, false, false, nil
	case '.':
		class, n := parseClass(s[i:])
		part.Classes = append(part.Classes, class)
		return i + n, false, false, nil
	case '[':
		attr, n, err := parseAttribute(s[i:])
		if err != nil {
			return i, false, first, err
		}
		if n == 0 {
			return i, false, first, fmt.Errorf("invalid attribute selector near %q", s[i:])
		}
		part.Attributes = append(part.Attributes, attr)
		return i + n, false, false, nil
	case ':':
		pseudo, n, err := parsePseudo(s[i:])
		if err != nil {
			return i, false, first, err
		}
		part.Pseudos = append(part.Pseudos, pseudo)
		return i + n, false, false, nil
	case '>', '+', '~', ',':
		return i, true, first, nil
	default:
		if !first {
			return i, true, first, nil
		}
		tag, n := parseIdent(s[i:])
		if n == 0 {
			return i, true, first, nil
		}
		part.Tag = tag
		return i + n, false, false, nil
	}
}

func parseHash(s string) (string, int) {
	if len(s) == 0 || s[0] != '#' {
		return "", 0
	}
	ident, n := parseIdent(s[1:])
	return ident, n + 1
}

func parseClass(s string) (string, int) {
	if len(s) == 0 || s[0] != '.' {
		return "", 0
	}
	ident, n := parseIdent(s[1:])
	return ident, n + 1
}

func parseIdent(s string) (string, int) {
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			i++
		} else {
			break
		}
	}
	return s[:i], i
}

func skipSelectorSpace(s string, i int) int {
	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		i++
	}
	return i
}

func parseAttrOperator(s string, i int) (op string, next int, ok bool) {
	if i >= len(s) {
		return "", i, false
	}
	switch s[i] {
	case ']':
		return "", i, true
	case '=':
		return "=", i + 1, true
	case '~', '|', '^', '$', '*':
		if i+1 < len(s) && s[i+1] == '=' {
			return s[i : i+2], i + 2, true
		}
		return "", i, false
	default:
		return "", i, false
	}
}

func parseAttrValue(s string, i int) (string, int) {
	if i < len(s) && (s[i] == '"' || s[i] == '\'') {
		quote := s[i]
		i++
		start := i
		for i < len(s) && s[i] != quote {
			i++
		}
		val := s[start:i]
		if i < len(s) {
			i++
		}
		return val, i
	}
	val, n := parseIdent(s[i:])
	return val, i + n
}

func parseAttribute(s string) (AttrSelector, int, error) {
	var attr AttrSelector
	if len(s) == 0 || s[0] != '[' {
		return attr, 0, nil
	}
	i := skipSelectorSpace(s, 1)
	name, n := parseIdent(s[i:])
	if n == 0 {
		return attr, 0, nil
	}
	attr.Name = strings.ToLower(name)
	i = skipSelectorSpace(s, i+n)

	if i < len(s) {
		op, next, ok := parseAttrOperator(s, i)
		if !ok {
			return attr, i, nil
		}
		attr.Op = op
		i = next
		if op != "" {
			i = skipSelectorSpace(s, i)
			attr.Value, i = parseAttrValue(s, i)
		}
	}

	i = skipSelectorSpace(s, i)
	if i >= len(s) || s[i] != ']' {
		return attr, i, fmt.Errorf("unterminated attribute selector")
	}
	return attr, i + 1, nil
}

func parsePseudo(s string) (PseudoSelector, int, error) {
	var pseudo PseudoSelector
	if len(s) == 0 || s[0] != ':' {
		return pseudo, 0, nil
	}
	i := 1
	// handle :: pseudo-elements as single colons for matching
	if i < len(s) && s[i] == ':' {
		i++
	}
	name, n := parseIdent(s[i:])
	if n == 0 {
		return pseudo, i, nil
	}
	pseudo.Name = name
	i += n

	if i < len(s) && s[i] == '(' {
		i++ // skip (
		start := i
		depth := 1
		for i < len(s) && depth > 0 {
			switch s[i] {
			case '(':
				depth++
			case ')':
				depth--
			}
			if depth > 0 {
				i++
			} else {
				break
			}
		}
		pseudo.Arg = s[start:i]
		if i < len(s) && s[i] == ')' {
			i++
		}
	}
	return pseudo, i, nil
}
