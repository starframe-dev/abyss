# DOM Queries and Traversal

This file documents the internal DOM query and traversal helpers implemented in `internal/js/dom_query.go`. The functions are unexported Go package helpers, rather than a public Go API. They operate on the package's node registry (`nodes`), identified by `NodeID`, and return `*dom.Node` values directly.

## API overview

| Helper | Behavior |
| --- | --- |
| `getParentNode(id NodeID) *dom.Node` | Returns the node's parent, or `nil` if the ID is unknown. |
| `getChildren(id NodeID) []*dom.Node` | Returns the node's child slice, or `nil` if the ID is unknown. |
| `getElementsByTagName(tagName string) []*dom.Node` | Finds registered element nodes with a case-insensitive tag-name match. |
| `getElementsByClassName(className string) []*dom.Node` | Finds registered element nodes whose whitespace-separated `class` attribute contains an exact token match. |
| `querySelector(nodeID NodeID, selector string) *dom.Node` | Returns the first selector match within the registered node, or `nil` if the ID is invalid or there is no match. |
| `querySelectorAll(nodeID NodeID, selector string) []*dom.Node` | Returns all selector matches within the registered node, or `nil` if the ID is invalid. |
| `matches(nodeID NodeID, selector string) bool` | Reports whether the registered node matches the selector; invalid IDs return `false`. |
| `closest(nodeID NodeID, selector string) *dom.Node` | Returns the closest matching node as defined by `dom.Node.Closest`; invalid IDs return `nil`. |
| `cloneNode(nodeID NodeID, deep bool) *dom.Node` | Clones a registered node, assigning IDs to the clone tree; returns `nil` on invalid input or cloning/registration limits. |
| `contains(parentID, childID NodeID) bool` | Reports whether the first node contains the second in its parent chain; a node contains itself. |
| `previousElementSibling(nodeID NodeID) *dom.Node` | Returns the nearest preceding element sibling, skipping non-element nodes. |
| `nextElementSibling(nodeID NodeID) *dom.Node` | Returns the nearest following element sibling, skipping non-element nodes. |
| `firstElementChild(nodeID NodeID) *dom.Node` | Returns the first element child, skipping non-element children. |
| `lastElementChild(nodeID NodeID) *dom.Node` | Returns the last element child, skipping non-element children. |

## Behavior and details

### Registry-wide lookups

`getElementsByTagName` and `getElementsByClassName` scan the `nodes` registry. Their results are not limited to a particular root or subtree. Only nodes whose type is `dom.ElementNode` are considered. Tag names are compared after lowercasing both the requested name and each node's tag name. Class values are split with `strings.Fields`, so any run of whitespace separates tokens; matching is exact and case-sensitive.

These functions append matches in the iteration order of the Go map, which is unspecified. Callers must not depend on a stable result order.

### Selector operations

`querySelector` and `querySelectorAll` delegate selector interpretation and tree traversal to `dom.Node.QuerySelector` and `dom.Node.QueryAll`, respectively. A missing or nil registry entry produces `nil` without delegation. `matches` and `closest` similarly delegate to `dom.Node.Matches` and `dom.Node.Closest` after retrieving the node. The selector syntax and behavior are therefore determined by the `internal/dom` implementation, not by this file.

### Cloning and identity

`cloneNode` rejects unknown or nil nodes and refuses to start when the registry already contains at least `dom.MaxDOMNodes` entries. It calls `n.Clone(deep)` and, on success, uses `assignIDsToNew` to register IDs for the cloned node (and, according to that helper's behavior, its descendants). It returns `nil` if cloning fails or if the resulting clone does not have an entry in `nodeIDs`.

### Ancestry and sibling traversal

`contains` treats identical IDs as containment before checking the registry. Otherwise it requires both IDs to resolve to non-nil nodes, then follows the candidate child's `Parent` pointers until it reaches the proposed parent or the chain ends.

Sibling helpers inspect the node's parent's `Children` slice and return only nodes of type `dom.ElementNode`. First/last element-child helpers scan the node's own children in forward/reverse order. If the node, parent, or suitable related node is absent, the result is `nil`.

## Synchronization and ownership

Registry reads and scans are protected with `nodeMu`. Several helpers release the mutex before traversing node relationships or invoking DOM methods (`matches`, `closest`, `contains`, and sibling/element-child helpers). Returned pointers and slices refer to the underlying DOM objects; this file does not copy them. Callers should account for the package's synchronization and mutation conventions when using those references.
