# JavaScript style and logging bridge

This file implements JavaScript-callable handlers for computed style lookup, element geometry lookup, and logging. The handlers live in the `js` package and exchange JavaScriptCore values through the package's `Context` wrapper and node-ID conversion helpers.

## Handlers

These functions are package-internal callback entry points, not exported Go API. They are intended to be registered or invoked by the JavaScript bridge.

### `getComputedStyleJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

- Requires at least one argument. With no arguments, returns JavaScript `null`.
- Converts the first argument to a node ID with `jsToNodeID`.
- If conversion yields a negative ID, returns JavaScript `null`.
- Otherwise returns the result of `getComputedStyle(id)` directly.
- Additional arguments are ignored.

### `getBoundingClientRectJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

- Requires at least one argument and a valid nonnegative node ID; otherwise returns JavaScript `null`.
- Obtains the rectangle using `getBoundingClientRect(id)` and creates a plain JavaScript object with numeric properties:

| Property | Value |
| --- | --- |
| `x` | `rect.X` |
| `y` | `rect.Y` |
| `width` | `rect.Width` |
| `height` | `rect.Height` |
| `top` | `rect.Y` |
| `left` | `rect.X` |
| `bottom` | `rect.Y + rect.Height` |
| `right` | `rect.X + rect.Width` |

The object is converted to a `JSValueRef` using `jsObjToVal`. Properties are created with no special JavaScript property attributes. Additional input arguments are ignored.

### `logJS(ctx *Context, args []C.JSValueRef) C.JSValueRef`

Converts each argument with `getArgString` and writes one line to standard error. Failed conversions are represented as `<error>`. The output consists of the `[JS] ` prefix, arguments joined by single spaces, and a trailing newline.

Before writing, it replaces control characters with the Unicode replacement character (`�`), except for newline and tab; the message's final newline is therefore preserved. The DEL character (`0x7f`) is also replaced. The resulting line is passed to the C bridge helper `_fprintf_stderr`. The handler returns JavaScript `undefined`.

## Implementation notes

- This file uses cgo and JavaScriptCore C APIs to create the rectangle object and its numeric properties.
- Temporary C strings and `JSStringRef` property names are released after each property is set; the assembled log message's C string is likewise freed after writing.
- The helper declarations in the cgo preamble are supplied by the bridge implementation. Only a subset is used directly in this file.
- Node lookup, style computation, rectangle calculation, argument string conversion, and JavaScript value conversion are delegated to helpers elsewhere in the package.
