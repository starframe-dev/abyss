package js

import (
    "strings"

    "github.com/starframe/abyss/internal/dom"
)

// getParentNode returns the parent node of a node.
func getParentNode(id NodeID) *dom.Node {
    nodeMu.Lock()
    defer nodeMu.Unlock()

    n, ok := nodes[id]
    if !ok {
        return nil
    }
    return n.Parent
}

// getChildren returns the children of a node.
func getChildren(id NodeID) []*dom.Node {
    nodeMu.Lock()
    defer nodeMu.Unlock()

    n, ok := nodes[id]
    if !ok {
        return nil
    }
    return n.Children
}

// getElementsByTagName returns all elements with the given tag name.
func getElementsByTagName(tagName string) []*dom.Node {
    nodeMu.Lock()
    defer nodeMu.Unlock()

    tagName = strings.ToLower(tagName)
    var result []*dom.Node
    for _, n := range nodes {
        if n.Type == dom.ElementNode && strings.ToLower(n.TagName) == tagName {
            result = append(result, n)
        }
    }
    return result
}

// getElementsByClassName returns all elements with the given class name.
func getElementsByClassName(className string) []*dom.Node {
    nodeMu.Lock()
    defer nodeMu.Unlock()

    var result []*dom.Node
    for _, n := range nodes {
        if n.Type == dom.ElementNode {
            cls := n.GetAttr("class")
            if cls != "" {
                classes := strings.Fields(cls)
                for _, c := range classes {
                    if c == className {
                        result = append(result, n)
                        break
                    }
                }
            }
        }
    }
    return result
}

// querySelector returns the first element matching the CSS selector from the given node.
func querySelector(nodeID NodeID, selector string) *dom.Node {
    nodeMu.Lock()
    defer nodeMu.Unlock()

    n, ok := nodes[nodeID]
    if !ok || n == nil {
        return nil
    }

    return n.QuerySelector(selector)
}

// querySelectorAll returns all elements matching the CSS selector from the given node.
func querySelectorAll(nodeID NodeID, selector string) []*dom.Node {
    nodeMu.Lock()
    defer nodeMu.Unlock()

    n, ok := nodes[nodeID]
    if !ok || n == nil {
        return nil
    }

    return n.QueryAll(selector)
}

// matches checks if a node matches a CSS selector.
func matches(nodeID NodeID, selector string) bool {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return false
    }
    return n.Matches(selector)
}

// closest returns the closest ancestor matching the selector.
func closest(nodeID NodeID, selector string) *dom.Node {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return nil
    }
    return n.Closest(selector)
}

// cloneNode creates a deep or shallow clone of a node.
func cloneNode(nodeID NodeID, deep bool) *dom.Node {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return nil
    }

    clone := n.Clone(deep)
    if clone == nil {
        return nil
    }

    // Register the clone in the node map
    nodeMu.Lock()
    id := nodeNextID
    nodeNextID++
    nodes[id] = clone
    nodeMu.Unlock()

    return clone
}

// contains checks if a node contains another node (including self).
func contains(parentID, childID NodeID) bool {
    if parentID == childID {
        return true
    }
    nodeMu.Lock()
    parent, ok := nodes[parentID]
    child, ok2 := nodes[childID]
    nodeMu.Unlock()
    if !ok || !ok2 || parent == nil || child == nil {
        return false
    }
    for cur := child; cur != nil; cur = cur.Parent {
        if cur == parent {
            return true
        }
    }
    return false
}

// previousElementSibling returns the previous sibling that is an element node.
func previousElementSibling(nodeID NodeID) *dom.Node {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil || n.Parent == nil {
        return nil
    }
    parent := n.Parent
    found := false
    for i := len(parent.Children) - 1; i >= 0; i-- {
        child := parent.Children[i]
        if child == n {
            found = true
            continue
        }
        if found && child.Type == dom.ElementNode {
            return child
        }
    }
    return nil
}

// nextElementSibling returns the next sibling that is an element node.
func nextElementSibling(nodeID NodeID) *dom.Node {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil || n.Parent == nil {
        return nil
    }
    parent := n.Parent
    found := false
    for _, child := range parent.Children {
        if child == n {
            found = true
            continue
        }
        if found && child.Type == dom.ElementNode {
            return child
        }
    }
    return nil
}

// firstElementChild returns the first child that is an element node.
func firstElementChild(nodeID NodeID) *dom.Node {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return nil
    }
    for _, child := range n.Children {
        if child.Type == dom.ElementNode {
            return child
        }
    }
    return nil
}

// lastElementChild returns the last child that is an element node.
func lastElementChild(nodeID NodeID) *dom.Node {
    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return nil
    }
    for i := len(n.Children) - 1; i >= 0; i-- {
        child := n.Children[i]
        if child.Type == dom.ElementNode {
            return child
        }
    }
    return nil
}
