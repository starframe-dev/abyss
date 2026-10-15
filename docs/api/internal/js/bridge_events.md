# JavaScript Event and Timer Bridge

`internal/js/bridge_events.go` implements the JavaScript-facing adapters for timer and node-event operations. Each adapter receives a JavaScriptCore context wrapper (`*Context`) and a slice of JavaScriptCore values, converts the arguments needed by the internal event system, and returns a JavaScriptCore value.

These functions are package-private implementation details. The bridge registers their names in the JavaScript global environment; they are not exported Go API methods.

## JavaScript bridge functions

| JavaScript name | Go adapter | Arguments | Result |
| --- | --- | --- | --- |
| `setTimeout` | `setTimeoutJS` | callback, delay in milliseconds | Numeric timer ID; `0` if arguments are missing or callback conversion fails |
| `setInterval` | `setIntervalJS` | callback, delay in milliseconds | Numeric timer ID; `0` if arguments are missing or callback conversion fails |
| `clearTimeout` | `clearTimeoutJS` | timer ID | `undefined` |
| `clearInterval` | `clearIntervalJS` | timer ID | `undefined` |
| `addEventListener` | `addEventListenerJS` | node ID, event type, callback | `undefined` |
| `removeEventListener` | `removeEventListenerJS` | node ID, event type, callback | `undefined` |
| `dispatchEvent` | `dispatchEventJS` | node ID, event type | `undefined` |

The adapters are registered under these names by the bridge setup in `bridge.go`. The names provide JavaScript-style access to the Go event system; this file does not define JavaScript language-level event targets or browser compatibility beyond the behavior described here.

## Argument handling

- Timer adapters require at least two arguments. The first is converted to a JavaScriptCore object and must not be null. The delay is converted with `Context.ToNumber`; if conversion fails, the adapter uses `0`. The resulting number is converted to a Go `int` without additional range or sign checks.
- Clear adapters require at least one argument. If its numeric conversion succeeds, it is converted to `int` and passed to the shared timer clearer. Missing or non-numeric IDs are ignored.
- Event adapters require the listed argument count. The node ID is converted using `jsToNodeID`; a negative result causes the operation to be ignored. Event type extraction uses `getArgString`; its error is ignored, so the helper's returned string is used even when conversion reports an error.
- Listener callbacks must convert to a non-nil JavaScriptCore object. Otherwise registration or removal is ignored.
- Invalid or incomplete calls return the same default values as their corresponding valid calls: timer creation returns numeric zero on early failure, while clear and event operations return `undefined`.

## Event behavior

The adapters delegate to the event subsystem in `dom_events.go`. Listeners are associated with a node ID and event-type string. Dispatch invokes the registered callbacks with one event object argument. That event object exposes `type`, `defaultPrevented`, `preventDefault()`, and `stopPropagation()`; calling `stopPropagation()` stops further callbacks for that dispatch. Listener registration and removal operate on callback identity, and removal removes one matching registration. The bridge adapters themselves do not report whether a listener was found or whether dispatch invoked any callbacks.

## Timer behavior

Timer creation delegates to the shared timer queue and returns its integer ID as a JavaScript number. `setTimeout` schedules a single callback; `setInterval` schedules a repeating callback. Both clear adapters call the same internal `clearTimer` operation, so either name can clear a timer created by either creation function. Actual execution is handled by `ProcessTimers` in the event subsystem, not by these bridge adapters; the host must process timers with the JavaScript context active.

## Types and implementation notes

- `Context` is the package's wrapper around a JavaScriptCore context and supplies value conversion and JavaScript value construction methods such as `ToNumber`, `MakeNumber`, and `MakeUndefined`.
- `C.JSValueRef` represents incoming JavaScript values and adapter return values; `C.JSObjectRef` is used for callbacks.
- The cgo preamble includes JavaScriptCore declarations and bridge helper declarations. Callback conversion uses `JSValueToObject` directly, while numeric conversion is delegated to `Context`.
- These adapters perform no explicit JavaScript exception handling or validation beyond the checks above. Internal timer/event state, callback lifetime protection, and execution details belong to `dom_events.go`.
