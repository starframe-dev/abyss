# DOM Events and Timers

This package file implements JavaScriptCore-backed event listener and timer registries for the DOM/runtime bridge. It keeps JavaScript callback objects protected from garbage collection while they are registered, invokes callbacks through JavaScriptCore, and releases those protections when registrations are removed or the event system is reset.

## Public API

| Function | Behavior |
| --- | --- |
| `InitEventSystem(ctx C.JSGlobalContextRef)` | Resets prior event-system state, installs the supplied JavaScriptCore global context, initializes an empty listener registry, and calls `MarkClean()`. |
| `AddEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef)` | Registers a callback for the given event type on a node. Repeated registrations are retained independently. |
| `RemoveEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef)` | Removes the first matching callback reference for that node and event type. No action is taken if the node, event type, or callback is absent. |
| `ProcessTimers() bool` | Executes all currently due timers. Returns `false` when none are ready and `true` if at least one callback was invoked. The caller must ensure the relevant JavaScriptCore context is active. |

`NodeID` and `C.JSGlobalContextRef` / `C.JSObjectRef` are supplied by the surrounding JavaScript bridge; this file does not define additional public data types.

## Event dispatch

`dispatchEvent(nodeID, eventType)` is package-private. It snapshots the registered callbacks while holding the event mutex, then releases the lock before invoking JavaScript. This avoids holding the registry lock across foreign-runtime calls. Callbacks are invoked in registration order with one argument: a newly created event object. The call uses a null `this` value.

The event object contains:

- `type`: the event type string;
- `defaultPrevented`, initially `false`;
- `preventDefault()`: sets `defaultPrevented` to `true`;
- `stopPropagation()`: sets the internal `_propagationStopped` flag to `true`.

After each callback, dispatch checks `_propagationStopped` and stops invoking further callbacks when it is truthy. `defaultPrevented` is exposed for consumers but does not otherwise affect dispatch in this implementation. JavaScript exceptions are passed to JavaScriptCore's exception output parameter but are not otherwise handled here.

The event object is constructed by evaluating a JavaScript expression in the active context. If evaluation returns a null result, `makeEventObject` returns JavaScript `null`.

## Timers

The package-private `setTimeout(callback, delayMs)` and `setInterval(callback, delayMs)` functions allocate monotonically increasing integer IDs and register protected callbacks with deadlines based on `time.Now()`. Both interpret the delay as milliseconds. A timeout is one-shot; an interval repeats using the same interval duration. `clearTimer(id)` removes a matching timer and releases its callback protection; unknown IDs are ignored.

`ProcessTimers` partitions the timer list into due and pending timers under the timer mutex, then invokes due callbacks outside that lock with no arguments and a null `this`. One-shot callbacks are unprotected after invocation. Repeating timers are scheduled again with a deadline of the processing pass's captured current time plus the interval. Thus, this is a polling mechanism: timers run only when the host calls `ProcessTimers`, and repeating timers do not catch up missed intervals. Callback exceptions are not surfaced by this function.

## Context lifecycle and callback ownership

`resetEventSystem` and `resetEventSystemForContext` are package-private lifecycle helpers. Reset releases protections for listener and timer callbacks using the old JavaScriptCore context, clears listener state and timer state, and resets timer IDs. The context-specific variant does nothing when a non-nil expected context differs from the currently installed context. `InitEventSystem` invokes the unconditional reset before installing its new context.

Listeners are stored by node ID and event type; timers are stored in a shared timer slice. Registry operations use mutexes, while JavaScript callback execution occurs after unlocking. The event context must remain valid whenever callbacks are protected, unprotected, or invoked. `ProcessTimers` specifically requires the current context to be active.
