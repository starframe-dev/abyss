package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
*/
import "C"
import (
	"strings"
	"sync"
	"time"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// NodeID identifies a DOM node in the JavaScript bridge.
type NodeID int

// Rect represents a layout rectangle.
type Rect struct {
	X, Y, Width, Height float64
}

var (
	nodeMu     sync.Mutex
	nodeNextID NodeID
	nodes      = map[NodeID]*dom.Node{}
	nodeIDs    = map[*dom.Node]NodeID{}
	rootDoc    *dom.Node

	// Computed styles for getComputedStyle bridge
	computedStyles   map[*dom.Node]css.ComputedStyle
	computedStylesMu sync.Mutex

	// Layout boxes for getBoundingClientRect bridge
	layoutBoxes   map[*dom.Node]Rect
	layoutBoxesMu sync.Mutex

	// Event listeners: nodeID -> eventType -> callbacks
	eventMu        sync.Mutex
	eventListeners = map[NodeID]map[string][]C.JSObjectRef{}
	eventCtx       C.JSGlobalContextRef

	// Current JSC context and URL for live re-rendering
	currentCtxMu sync.Mutex
	currentCtx   *Context
	currentURL   string

	// Dirty flag: true when DOM was modified and re-render is needed
	dirtyMu sync.Mutex
	dirty   bool

	// Timers for setTimeout/setInterval
	timerMu     sync.Mutex
	timers      []timer
	nextTimerID int
)

// timer represents a scheduled callback.
type timer struct {
	ID       int
	Callback C.JSObjectRef
	Deadline time.Time
	Interval time.Duration
	Repeat   bool
}

// SetCurrentContext stores the active JSC context and URL for live re-rendering.
func SetCurrentContext(ctx *Context, url string) {
	currentCtxMu.Lock()
	defer currentCtxMu.Unlock()
	if currentCtx != nil && currentCtx != ctx {
		resetEventSystem()
		currentCtx.Close()
	}
	currentCtx = ctx
	currentURL = url
}

// GetCurrentContext returns the active JSC context and URL.
func GetCurrentContext() (*Context, string) {
	currentCtxMu.Lock()
	defer currentCtxMu.Unlock()
	return currentCtx, currentURL
}

// CloseCurrentContext closes the active JSC context.
func CloseCurrentContext() {
	currentCtxMu.Lock()
	if currentCtx != nil {
		resetEventSystem()
		currentCtx.Close()
		currentCtx = nil
		currentURL = ""
	}
	currentCtxMu.Unlock()
	resetDOMState()
}

// SetDirty marks the DOM as modified so the browser can re-render.
func SetDirty() {
	dirtyMu.Lock()
	dirty = true
	dirtyMu.Unlock()
}

// IsDirty reports whether the DOM has been modified since the last render.
func IsDirty() bool {
	dirtyMu.Lock()
	defer dirtyMu.Unlock()
	return dirty
}

// MarkClean clears the dirty flag after a render.
func MarkClean() {
	dirtyMu.Lock()
	dirty = false
	dirtyMu.Unlock()
}

// resetDOMState releases references to the current page DOM and layout state.
func resetDOMState() {
	computedStylesMu.Lock()
	computedStyles = nil
	computedStylesMu.Unlock()
	layoutBoxesMu.Lock()
	layoutBoxes = nil
	layoutBoxesMu.Unlock()
	nodeMu.Lock()
	nodes = nil
	nodeIDs = nil
	nodeNextID = 0
	rootDoc = nil
	nodeMu.Unlock()
	dirtyMu.Lock()
	dirty = false
	dirtyMu.Unlock()
}

// InitDOM initializes the DOM state with a parsed document.
func InitDOM(doc *dom.Node) {
	resetDOMState()
	nodeMu.Lock()
	defer nodeMu.Unlock()

	nodes = map[NodeID]*dom.Node{}
	nodeIDs = map[*dom.Node]NodeID{}
	nodeNextID = 1
	rootDoc = doc

	// Assign IDs to all nodes
	assignIDs(doc)
}

// assignIDs recursively assigns IDs to all nodes.
func registerNodeLocked(n *dom.Node) (NodeID, bool) {
	if n == nil || len(nodes) >= dom.MaxDOMNodes {
		return -1, false
	}
	id := nodeNextID
	nodeNextID++
	nodes[id] = n
	nodeIDs[n] = id
	return id, true
}

func assignIDs(n *dom.Node) {
	if _, ok := registerNodeLocked(n); !ok {
		return
	}
	for _, child := range n.Children {
		assignIDs(child)
	}
}

// createElement creates a new element node.
func createElement(tagName string) *dom.Node {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	n := &dom.Node{
		Type:       dom.ElementNode,
		TagName:    strings.ToLower(tagName),
		Attributes: map[string]string{},
		Children:   []*dom.Node{},
	}
	if _, ok := registerNodeLocked(n); !ok {
		return nil
	}
	return n
}

// createTextNode creates a new text node.
func createTextNode(text string) *dom.Node {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	n := &dom.Node{Type: dom.TextNode, Text: text}
	if _, ok := registerNodeLocked(n); !ok {
		return nil
	}
	return n
}

// createDocumentFragment creates a new document fragment node.
func createDocumentFragment() *dom.Node {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	n := &dom.Node{
		Type:       dom.DocumentFragmentNode,
		TagName:    "#document-fragment",
		Attributes: map[string]string{},
		Children:   []*dom.Node{},
	}
	if _, ok := registerNodeLocked(n); !ok {
		return nil
	}
	return n
}

// removeNodeFromParentLocked removes a node from its current parent.
// Caller must hold nodeMu.
func removeNodeFromParentLocked(child *dom.Node) {
	if child.Parent == nil {
		return
	}
	oldParent := child.Parent
	for i, c := range oldParent.Children {
		if c == child {
			oldParent.Children = append(oldParent.Children[:i], oldParent.Children[i+1:]...)
			break
		}
	}
}

// moveAllChildren moves all children from source to parent.
// Caller must hold nodeMu.
func moveAllChildren(parent, source *dom.Node) {
	children := make([]*dom.Node, len(source.Children))
	copy(children, source.Children)
	for _, child := range children {
		child.Parent = nil
		child.Parent = parent
		parent.Children = append(parent.Children, child)
	}
	source.Children = nil
}

// appendChild adds a child to a parent node. Document fragments are unwrapped.
func wouldCreateCycle(parent, child *dom.Node) bool {
	for current := parent; current != nil; current = current.Parent {
		if current == child {
			return true
		}
	}
	return false
}

func appendChild(parentID, childID NodeID) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	parent, ok := nodes[parentID]
	if !ok {
		return
	}
	child, ok := nodes[childID]
	if !ok || parent == child {
		return
	}

	if child.Type == dom.DocumentFragmentNode {
		for _, fragmentChild := range child.Children {
			if wouldCreateCycle(parent, fragmentChild) {
				return
			}
		}
		moveAllChildren(parent, child)
		return
	}
	if wouldCreateCycle(parent, child) {
		return
	}

	removeNodeFromParentLocked(child)
	child.Parent = parent
	parent.Children = append(parent.Children, child)
}

// setInnerHTML sets the inner HTML of a node.
func setInnerHTML(id NodeID, html string) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	node, ok := nodes[id]
	if !ok {
		return
	}

	// Parse innerHTML as an unwrapped fragment so direct children are inserted.
	fragment := dom.ParseFragment(html)
	children := fragment.Children

	// Clear existing children
	for _, child := range node.Children {
		child.Parent = nil
	}
	node.Children = nil

	// Add new children
	for _, child := range children {
		child.Parent = node
		node.Children = append(node.Children, child)
		assignIDsToNew(child)
	}
}

// assignIDsToNew assigns IDs to nodes that don't have one yet.
func assignIDsToNew(n *dom.Node) {
	if n == nil {
		return
	}
	if _, found := nodeIDs[n]; !found {
		if _, ok := registerNodeLocked(n); !ok {
			return
		}
	}
	for _, child := range n.Children {
		assignIDsToNew(child)
	}
}

// setAttribute sets an attribute on a node.
func setAttribute(id NodeID, name, value string) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	node, ok := nodes[id]
	if !ok {
		return
	}
	if node.Attributes == nil {
		node.Attributes = map[string]string{}
	}
	node.Attributes[strings.ToLower(name)] = value
}

// setTextContent sets the text content of a node.
func setTextContent(id NodeID, text string) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	node, ok := nodes[id]
	if !ok || len(nodes) >= dom.MaxDOMNodes {
		return
	}

	// Clear existing children
	for _, child := range node.Children {
		child.Parent = nil
	}
	node.Children = nil

	// Create a text node
	tn := &dom.Node{Type: dom.TextNode, Text: text, Parent: node}
	if _, ok := registerNodeLocked(tn); !ok {
		return
	}
	node.Children = append(node.Children, tn)
}

// getBody returns the body element of the document.
func getBody() *dom.Node {
	if rootDoc == nil {
		return nil
	}
	for _, child := range rootDoc.Children {
		if child.Type == dom.ElementNode && child.TagName == "html" {
			for _, gc := range child.Children {
				if gc.Type == dom.ElementNode && gc.TagName == "body" {
					return gc
				}
			}
		}
	}
	return nil
}

// getNodeID returns the ID of a node.
func getNodeID(n *dom.Node) NodeID {
	nodeMu.Lock()
	defer nodeMu.Unlock()
	if id, ok := nodeIDs[n]; ok {
		return id
	}
	return -1
}

// getElementByID returns an element by its id attribute.
func getElementByID(id string) *dom.Node {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	var find func(*dom.Node) *dom.Node
	find = func(node *dom.Node) *dom.Node {
		if node == nil {
			return nil
		}
		if node.Type == dom.ElementNode && node.GetAttr("id") == id {
			return node
		}
		for _, child := range node.Children {
			if match := find(child); match != nil {
				return match
			}
		}
		return nil
	}
	return find(rootDoc)
}

// getAttribute returns an attribute value from a node.
func getAttribute(id NodeID, name string) string {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	n, ok := nodes[id]
	if !ok {
		return ""
	}
	return n.GetAttr(name)
}

// removeChild removes a child from a parent node.
func removeChild(parentID, childID NodeID) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	parent, ok := nodes[parentID]
	if !ok {
		return
	}
	child, ok := nodes[childID]
	if !ok {
		return
	}

	for i, c := range parent.Children {
		if c == child {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			child.Parent = nil
			return
		}
	}
}
