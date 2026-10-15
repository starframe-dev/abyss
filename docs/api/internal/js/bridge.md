# JavaScript Bridge

`internal/js/bridge.go` connects JavaScriptCore values and callable JavaScript functions to Go implementations of the DOM API. It registers bridge functions on a JavaScript context's global object, converts incoming arguments, and dispatches calls to handlers.

## Public API

### `LogWriter`

```go
var LogWriter io.Writer = io.Discard
```

Writer used by the `console.log` bridge handler. Its default is `io.Discard`, so output is discarded unless the application assigns another writer. The actual formatting and write behavior are implemented by `logJS`.

### `(*Context).RegisterBridgeFunctions`

```go
func (c *Context) RegisterBridgeFunctions() error
```

Registers the bridge entry points on the context's global object under names prefixed with `__go_` (for example, `__go_createElement`). Returns the first error from setting a global property; registration stops at that point. The registered names are:

- DOM creation and mutation: `createElement`, `createTextNode`, `createDocumentFragment`, `appendChild`, `insertBefore`, `replaceChild`, `removeChild`, `cloneNode`, `setInnerHTML`, `setOuterHTML`, `setTextContent`, `setAttribute`, `insertAdjacentHTML`, `insertAdjacentText`, `insertAdjacentElement`
- DOM access and queries: `getBody`, `getElementById`, `getAttribute`, `getParentNode`, `getChildren`, `getInnerHTML`, `getTextContent`, `getElementsByTagName`, `getElementsByClassName`, `querySelector`, `querySelectorAll`, `matches`, `closest`, `contains`, `outerHTML`
- Events, timers, and presentation: `addEventListener`, `removeEventListener`, `dispatchEvent`, `setTimeout`, `setInterval`, `clearTimeout`, `clearInterval`, `getComputedStyle`, `getBoundingClientRect`
- Element relationships: `previousElementSibling`, `nextElementSibling`, `firstElementChild`, `lastElementChild`
- Logging: `log`

The corresponding Go handlers are held in the package's `bridgeHandlers` map. Registration passes each name to the C helper `makeGoFunction` and installs the resulting function with `Context.SetProperty`.

## Dispatch and conversions

The exported cgo callback `goBridgeDispatch` is the common entry point invoked by the C bridge. It converts the C function name to a Go string, wraps the JavaScriptCore context in the package's `Context`, copies the provided argument array into Go-managed memory, and invokes the matching handler. If no handler exists, it creates a JavaScript string exception of the form `unknown function: <name>` and returns a null value.

Internal helpers support this boundary:

- `getArgString` converts a JavaScript value to a Go string through `_valueToString`, frees the returned C allocation, and returns an error if conversion yields a null pointer.
- `jsToNodeID` accepts only a JavaScript object with a numeric `__go_id` property convertible by `Context.ToNumber`. Any failed type check, object/property lookup, or number conversion produces the sentinel `NodeID(-1)`.
- `nodeToJSValue` wraps a `*dom.Node` using `nodeToJS`.
- `jsObjToVal` reinterprets a `JSObjectRef` as a `JSValueRef` for property installation.

## Implementation notes

This file uses cgo and JavaScriptCore types from `<JavaScriptCore/JavaScriptCore.h>`, with supporting functions declared in the C preamble and implemented by the bridge's C code. The callback copies arguments before passing them to Go handlers, avoiding reliance on the C argument slice beyond the callback. Bridge function names are maintained in both the handler map and registration list; adding a handler requires making its name available through registration as well.

`bridgeHandlers`, `bridgeFunc`, and the conversion helpers are package-internal rather than public API. The bridge's DOM behavior is provided by the individual handler implementations and the `internal/dom` package; this file defines dispatch, registration, and value conversion rather than those operations' detailed semantics.
