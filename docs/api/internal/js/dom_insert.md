# DOM insertion

This module implements DOM insertion operations over the node registry in `internal/js`. Mutations are guarded by `nodeMu`; successful operations that affect CSS-relevant structure invalidate the CSS cache.

## Public API

### `InsertAdjacentHTML(nodeID NodeID, position, html string)`

Parses `html` with `dom.ParseFragment` and inserts the resulting top-level nodes relative to the registered node. The fragment wrapper is not inserted, so top-level text nodes are preserved as ordinary siblings or children. Newly inserted nodes are assigned registry IDs with `assignIDsToNew`.

Supported positions:

| Position | Effect |
| --- | --- |
| `beforebegin` | Insert immediately before the target among its parent's children. Requires a parent. |
| `afterbegin` | Prepend as children of the target. |
| `beforeend` | Append as children of the target. |
| `afterend` | Insert immediately after the target among its parent's children. Requires a parent. |

Unknown positions, missing targets, nil parses, and empty fragments are no-ops. A target without a parent cannot use `beforebegin` or `afterend`.

## Internal operations

### `insertBefore(parentID, newNodeID, refNodeID NodeID)`

Inserts a registered node before a reference child, or appends when the reference is absent, negative, or not a child of the specified parent. It does nothing if the parent or new node is missing/nil, the new node is the reference itself, or moving it would create a cycle. An existing parent relationship is removed before insertion.

A `DocumentFragmentNode` is special: its children are inserted in order before the reference (or appended), and the fragment is emptied. The operation checks every fragment child for cycles before mutating.

### `replaceChild(parentID, newNodeID, oldNodeID NodeID)`

Replaces `oldNode` in the specified parent's child list. Missing nodes, invalid identities, a new node equal to the parent or old node, a missing old child, and cycle-producing moves are ignored. A normal replacement first detaches the new node from its current parent, then clears the old node's parent pointer and installs the new node.

When the replacement is a document fragment, its children replace the old child in order and the fragment is emptied. The old node is detached. If the old node is not a child of the requested parent, no replacement occurs.

### `insertAdjacentText(nodeID NodeID, position, text string)`

Creates a `TextNode` and inserts it at one of the four adjacent positions listed above. It is an internal helper. Unlike `InsertAdjacentHTML`, it does not assign a registry ID or invalidate the CSS cache in this function. Invalid targets are ignored. Position handling is implemented by the switch; unsupported positions cause no insertion.

### `insertAdjacentElement(nodeID, elemID NodeID, position string)`

Moves an existing registered node to one of the four adjacent positions. It validates the target and element, requires a parent for sibling positions, rejects unsupported positions and cycle-producing moves, detaches the element from its current parent, then inserts it. Successful calls invalidate the CSS cache.

### `insertSiblingNodes(n *dom.Node, newNodes []*dom.Node, after bool)`

Internal helper that inserts nodes into `n`'s parent's child list immediately before or after `n`. It sets each inserted node's parent pointer. It does nothing if `n` has no parent or is not present in that parent's child list.

### `insertFragmentChildrenLocked(parent, fragment, refNode *dom.Node)`

Internal helper for fragment insertion. The caller must hold `nodeMu`. It copies the fragment child slice before splicing, sets each child's parent to `parent`, inserts before a matching reference or appends, and clears `fragment.Children`.

## Implementation notes

- Node lookup and tree mutation use the package-level `nodes` registry and `nodeMu` lock.
- Cycle prevention is performed with `wouldCreateCycle` for node moves and fragment children where applicable.
- Fragment insertion transfers children rather than attaching the fragment node itself.
- Cache invalidation occurs after successful `insertBefore`, `replaceChild`, `InsertAdjacentHTML`, and `insertAdjacentElement` mutations. `insertAdjacentText` does not invalidate it here.
- Invalid operations generally return silently; these helpers do not expose errors or results.
