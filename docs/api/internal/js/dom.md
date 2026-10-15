# JavaScript DOM Bridge State

`internal/js/dom.go` implements the Go-side DOM state used by the JavaScriptCore bridge. It associates JavaScript-facing node identifiers with `internal/dom` nodes, maintains per-page bridge state, and provides mutation and lookup helpers used by other package files. Most DOM helpers are unexported; this page documents them as implementation API rather than callable external-package API.

## Public API

| Name | Description |
| --- | --- |
| `NodeID` | Integer identifier assigned to a DOM node for bridge operations. |
| `Rect` | Layout rectangle with `X`, `Y`, `Width`, and `Height` as `float64`. |
| `Runtime` | Snapshot of one page's JavaScript context and bridge state while inactive. |
| `SetCurrentContext` | Installs an active context and URL, closing the prior distinct context. |
| `GetCurrentContext` | Returns the active context and its URL. |
| `CaptureRuntime` | Detaches the active page context and its bridge state into a `Runtime`. |
| `ActivateRuntime` | Restores a captured runtime as the active page. A nil runtime is ignored. |
| `CloseRuntime` | Activates a detached runtime and closes it, releasing its callbacks and DOM state. |
| `CloseCurrentContext` | Closes the active JavaScriptCore context and resets DOM state. |
| `SetDirty` | Marks the DOM as changed. |
| `IsDirty` | Reports whether the DOM has been marked changed. |
| `MarkClean` | Clears the dirty flag, typically after rendering. |
| `InitDOM` | Resets bridge DOM state and registers nodes in a parsed document. |

`Runtime` fields are unexported; callers should use the capture/activate/close functions rather than manipulate snapshot contents.

## State and lifecycle

The bridge keeps package-level state for the active context and URL, node ID mappings, root document, computed styles, layout rectangles, event callbacks, timers, and a dirty flag. Mutexes protect these individual state groups. The browser is responsible for serializing activation and JavaScript execution at the lifecycle level; the mutexes do not make arbitrary concurrent DOM operations safe as a substitute for that requirement.

`CaptureRuntime` transfers (rather than copies) the active context and state maps/slices into a new runtime, then clears the package-level state. `ActivateRuntime` performs the reverse transfer and empties the snapshot. A runtime therefore represents detached state and should not be activated concurrently or repeatedly as if it were an immutable copy. Callers must ensure that no other page runtime is active when activating it.

Replacing a non-nil active context with a different context in `SetCurrentContext` resets the event system and closes the old context. `CloseCurrentContext` also resets event handling, closes the current context if present, and resets node, style, layout, and dirty state. Runtime context lifecycle and callback cleanup are coordinated with other bridge files.

## Node identity and initialization

`InitDOM` clears existing DOM/layout/style state, sets the supplied node as the root document, and recursively registers it and its descendants. IDs start at `1`; the maps provide lookup in both directions. Registration stops when the number of registered nodes reaches `dom.MaxDOMNodes`; the sentinel returned for failed registration is `-1`. New elements, text nodes, fragments, and parsed inner-HTML descendants are assigned IDs through the same registry.

`Rect` is the value shape used by the computed layout bridge. Computed styles and layout boxes are stored separately from the DOM node registry and are transferred with a runtime snapshot.

## DOM operations

The unexported helpers are the bridge's internal mutation and lookup surface:

- `createElement` lowercases the tag name and creates an element with initialized attribute and child collections.
- `createTextNode` creates a text node containing the supplied text.
- `createDocumentFragment` creates a fragment named `#document-fragment`.
- `appendChild` validates both IDs, rejects self-parenting and cycles, detaches ordinary children from their previous parent, then appends them. Appending a fragment moves its children into the destination (the fragment itself is not inserted); if any fragment child would create a cycle, the operation is rejected.
- `setInnerHTML` parses markup using `dom.ParseFragment`, clears existing children and detaches them, then inserts parsed direct children and registers their descendants. It does not replace the target node itself.
- `setAttribute` lowercases the attribute name and stores the value, initializing the attribute map when needed.
- `setTextContent` clears and detaches existing children and adds a new text node. It does nothing for an unknown ID or when the node limit has already been reached.
- `removeChild` removes the specified node only if it is currently a direct child of the specified parent, and clears its parent pointer. It does not unregister the node from the ID maps.
- `getBody` finds a direct `html` element under the root document and then its direct `body` child; it returns nil if that structure is absent.
- `getElementByID` recursively searches the root tree for the first element whose `id` attribute equals the supplied value.
- `getNodeID` returns a registered node's ID or `-1`; `getAttribute` returns an attribute value or an empty string for an unknown node.

Invalid IDs and rejected structural mutations generally result in a no-op or sentinel/empty return rather than an error. Callers should not interpret an empty attribute as proof that the attribute exists with an empty value.

## Implementation details

Node-tree operations use `nodeMu`; helpers whose names end in `Locked` require that mutex to be held. Cycle detection walks upward from the proposed parent through `Parent` links. Moving children preserves each child's identity and updates its parent pointer. Removed and replaced nodes remain registered in the ID map, so IDs are not recycled during a DOM lifetime. The global registration limit applies to initial traversal and subsequent node creation.

The module uses cgo's JavaScriptCore object/context reference types for event callbacks and timers. Those references are retained as part of runtime state; their lifecycle is handled alongside the JavaScript context. The `timer` record stores an identifier, callback reference, deadline, interval, and repeat flag for the timer bridge.
