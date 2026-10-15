# `internal/js`

Package `js` provides JavaScript execution and value/property conversion through Apple's JavaScriptCore (JSC) framework. It is a Go-to-C bridge and requires the JavaScriptCore framework on macOS; the source enables cgo and compiles Objective-C bridge code.

## Context lifecycle

`Context` owns a JSC global execution context. Create one with `NewContext`, use its methods for evaluation and value operations, then call `Close` to release the native context. `Close` also resets the package event system for that context. It is safe to call `Close` repeatedly: after releasing, the stored native reference is nil. Other methods assume a live context; they do not check for a closed context.

```go
ctx := js.NewContext()
defer ctx.Close()

result, err := ctx.Evaluate("1 + 2")
```

## Public API

| API | Behavior |
| --- | --- |
| `NewContext() *Context` | Creates a fresh JavaScriptCore global context. |
| `(*Context).Close()` | Resets event-system state for the context and releases it. |
| `(*Context).GetContextRef() C.JSGlobalContextRef` | Exposes the underlying cgo/JSC context reference. |
| `(*Context).Evaluate(script string) (string, error)` | Evaluates source and converts the result to a JavaScript string representation. Uses `script.js` as the source URL. |
| `(*Context).EvaluateRaw(script string) (C.JSValueRef, error)` | Evaluates source and returns the raw JSC value reference. Uses `script.js` as the source URL. |
| `(*Context).GlobalObject() C.JSObjectRef` | Returns the context's global object. |
| `(*Context).SetProperty(name string, value C.JSValueRef) error` | Sets a property on the global object. |
| `(*Context).SetPropertyOn(obj C.JSObjectRef, name string, value C.JSValueRef) error` | Sets a property on the supplied object. |
| `(*Context).GetProperty(name string) (C.JSValueRef, error)` | Gets a property from the global object. |
| `(*Context).ValueToString(value C.JSValueRef) (string, error)` | Converts a JSC value to a Go string. |
| `(*Context).MakeString(value string) C.JSValueRef` | Creates a JavaScript string. |
| `(*Context).MakeNumber(value float64) C.JSValueRef` | Creates a JavaScript number. |
| `(*Context).MakeBoolean(value bool) C.JSValueRef` | Creates a JavaScript boolean. |
| `(*Context).MakeUndefined() C.JSValueRef` | Creates JavaScript `undefined`. |
| `(*Context).MakeNull() C.JSValueRef` | Creates JavaScript `null`. |
| `(*Context).ToNumber(value C.JSValueRef) (float64, error)` | Applies JavaScriptCore numeric conversion and returns a Go `float64`. |
| `(*Context).ToBoolean(value C.JSValueRef) bool` | Applies JavaScript truthiness conversion. |
| `(*Context).GetType(value C.JSValueRef) string` | Returns a normalized JavaScript type name. |

The `C.JSValueRef`, `C.JSObjectRef`, and `C.JSGlobalContextRef` types are cgo representations of JavaScriptCore references. They are not Go-owned values; callers should use them with the context that created them and avoid retaining them beyond its lifetime.

## Evaluation and errors

`Evaluate` and `EvaluateRaw` invoke the C bridge evaluator and inspect the JavaScript exception output. Exceptions become Go errors prefixed with `js error:`; if the exception cannot be stringified, the methods return `unknown js error`. `Evaluate` stringifies a successful result, while `EvaluateRaw` preserves the native value reference. A nil result from `Evaluate` is represented as an empty string with no error.

Property operations and numeric conversion similarly report JSC exceptions as Go errors with operation-specific prefixes (`set property error:`, `get property error:`, or `to number error:`), falling back to an `unknown ... error` message if the exception cannot be converted. `ValueToString` reports conversion exceptions with `value to string error:`; if there is neither a string result nor a reported exception, it returns an empty string without error.

## Type and conversion behavior

`GetType` maps JSC types to `undefined`, `null`, `boolean`, `number`, `string`, or `object`. Any JSC type not covered by those cases maps to `unknown`.

`ToBoolean` returns JSC's boolean coercion directly and does not expose an error result. `ToNumber` can fail if JavaScriptCore raises an exception during conversion. Constructors and property methods allocate temporary C strings for names and string values and free those buffers after the bridge call.

## Implementation notes

The Go wrapper delegates native operations to the Objective-C/C bridge declared in the cgo preamble, including evaluation, property access, primitive construction, conversion, and type inspection. `Evaluate` and `EvaluateRaw` use the fixed source URL `script.js`; the API does not currently allow callers to specify a URL. JSC calls and native references are bound to the context, so use a context consistently and close it when finished.
