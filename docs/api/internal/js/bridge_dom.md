# DOM JavaScript Bridge

`internal/js/bridge_dom.go` implements JavaScript-callable adapters for the project's DOM operations. Each handler receives the JavaScriptCore context and a slice of raw JavaScript arguments, converts values to internal node IDs or strings, invokes the DOM implementation, then returns an appropriate JavaScript value.

These functions are package-private Go functions, not exported Go API. They are intended to be registered and called through the JavaScript bridge elsewhere in the `internal/js` package.

## API summary

| Handler | Argument order | Result |
| --- | --- | --- |
| `createElementJS` | `tagName` | Wrapped newly created element; missing argument yields `undefined`. An empty tag name defaults to `div`. |
| `appendChildJS` | `parent, child` | The original child argument on success; otherwise `undefined`. |
| `setInnerHTMLJS` | `node, html` | `undefined`. |
| `setAttributeJS` | `node, name, value` | `undefined`. |
| `setTextContentJS` | `node, text` | `undefined`. |
| `createTextNodeJS` | `text` | Wrapped newly created text node; missing argument yields `undefined`. |
| `createDocumentFragmentJS` | none | Wrapped newly created document fragment. |
| `getInnerHTMLJS` | `node` | HTML string; missing or invalid node yields an empty string. |
| `getTextContentJS` | `node` | Text string; missing or invalid node yields an empty string. |
| `removeChildJS` | `parent, child` | The original child argument on success; otherwise `undefined`. |
| `insertBeforeJS` | `parent, newNode, referenceNode` | The original new-node argument on success; otherwise `undefined`. |
| `replaceChildJS` | `parent, newNode, oldNode` | The original old-node argument on success; otherwise `undefined`. |
| `cloneNodeJS` | `node, deep?` | Wrapped clone; missing/invalid node or failed clone yields `null`. `deep` is converted using JavaScript truthiness and defaults to `false`. |
| `containsJS` | `parent, child` | Boolean; invalid or missing arguments yield `false`. |
| `insertAdjacentHTMLJS` | `node, position, html` | `undefined`. |
| `insertAdjacentTextJS` | `node, position, text` | `undefined`. |
| `insertAdjacentElementJS` | `node, element, position` | Original element argument; invalid/missing node arguments yield `null`. |
| `outerHTMLJS` | `node` | Serialized outer HTML string; missing or invalid node yields an empty string. |
| `setOuterHTMLJS` | `node, html` | `undefined`. |

Node arguments are converted with `jsToNodeID`; a negative ID is treated as invalid. String arguments are converted with `getArgString`. The handlers generally ignore conversion errors and use the returned string value, so callers should provide values in the expected shape.

## DOM mutations and rendering state

Successful mutation handlers call `SetDirty()` so the host can refresh its rendered state. `appendChildJS`, `removeChildJS`, `setInnerHTMLJS`, and `setAttributeJS` also call `css.InvalidateCache()` because those operations can affect style resolution. The other mutation adapters mark the DOM dirty but do not explicitly invalidate the CSS cache in this file.

The adapters do not expose mutation errors as JavaScript exceptions. Most malformed calls return `undefined`; queries use empty strings or `false`, while cloning and adjacent-element insertion use `null` for their specified failure cases. Underlying DOM functions determine the details of tree mutation and serialization.

## Implementation details

- Returned node wrappers are created by `nodeToJS`; created nodes come from `createElement`, `createTextNode`, and `createDocumentFragment`.
- Mutating methods return the original JavaScript argument where DOM-style operations conventionally return a node. This preserves the passed JS value rather than constructing another wrapper.
- `insertBeforeJS` passes the reference node ID through without rejecting a negative value explicitly; only the parent and new-node IDs are checked before calling `insertBefore`.
- `outerHTMLJS` converts the Go string to a C string, calls the JavaScriptCore bridge helper `_makeString`, and frees the temporary C allocation with `C.free`.
- The file imports JavaScriptCore declarations through cgo and uses the CSS package solely for explicit cache invalidation.
