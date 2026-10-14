package js

import (
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// insertBefore inserts a new node before a reference node. Document fragments are unwrapped.
func insertBefore(parentID, newNodeID, refNodeID NodeID) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	parent, ok := nodes[parentID]
	if !ok || parent == nil {
		return
	}
	newNode, ok := nodes[newNodeID]
	if !ok || newNode == nil || newNodeID == refNodeID {
		return
	}

	// Find the reference node
	var refNode *dom.Node
	if refNodeID >= 0 {
		refNode, _ = nodes[refNodeID]
	}

	if newNode.Type == dom.DocumentFragmentNode {
		for _, child := range newNode.Children {
			if wouldCreateCycle(parent, child) {
				return
			}
		}
		insertFragmentChildrenLocked(parent, newNode, refNode)
		css.InvalidateCache()
		return
	}
	if wouldCreateCycle(parent, newNode) {
		return
	}

	// Remove newNode from its current parent if any
	if newNode.Parent != nil {
		newNode.Parent.RemoveChild(newNode)
	}

	// Insert before refNode or append
	newNode.Parent = parent
	if refNode != nil {
		// Find index of refNode in parent's children
		idx := -1
		for i, child := range parent.Children {
			if child == refNode {
				idx = i
				break
			}
		}
		if idx >= 0 {
			// Insert at idx
			parent.Children = append(parent.Children[:idx], append([]*dom.Node{newNode}, parent.Children[idx:]...)...)
		} else {
			parent.Children = append(parent.Children, newNode)
		}
	} else {
		parent.Children = append(parent.Children, newNode)
	}

	// Invalidate CSS cache
	css.InvalidateCache()
}

// insertFragmentChildrenLocked inserts all children of a fragment before refNode.
// Caller must hold nodeMu.
func insertFragmentChildrenLocked(parent, fragment, refNode *dom.Node) {
	children := make([]*dom.Node, len(fragment.Children))
	copy(children, fragment.Children)
	if refNode != nil {
		idx := -1
		for i, child := range parent.Children {
			if child == refNode {
				idx = i
				break
			}
		}
		if idx >= 0 {
			before := parent.Children[:idx]
			after := parent.Children[idx:]
			for _, child := range children {
				child.Parent = parent
			}
			parent.Children = append(before, append(children, after...)...)
			fragment.Children = nil
			return
		}
	}
	for _, child := range children {
		child.Parent = parent
		parent.Children = append(parent.Children, child)
	}
	fragment.Children = nil
}

// replaceChild replaces a child node with a new one. Document fragments are unwrapped.
func replaceChild(parentID, newNodeID, oldNodeID NodeID) {
	nodeMu.Lock()
	defer nodeMu.Unlock()

	parent, ok := nodes[parentID]
	if !ok || parent == nil {
		return
	}
	newNode, ok := nodes[newNodeID]
	if !ok || newNode == nil || parent == newNode {
		return
	}
	oldNode, ok := nodes[oldNodeID]
	if !ok || oldNode == nil || oldNode == newNode {
		return
	}

	if newNode.Type == dom.DocumentFragmentNode {
		for _, child := range newNode.Children {
			if wouldCreateCycle(parent, child) {
				return
			}
		}
		idx := -1
		for i, child := range parent.Children {
			if child == oldNode {
				idx = i
				break
			}
		}
		if idx < 0 {
			return
		}
		oldNode.Parent = nil
		children := make([]*dom.Node, len(newNode.Children))
		copy(children, newNode.Children)
		for _, child := range children {
			child.Parent = parent
		}
		before := parent.Children[:idx]
		after := parent.Children[idx+1:]
		parent.Children = append(before, append(children, after...)...)
		newNode.Children = nil
		css.InvalidateCache()
		return
	}
	if wouldCreateCycle(parent, newNode) {
		return
	}

	// Remove newNode from its current parent if any
	if newNode.Parent != nil {
		newNode.Parent.RemoveChild(newNode)
	}

	// Find and replace oldNode
	for i, child := range parent.Children {
		if child == oldNode {
			newNode.Parent = parent
			oldNode.Parent = nil
			parent.Children[i] = newNode
			break
		}
	}

	// Invalidate CSS cache
	css.InvalidateCache()
}

// InsertAdjacentHTML parses HTML and inserts it at the specified position (public API).
// Positions: beforebegin, afterbegin, beforeend, afterend
func InsertAdjacentHTML(nodeID NodeID, position, html string) {
	nodeMu.Lock()
	n, ok := nodes[nodeID]
	nodeMu.Unlock()
	if !ok || n == nil {
		return
	}

	// Parse the HTML as an unwrapped fragment so the insertion does not add a
	// synthetic <fragment> node and preserves top-level text nodes.
	fragment := dom.ParseFragment(html)
	if fragment == nil {
		return
	}
	newNodes := fragment.Children
	if len(newNodes) == 0 {
		return
	}

	nodeMu.Lock()
	defer nodeMu.Unlock()
	for _, newNode := range newNodes {
		assignIDsToNew(newNode)
	}

	switch position {
	case "beforebegin":
		insertSiblingNodes(n, newNodes, false)
	case "afterbegin":
		for _, newNode := range newNodes {
			newNode.Parent = n
		}
		n.Children = append(newNodes, n.Children...)
	case "beforeend":
		for _, newNode := range newNodes {
			newNode.Parent = n
		}
		n.Children = append(n.Children, newNodes...)
	case "afterend":
		insertSiblingNodes(n, newNodes, true)
	}

	// Invalidate CSS cache
	css.InvalidateCache()
}

// insertSiblingNodes inserts nodes before (after=false) or after (after=true)
// an existing node in its parent's children.
func insertSiblingNodes(n *dom.Node, newNodes []*dom.Node, after bool) {
	if n.Parent == nil {
		return
	}
	parent := n.Parent
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
	for _, newNode := range newNodes {
		newNode.Parent = parent
	}
	if after {
		idx++
	}
	before := parent.Children[:idx]
	rest := parent.Children[idx:]
	parent.Children = append(before, append(newNodes, rest...)...)
}

// insertAdjacentText creates a text node and inserts it at the specified position.
func insertAdjacentText(nodeID NodeID, position, text string) {
	nodeMu.Lock()
	n, ok := nodes[nodeID]
	nodeMu.Unlock()
	if !ok || n == nil {
		return
	}

	textNode := &dom.Node{Type: dom.TextNode, Text: text}

	nodeMu.Lock()
	defer nodeMu.Unlock()

	switch position {
	case "beforebegin":
		insertSiblingNodes(n, []*dom.Node{textNode}, false)
	case "afterbegin":
		textNode.Parent = n
		n.Children = append([]*dom.Node{textNode}, n.Children...)
	case "beforeend":
		textNode.Parent = n
		n.Children = append(n.Children, textNode)
	case "afterend":
		insertSiblingNodes(n, []*dom.Node{textNode}, true)
	}
}

// insertAdjacentElement inserts an existing element at the specified position.
func insertAdjacentElement(nodeID, elemID NodeID, position string) {
	nodeMu.Lock()
	defer nodeMu.Unlock()
	n, ok := nodes[nodeID]
	elem, ok2 := nodes[elemID]
	if !ok || !ok2 || n == nil || elem == nil {
		return
	}

	targetParent := n
	if position == "beforebegin" || position == "afterend" {
		targetParent = n.Parent
		if targetParent == nil {
			return
		}
	} else if position != "afterbegin" && position != "beforeend" {
		return
	}
	if wouldCreateCycle(targetParent, elem) {
		return
	}
	removeNodeFromParentLocked(elem)

	switch position {
	case "beforebegin":
		insertSiblingNodes(n, []*dom.Node{elem}, false)
	case "afterbegin":
		elem.Parent = n
		n.Children = append([]*dom.Node{elem}, n.Children...)
	case "beforeend":
		elem.Parent = n
		n.Children = append(n.Children, elem)
	case "afterend":
		insertSiblingNodes(n, []*dom.Node{elem}, true)
	}
	css.InvalidateCache()
}
