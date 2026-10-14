package js

import "github.com/starframe-dev/abyss/internal/dom"

// Public Go DOM helpers used by integration tests. Thin wrappers over the
// unexported bridge implementations in this package.

// CreateElement creates a new element node.
func CreateElement(tagName string) *dom.Node {
	return createElement(tagName)
}

// CreateTextNode creates a new text node.
func CreateTextNode(text string) *dom.Node {
	return createTextNode(text)
}

// CreateDocumentFragment creates a new document fragment node.
func CreateDocumentFragment() *dom.Node {
	return createDocumentFragment()
}

// GetNodeID returns the ID of a node.
func GetNodeID(n *dom.Node) NodeID {
	return getNodeID(n)
}

// GetNodeByID returns a node by its ID.
func GetNodeByID(id NodeID) (*dom.Node, bool) {
	nodeMu.Lock()
	defer nodeMu.Unlock()
	node, ok := nodes[id]
	return node, ok
}

// AppendChild adds a child to a parent node.
func AppendChild(parentID, childID NodeID) {
	appendChild(parentID, childID)
}

// InsertBefore inserts a new node before a reference node.
func InsertBefore(parentID, newNodeID, refNodeID NodeID) {
	insertBefore(parentID, newNodeID, refNodeID)
}

// ReplaceChild replaces a child node with a new one.
func ReplaceChild(parentID, newNodeID, oldNodeID NodeID) {
	replaceChild(parentID, newNodeID, oldNodeID)
}

// RemoveChild removes a child from a parent node.
func RemoveChild(parentID, childID NodeID) {
	removeChild(parentID, childID)
}

// SetInnerHTML sets the inner HTML of a node.
func SetInnerHTML(id NodeID, html string) {
	setInnerHTML(id, html)
}

// GetInnerHTML returns the inner HTML of a node.
func GetInnerHTML(id NodeID) string {
	return getInnerHTML(id)
}

// SetAttribute sets an attribute on a node.
func SetAttribute(id NodeID, name, value string) {
	setAttribute(id, name, value)
}

// GetAttribute returns an attribute value from a node.
func GetAttribute(id NodeID, name string) string {
	return getAttribute(id, name)
}

// SetTextContent sets the text content of a node.
func SetTextContent(id NodeID, text string) {
	setTextContent(id, text)
}

// GetTextContent returns the text content of a node.
func GetTextContent(id NodeID) string {
	return getTextContent(id)
}

// GetBody returns the body element of the document.
func GetBody() *dom.Node {
	return getBody()
}

// GetElementByID returns an element by its id attribute.
func GetElementByID(id string) *dom.Node {
	return getElementByID(id)
}

// CloneNode clones a node with or without children.
func CloneNode(id NodeID, deep bool) *dom.Node {
	return cloneNode(id, deep)
}

// Contains reports whether node id contains otherID.
func Contains(id NodeID, otherID NodeID) bool {
	return contains(id, otherID)
}

// FirstElementChild returns the first element child.
func FirstElementChild(id NodeID) *dom.Node {
	return firstElementChild(id)
}

// LastElementChild returns the last element child.
func LastElementChild(id NodeID) *dom.Node {
	return lastElementChild(id)
}

// QuerySelector returns the first element matching the selector.
func QuerySelector(id NodeID, selector string) *dom.Node {
	return querySelector(id, selector)
}

// QuerySelectorAll returns all elements matching the selector.
func QuerySelectorAll(id NodeID, selector string) []*dom.Node {
	return querySelectorAll(id, selector)
}

// Matches reports whether the element matches the selector.
func Matches(id NodeID, selector string) bool {
	return matches(id, selector)
}

// Closest returns the closest ancestor matching the selector.
func Closest(id NodeID, selector string) *dom.Node {
	return closest(id, selector)
}
