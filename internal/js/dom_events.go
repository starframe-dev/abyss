package js

/*
#include <JavaScriptCore/JavaScriptCore.h>

extern void _protectValue(JSContextRef ctx, JSValueRef value);
extern void _unprotectValue(JSContextRef ctx, JSValueRef value);
extern JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception);
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
extern JSValueRef _makeBoolean(JSContextRef ctx, bool value);
extern JSValueRef _evaluate(JSContextRef ctx, const char* script, const char* sourceURL, JSValueRef* exception);
*/
import "C"
import (
    "time"
    "unsafe"
)

// InitEventSystem initializes the event system with a JSC context.
func InitEventSystem(ctx C.JSGlobalContextRef) {
    eventMu.Lock()
    defer eventMu.Unlock()
    eventCtx = ctx
    eventListeners = map[NodeID]map[string][]C.JSObjectRef{}

    timerMu.Lock()
    timers = nil
    nextTimerID = 0
    timerMu.Unlock()

    MarkClean()
}

// AddEventListener registers a callback for an event on a node.
func AddEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef) {
    addEventListener(nodeID, eventType, callback)
}

// RemoveEventListener removes a callback for an event on a node.
func RemoveEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef) {
    removeEventListener(nodeID, eventType, callback)
}

// addEventListener registers a callback for an event on a node.
func addEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef) {
    eventMu.Lock()
    defer eventMu.Unlock()

    if eventListeners[nodeID] == nil {
        eventListeners[nodeID] = map[string][]C.JSObjectRef{}
    }
    eventListeners[nodeID][eventType] = append(eventListeners[nodeID][eventType], callback)
    C._protectValue(eventCtx, jsObjToVal(callback))
}

// removeEventListener removes a callback for an event on a node.
func removeEventListener(nodeID NodeID, eventType string, callback C.JSObjectRef) {
    eventMu.Lock()
    defer eventMu.Unlock()

    byType, ok := eventListeners[nodeID]
    if !ok {
        return
    }
    callbacks, ok := byType[eventType]
    if !ok {
        return
    }
    for i, cb := range callbacks {
        if cb == callback {
            byType[eventType] = append(callbacks[:i], callbacks[i+1:]...)
            C._unprotectValue(eventCtx, jsObjToVal(cb))
            break
        }
    }
}

// dispatchEvent fires an event on a node, calling all registered listeners.
func dispatchEvent(nodeID NodeID, eventType string) {
    eventMu.Lock()
    byType, ok := eventListeners[nodeID]
    if !ok {
        eventMu.Unlock()
        return
    }
    callbacks, ok := byType[eventType]
    if !ok || len(callbacks) == 0 {
        eventMu.Unlock()
        return
    }
    // Copy callbacks to avoid holding lock during JS calls
    cbs := make([]C.JSObjectRef, len(callbacks))
    copy(cbs, callbacks)
    eventMu.Unlock()

    // Create event object
    eventObj := makeEventObject(eventType)

    for _, cb := range cbs {
        var exc C.JSValueRef
        C._callAsFunction(eventCtx, cb, nil, 1, &eventObj, &exc)

        // Check if propagation was stopped
        propName := C.JSStringCreateWithUTF8CString(C.CString("_propagationStopped"))
        stoppedVal := C.JSObjectGetProperty(eventCtx, C.JSValueToObject(eventCtx, eventObj, nil), propName, nil)
        C.JSStringRelease(propName)
        if C.JSValueToBoolean(eventCtx, stoppedVal) {
            break
        }
    }
}

// makeEventObject creates a JS event object with the given type.
// The event object has: type, defaultPrevented, preventDefault(), stopPropagation().
func makeEventObject(eventType string) C.JSValueRef {
    ctx := eventCtx

    // Create event object via JS evaluation for proper function support
    js := `(function(type){
        var e = {type:type, defaultPrevented:false, _propagationStopped:false};
        e.preventDefault = function(){ e.defaultPrevented = true; };
        e.stopPropagation = function(){ e._propagationStopped = true; };
        return e;
    })("` + eventType + `")`

    cjs := C.CString(js)
    defer C.free(unsafe.Pointer(cjs))

    var exc C.JSValueRef
    result := C._evaluate(ctx, cjs, nil, &exc)
    if result == nil {
        return C.JSValueMakeNull(ctx)
    }
    return result
}

// setTimeout registers a callback to be called after delayMs milliseconds.
func setTimeout(callback C.JSObjectRef, delayMs int) int {
    timerMu.Lock()
    defer timerMu.Unlock()

    nextTimerID++
    id := nextTimerID
    C._protectValue(eventCtx, jsObjToVal(callback))
    timers = append(timers, timer{
        ID:       id,
        Callback: callback,
        Deadline: time.Now().Add(time.Duration(delayMs) * time.Millisecond),
        Interval: time.Duration(delayMs) * time.Millisecond,
        Repeat:   false,
    })
    return id
}

// setInterval registers a callback to be called every delayMs milliseconds.
func setInterval(callback C.JSObjectRef, delayMs int) int {
    timerMu.Lock()
    defer timerMu.Unlock()

    nextTimerID++
    id := nextTimerID
    C._protectValue(eventCtx, jsObjToVal(callback))
    timers = append(timers, timer{
        ID:       id,
        Callback: callback,
        Deadline: time.Now().Add(time.Duration(delayMs) * time.Millisecond),
        Interval: time.Duration(delayMs) * time.Millisecond,
        Repeat:   true,
    })
    return id
}

// clearTimer removes a timer by ID.
func clearTimer(id int) {
    timerMu.Lock()
    defer timerMu.Unlock()

    for i, t := range timers {
        if t.ID == id {
            C._unprotectValue(eventCtx, jsObjToVal(t.Callback))
            timers = append(timers[:i], timers[i+1:]...)
            return
        }
    }
}

// ProcessTimers executes all ready timers. Must be called with current context active.
func ProcessTimers() bool {
    timerMu.Lock()
    now := time.Now()
    var ready []timer
    var remaining []timer
    for _, t := range timers {
        if now.After(t.Deadline) || now.Equal(t.Deadline) {
            ready = append(ready, t)
        } else {
            remaining = append(remaining, t)
        }
    }
    timers = remaining
    timerMu.Unlock()

    if len(ready) == 0 {
        return false
    }

    executed := false
    for _, t := range ready {
        var exc C.JSValueRef
        C._callAsFunction(eventCtx, t.Callback, nil, 0, nil, &exc)
        executed = true
        if t.Repeat {
            t.Deadline = now.Add(t.Interval)
            timerMu.Lock()
            timers = append(timers, t)
            timerMu.Unlock()
        } else {
            C._unprotectValue(eventCtx, jsObjToVal(t.Callback))
        }
    }
    return executed
}
