package js

import (
	"sort"
	"strings"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// getInnerHTML returns the inner HTML of a node.
func getInnerHTML(id NodeID) string {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	n, ok := nodes[id]
	if !ok {
		return ""
	}
	var b strings.Builder
	for _, child := range n.Children {
		serializeNode(&b, child)
	}
	return b.String()
}

// getTextContent returns the text content of a node.
func getTextContent(id NodeID) string {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	n, ok := nodes[id]
	if !ok {
		return ""
	}
	return collectText(n)
}

// collectText recursively collects text from a node and its children.
func collectText(n *dom.Node) string {
	if n.Type == dom.TextNode {
		return n.Text
	}
	var b strings.Builder
	for _, child := range n.Children {
		b.WriteString(collectText(child))
	}
	return b.String()
}

// serializeNode serializes a DOM node to HTML.
func serializeNode(b *strings.Builder, n *dom.Node) {
	switch n.Type {
	case dom.TextNode:
		b.WriteString(escapeHTML(n.Text))
	case dom.ElementNode:
		b.WriteString("<")
		b.WriteString(n.TagName)
		// Sort attributes for deterministic output
		attrKeys := make([]string, 0, len(n.Attributes))
		for k := range n.Attributes {
			attrKeys = append(attrKeys, k)
		}
		sort.Strings(attrKeys)
		for _, k := range attrKeys {
			v := n.Attributes[k]
			b.WriteString(" ")
			b.WriteString(k)
			if v != "" {
				b.WriteString("=\"")
				b.WriteString(escapeHTML(v))
				b.WriteString("\"")
			}
		}
		if len(n.Children) == 0 && voidElements[n.TagName] {
			b.WriteString(" />")
			return
		}
		b.WriteString(">")
		for _, child := range n.Children {
			serializeNode(b, child)
		}
		b.WriteString("</")
		b.WriteString(n.TagName)
		b.WriteString(">")
	}
}

// escapeHTML escapes special HTML characters.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// voidElements are HTML elements that don't need closing tags.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

// outerHTMLString returns the HTML serialization of a node.
func outerHTMLString(nodeID NodeID) string {
	nodeMu.Lock()
	n, ok := nodes[nodeID]
	nodeMu.Unlock()
	if !ok || n == nil {
		return ""
	}
	return n.OuterHTML()
}

// setOuterHTML replaces a node with parsed HTML.
func setOuterHTML(nodeID NodeID, html string) {
	nodeMu.Lock()
	n, ok := nodes[nodeID]
	nodeMu.Unlock()
	if !ok || n == nil || n.Parent == nil {
		return
	}

	parent := n.Parent
	fragment := dom.Parse("<fragment>" + html + "</fragment>")
	if fragment == nil {
		return
	}

	var newNodes []*dom.Node
	for _, child := range fragment.Children {
		if child.Type == dom.ElementNode {
			newNodes = append(newNodes, child)
		}
	}
	if len(newNodes) == 0 {
		return
	}

	nodeMu.Lock()
	defer nodeMu.Unlock()

	// Find index of n in parent
	idx := -1
	for i, child := range parent.Children {
		if child == n {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	// Remove n
	n.Parent = nil
	parent.Children = append(parent.Children[:idx], parent.Children[idx+1:]...)

	// Insert new nodes at idx
	for _, newNode := range newNodes {
		newNode.Parent = parent
	}
	before := parent.Children[:idx]
	after := parent.Children[idx:]
	parent.Children = append(before, append(newNodes, after...)...)

	css.InvalidateCache()
}
