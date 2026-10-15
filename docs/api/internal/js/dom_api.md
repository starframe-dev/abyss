# `internal/js` DOM API

`dom_api.go` exposes Go-callable DOM helpers from package `js`. The functions are thin wrappers around the package's canonical bridge implementations and operate on `*internal/dom.Node` values or `NodeID` handles. The file specifically describes these helpers as public Go DOM helpers used by integration tests; it does not define a separate DOM implementation.

## Types and handles

- **`NodeID`** is the package-level identifier used to refer to registered nodes. Its underlying representation and allocation rules are defined elsewhere in package `js`.
- **`*dom.Node`** is the node type returned by constructors and lookup/traversal operations.
- **`GetNodeByID`** returns both the node and an `ok` flag, so callers can distinguish an absent ID from a valid mapping. Other wrappers expose the behavior of their corresponding canonical bridge functions; this file does not add error results or validation.

## Node creation and lookup

| Function | Behavior |
| --- | --- |
| `CreateElement(tagName string) *dom.Node` | Creates an element with the supplied tag name. |
| `CreateTextNode(text string) *dom.Node` | Creates a text node containing the supplied text. |
| `CreateDocumentFragment() *dom.Node` | Creates a document-fragment node. |
| `GetNodeID(n *dom.Node) NodeID` | Returns the node's identifier. |
| `GetNodeByID(id NodeID) (*dom.Node, bool)` | Looks up an identifier in the package node registry. |
| `GetBody() *dom.Node` | Returns the document body element. |
| `GetElementByID(id string) *dom.Node` | Finds an element by its `id` attribute. |

`GetNodeByID` acquires `nodeMu` while reading the shared `nodes` registry and releases it before returning. A missing registry entry returns the zero node pointer and `false`.

## Tree mutation

| Function | Behavior |
| --- | --- |
| `AppendChild(parentID, childID NodeID)` | Adds the identified child to the identified parent. |
| `InsertBefore(parentID, newNodeID, refNodeID NodeID)` | Inserts a node before a reference node. |
| `ReplaceChild(parentID, newNodeID, oldNodeID NodeID)` | Replaces a child with another node. |
| `RemoveChild(parentID, childID NodeID)` | Removes a child from its parent. |
| `CloneNode(id NodeID, deep bool) *dom.Node` | Clones a node; `deep` selects whether its descendants are cloned. |

These functions delegate directly to the bridge helpers. Their signatures do not expose errors or mutation status; callers should consult the canonical helper behavior for handling invalid IDs or tree constraints.

## Content and attributes

| Function | Behavior |
| --- | --- |
| `SetInnerHTML(id NodeID, html string)` | Sets a node's inner HTML. |
| `GetInnerHTML(id NodeID) string` | Returns a node's inner HTML. |
| `SetAttribute(id NodeID, name, value string)` | Sets the named attribute. |
| `GetAttribute(id NodeID, name string) string` | Returns the named attribute value. |
| `SetTextContent(id NodeID, text string)` | Sets a node's text content. |
| `GetTextContent(id NodeID) string` | Returns a node's text content. |

The wrapper layer does not document or alter parsing, serialization, absent-attribute, or invalid-node semantics; those are determined by the delegated implementations.

## Tree inspection and selectors

| Function | Behavior |
| --- | --- |
| `Contains(id, otherID NodeID) bool` | Reports whether the node identified by `id` contains the other node. |
| `FirstElementChild(id NodeID) *dom.Node` | Returns the first element child. |
| `LastElementChild(id NodeID) *dom.Node` | Returns the last element child. |
| `QuerySelector(id NodeID, selector string) *dom.Node` | Returns the first matching element in the relevant search scope. |
| `QuerySelectorAll(id NodeID, selector string) []*dom.Node` | Returns all matching elements in the relevant search scope. |
| `Matches(id NodeID, selector string) bool` | Reports whether the identified element matches the selector. |
| `Closest(id NodeID, selector string) *dom.Node` | Returns the closest matching ancestor according to the delegated implementation. |

Selector syntax, search-scope details, ordering, and behavior for unmatched selectors are not implemented in this file and follow the canonical bridge functions.

## Implementation notes

All functions except `GetNodeByID` are direct delegates and add no synchronization, conversion, validation, or error handling of their own. `GetNodeByID` is the exception: it protects registry access with the package mutex. The API is therefore a convenience surface over package internals rather than an independent abstraction with additional guarantees.
