# DOM Script API

This package file bridges DOM nodes to JavaScriptCore values and executes scripts found in a DOM document. It also exposes access to the final DOM tree.

## Public API

### `GetFinalDOM`

```go
func GetFinalDOM() *dom.Node
```

Returns `rootDoc`, the package-level root document after script execution. The function returns the stored pointer directly; it does not clone the tree or check for a nil value.

### `ExecuteScripts`

```go
func ExecuteScripts(ctx *Context, doc *dom.Node, baseURL string) error
```

Executes scripts in `doc`, using `baseURL` to resolve external script URLs. This convenience function uses `context.Background`, so callers cannot cancel the operation directly. Use `ExecuteScriptsContext` when cancellation or a caller deadline is needed.

### `ExecuteScriptsContext`

```go
func ExecuteScriptsContext(execCtx context.Context, ctx *Context, doc *dom.Node, baseURL string) error
```

Processes script elements in document order. It applies an overall 30-second timeout derived from `execCtx`, and enforces per-page limits:

- At most 128 script elements are considered; additional elements are ignored.
- At most 8 MiB of script source is evaluated in total. Source length is measured in bytes. If the next script exceeds the remaining budget, processing stops rather than skipping that script and continuing.

For an element with a non-empty `src` attribute, the URL is resolved against `baseURL`, fetched, then evaluated. Otherwise, its text content is treated as inline JavaScript; whitespace-only inline content is skipped. Scripts are evaluated sequentially.

Fetch failures and JavaScript evaluation errors are deliberately ignored so that an individual problematic script does not fail processing of the page. Cancellation or expiration of the overall context is returned as an error when observed at the start of an iteration. The function otherwise returns nil, including when individual scripts fail. A nil document or JavaScript context is not checked and is not a supported input.

## DOM-to-JavaScript representation

The unexported `nodeToJS` helper creates a plain JavaScriptCore object for a DOM node. It stores the internal node ID in `__go_id` and exposes `nodeType` using standard DOM numeric values for supported node kinds:

| Internal node kind | JavaScript `nodeType` |
| --- | ---: |
| Element | 1 |
| Text | 3 |
| Comment | 8 |
| Document | 9 |
| Document fragment | 11 |

Element wrappers additionally have `tagName`, uppercased from the internal tag name. Text-node wrappers have `textContent` containing the node's text. Other DOM properties and methods are not added here. If `getNodeID` returns a negative ID, the helper returns JavaScript `null` instead of a wrapper object.

## External script fetching

The unexported `fetchScript` helper provides a three-second timeout using a background context. `fetchScriptContext` parses the base and source as URLs, resolves the source reference against the base, then calls `fetcher.FetchJS` with a three-second timeout derived from the supplied parent context. URL parse errors are wrapped with context identifying whether the base URL or script source failed to parse. In normal script execution, fetch errors are swallowed by `ExecuteScriptsContext`.

## Implementation notes

JavaScript values and properties are created through JavaScriptCore's C API. Temporary C strings and JavaScriptCore string references used for property names are released with deferred cleanup. Script execution relies on the existing `Context.Evaluate` implementation; this file does not provide transactional rollback, per-script error reporting, or DOM wrapper synchronization beyond the node properties described above.
