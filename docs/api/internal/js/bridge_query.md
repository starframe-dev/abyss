# JavaScript Bridge Query Operations

`internal/js/bridge_query.go` implements JavaScript-callable adapters for querying the document/node model. Each function receives a JavaScriptCore `Context` and the arguments passed by JavaScript, delegates to the corresponding Go query operation, and converts results back into JavaScript values.

These adapter functions are package-private. They are intended to be registered or invoked by the surrounding bridge; this file does not itself define JavaScript global names or registration behavior.

## Argument and result conventions

- Node arguments are converted to internal node IDs with `jsToNodeID`. A negative result is treated as an invalid node reference.
- String arguments are read with `getArgString`. Most adapters discard conversion errors; `matchesJS` and `closestJS` explicitly treat conversion errors as failure.
- Nodes are exposed through `nodeToJS`.
- Missing arguments and invalid references return the sentinel shown in the table. The distinction between JavaScript `null` and `undefined` is intentional.
- Collection results are JavaScript arrays, including when there are no results.

## Bridge operations

| Adapter | Arguments | Result and behavior |
| --- | --- | --- |
| `getBodyJS` | None | Converts the document body to JavaScript; returns `null` if no body exists. Extra arguments are ignored. |
| `getElementByIdJS` | `id` string | Looks up an element by ID and returns its converted node, or `null` when the argument is missing or no node is found. |
| `querySelectorJS` | `nodeID`, selector string | Queries within the referenced node and returns the first matching node, or `null` for missing/invalid input or no match. |
| `querySelectorAllJS` | `nodeID`, selector string | Queries within the referenced node and returns an array of converted matches. Missing arguments or an invalid node ID produce `undefined`; no matches produce an empty array. |
| `getAttributeJS` | `nodeID`, attribute name | Returns the attribute value as a JavaScript string. Missing arguments or an invalid node ID produce `null`. The underlying lookup's returned string is converted directly. |
| `getParentNodeJS` | `nodeID` | Returns the parent node, or `null` for missing/invalid input or when there is no parent. |
| `getChildrenJS` | `nodeID` | Returns an array of converted children. Missing arguments or an invalid node ID produce `undefined`; no children produce an empty array. |
| `getElementsByTagNameJS` | Tag name | Returns an array of matching nodes. Missing argument produces `undefined`; no matches produce an empty array. |
| `getElementsByClassNameJS` | Class name | Returns an array of matching nodes. Missing argument produces `undefined`; no matches produce an empty array. |
| `matchesJS` | `nodeID`, selector string | Returns a JavaScript boolean. Missing arguments, invalid node IDs, or selector conversion errors return `false`; otherwise delegates to `matches`. |
| `closestJS` | `nodeID`, selector string | Returns the closest matching node, or `null`. Missing/invalid input, selector conversion errors, and no match return `null`. |
| `previousElementSiblingJS` | `nodeID` | Returns the previous element sibling, or `null` if unavailable or input is invalid. |
| `nextElementSiblingJS` | `nodeID` | Returns the next element sibling, or `null` if unavailable or input is invalid. |
| `firstElementChildJS` | `nodeID` | Returns the first element child, or `null` if unavailable or input is invalid. |
| `lastElementChildJS` | `nodeID` | Returns the last element child, or `null` if unavailable or input is invalid. |

## Implementation details

The adapters preserve the underlying query functions' matching and traversal semantics; selector parsing, tree traversal, and node lookup are implemented elsewhere. Result conversion is centralized through `nodeToJS`, while primitive values use the `Context` constructors.

For list operations, each Go node is converted individually to a `C.JSValueRef`, then `_makeArray` constructs the JavaScript array. Empty results use `_makeArray` with a nil values pointer and a zero count. The local exception reference is passed to the C helper, but this file does not inspect or translate it into a Go error.

`getBodyJS` and `getElementByIdJS` do not require a node ID argument. All other node-scoped operations reject an ID that converts to a negative value before calling the query layer. `getAttributeJS`, `querySelectorJS`, and the tag/class lookups ignore string-conversion errors, whereas selector conversion errors are explicitly handled by `matchesJS` and `closestJS`.
