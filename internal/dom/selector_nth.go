package dom

import (
	"strconv"
	"strings"
)

// childIndex returns the 1-based index of n among its parent's element children.
// If ofType is true, only counts children with the same tag.
func childIndex(n *Node, ofType bool) int {
	if n == nil || n.Parent == nil {
		return 0
	}
	idx := 0
	for _, child := range n.Parent.Children {
		if child.Type != ElementNode {
			continue
		}
		if !ofType || strings.EqualFold(child.TagName, n.TagName) {
			idx++
		}
		if child == n {
			return idx
		}
	}
	return 0
}

// lastChildIndex returns the 1-based index of n counting from the last element child.
func lastChildIndex(n *Node, ofType bool) int {
	if n == nil || n.Parent == nil {
		return 0
	}
	idx := 0
	for i := len(n.Parent.Children) - 1; i >= 0; i-- {
		child := n.Parent.Children[i]
		if child.Type != ElementNode {
			continue
		}
		if !ofType || strings.EqualFold(child.TagName, n.TagName) {
			idx++
		}
		if child == n {
			return idx
		}
	}
	return 0
}

func matchNth(arg string, idx int) bool {
	arg = strings.TrimSpace(strings.ToLower(arg))
	if arg == "" {
		return false
	}

	// Handle "0" as a special case - matches only the first element
	if arg == "0" {
		return idx == 0
	}

	// Handle keywords
	switch arg {
	case "even":
		return idx%2 == 0
	case "odd":
		return idx%2 == 1
	}

	// Parse an+b format
	a, b, ok := parseNth(arg)
	if !ok {
		// Try plain number
		n, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return idx == n
	}

	if a == 0 {
		return idx == b
	}
	if idx < b && a > 0 {
		return false
	}
	if idx > b && a < 0 {
		return false
	}
	return (idx-b)%a == 0
}

// parseNth parses strings like "2n+1", "3n", "-n+3", "n", "0".
func parseNth(s string) (a, b int, ok bool) {
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, 0, false
	}

	// Handle "0" specially - it's a valid formula with a=0, b=0
	if s == "0" {
		return 0, 0, true
	}

	// Find the position of 'n' (the variable part)
	nIdx := strings.Index(s, "n")
	if nIdx < 0 {
		return 0, 0, false
	}

	// Parse a
	aStr := s[:nIdx]
	switch aStr {
	case "":
		a = 1
	case "-":
		a = -1
	default:
		v, err := strconv.Atoi(aStr)
		if err != nil {
			return 0, 0, false
		}
		a = v
	}

	// Parse b
	bStr := s[nIdx+1:]
	if bStr == "" {
		b = 0
	} else {
		v, err := strconv.Atoi(bStr)
		if err != nil {
			return 0, 0, false
		}
		b = v
	}

	return a, b, true
}
